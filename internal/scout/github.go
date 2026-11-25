package scout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// githubAPIBase is overridden in tests to point at an httptest.Server.
var githubAPIBase = "https://api.github.com"

type ghSearchResult struct {
	Items []Repo `json:"items"`
}

// Repo is a GitHub repository as returned by the search or single-repo API.
type Repo struct {
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	Stars           int      `json:"stargazers_count"`
	OpenIssuesCount int      `json:"open_issues_count"`
	HTMLURL         string   `json:"html_url"`
	Language        string   `json:"language"`
	Topics          []string `json:"topics"`
}

func get(ctx context.Context, rawURL, token string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

// FetchRepos returns either the single configured repo (Config.SingleRepo) or the
// GitHub search results matching Config's filters.
func (s *Scout) FetchRepos(ctx context.Context) ([]Repo, error) {
	if s.cfg.SingleRepo != "" {
		parts := strings.SplitN(s.cfg.SingleRepo, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid repo format %q — expected owner/repo", s.cfg.SingleRepo)
		}
		body, err := get(ctx, githubAPIBase+"/repos/"+s.cfg.SingleRepo, s.cfg.Token)
		if err == nil {
			var repo Repo
			if json.Unmarshal(body, &repo) == nil {
				return []Repo{repo}, nil
			}
		}
		// fallback: synthetic repo (no stars, no description)
		return []Repo{{
			FullName: s.cfg.SingleRepo,
			HTMLURL:  "https://github.com/" + s.cfg.SingleRepo,
		}}, nil
	}
	return s.searchGitHub(ctx)
}

func (s *Scout) searchGitHub(ctx context.Context) ([]Repo, error) {
	cfg := s.cfg
	query := ""
	if cfg.Keyword != "" {
		query = cfg.Keyword + " "
	}
	query += fmt.Sprintf("stars:>%d", cfg.MinStars)
	if cfg.Language != "" {
		query += " language:" + cfg.Language
	}
	if cfg.Topic != "" {
		query += " topic:" + cfg.Topic
	}
	query += " is:public archived:false"
	if cfg.PushedAfter != "" {
		query += " pushed:>=" + cfg.PushedAfter
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "stars")
	params.Set("order", "desc")
	params.Set("per_page", fmt.Sprint(defaultPageSize))

	var all []Repo
	page := 1
	for len(all) < cfg.Limit {
		params.Set("page", fmt.Sprint(page))
		apiURL := githubAPIBase + "/search/repositories?" + params.Encode()

		body, err := get(ctx, apiURL, cfg.Token)
		if err != nil {
			return nil, fmt.Errorf("github search: %w", err)
		}
		var sr ghSearchResult
		if err := json.Unmarshal(body, &sr); err != nil {
			return nil, err
		}
		if len(sr.Items) == 0 {
			break
		}
		all = append(all, sr.Items...)
		page++

		select {
		case <-ctx.Done():
			return all, ctx.Err()
		case <-time.After(searchPageDelay):
		}
	}
	if len(all) > cfg.Limit {
		all = all[:cfg.Limit]
	}
	return all, nil
}

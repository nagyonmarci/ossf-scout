package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ghSearchResult struct {
	Items []ghRepo `json:"items"`
}

type ghRepo struct {
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	Stars           int      `json:"stargazers_count"`
	OpenIssuesCount int      `json:"open_issues_count"`
	HTMLURL         string   `json:"html_url"`
	Language        string   `json:"language"`
	Topics          []string `json:"topics"`
}

func ghGet(rawURL, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

func fetchRepos(cfg config) ([]ghRepo, error) {
	if cfg.singleRepo != "" {
		parts := strings.SplitN(cfg.singleRepo, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid repo format %q — expected owner/repo", cfg.singleRepo)
		}
		body, err := ghGet("https://api.github.com/repos/"+cfg.singleRepo, cfg.token)
		if err == nil {
			var repo ghRepo
			if json.Unmarshal(body, &repo) == nil {
				return []ghRepo{repo}, nil
			}
		}
		// fallback: synthetic repo (no stars, no description)
		return []ghRepo{{
			FullName: cfg.singleRepo,
			HTMLURL:  "https://github.com/" + cfg.singleRepo,
		}}, nil
	}
	return searchGitHub(cfg)
}

func searchGitHub(cfg config) ([]ghRepo, error) {
	query := ""
	if cfg.keyword != "" {
		query = cfg.keyword + " "
	}
	query += fmt.Sprintf("stars:>%d", cfg.minStars)
	if cfg.language != "" {
		query += " language:" + cfg.language
	}
	if cfg.topic != "" {
		query += " topic:" + cfg.topic
	}
	query += " is:public archived:false"
	if cfg.pushedAfter != "" {
		query += " pushed:>=" + cfg.pushedAfter
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "stars")
	params.Set("order", "desc")
	params.Set("per_page", "100")

	var all []ghRepo
	page := 1
	for len(all) < cfg.limit {
		params.Set("page", fmt.Sprint(page))
		apiURL := "https://api.github.com/search/repositories?" + params.Encode()

		body, err := ghGet(apiURL, cfg.token)
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
		time.Sleep(500 * time.Millisecond)
	}
	if len(all) > cfg.limit {
		all = all[:cfg.limit]
	}
	return all, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type trendingEntry struct {
	FullName   string
	StarsToday int
	Language   string
}

type trendingResult struct {
	Repo         string   `json:"repo"`
	RepoURL      string   `json:"repo_url"`
	Stars        int      `json:"stars"`
	StarsToday   int      `json:"stars_today"`
	OpenIssues   int      `json:"open_issues"`
	Score        float64  `json:"score"`
	ScorecardURL string   `json:"scorecard_url"`
	WeakChecks   []string `json:"weak_checks"`
	Description  string   `json:"description"`
	Language     string   `json:"language"`
}

var (
	// Match any <article> — not tied to a specific class name
	reArticle = regexp.MustCompile(`(?s)<article\b[^>]*>(.*?)</article>`)
	// Repo link lives in the first <h2><a href="/owner/repo"> inside the article
	reH2Link        = regexp.MustCompile(`(?s)<h2\b[^>]*>.*?<a\b[^>]*href="/([A-Za-z0-9][A-Za-z0-9._-]*/[A-Za-z0-9][A-Za-z0-9._-]*)"`)
	reLangAttr      = regexp.MustCompile(`itemprop="programmingLanguage">([^<]+)<`)
	reStarsTrending = regexp.MustCompile(`([\d,]+)\s+stars?\s+(?:today|this (?:week|month|year))`)
)

// maxTrendingArticles caps how many <article> blocks we parse, so pathological
// or hostile input can't blow up memory. GitHub's trending page shows ~25.
const maxTrendingArticles = 100

func scrapeTrending(language, since, token string) ([]trendingEntry, error) {
	trendURL := "https://github.com/trending"
	if language != "" {
		trendURL += "/" + strings.ToLower(strings.TrimSpace(language))
	}
	if since != "" {
		trendURL += "?since=" + since
	}

	req, err := http.NewRequest("GET", trendURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Reuse the shared client (15s timeout) instead of http.DefaultClient (no timeout).
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub trending returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	entries := parseTrending(string(body))
	if len(entries) == 0 {
		return nil, fmt.Errorf("no trending repos found — GitHub may have changed its HTML structure")
	}
	return entries, nil
}

func parseTrending(html string) []trendingEntry {
	articles := reArticle.FindAllStringSubmatch(html, maxTrendingArticles)
	var entries []trendingEntry

	for _, a := range articles {
		body := a[1]

		// Repo link is always in the first <h2><a href="/owner/repo">
		m := reH2Link.FindStringSubmatch(body)
		if len(m) < 2 {
			continue
		}
		fullName := m[1]

		lang := ""
		if lm := reLangAttr.FindStringSubmatch(body); len(lm) > 1 {
			lang = strings.TrimSpace(lm[1])
		}

		starsToday := 0
		if sm := reStarsTrending.FindStringSubmatch(body); len(sm) > 1 {
			n, _ := strconv.Atoi(strings.ReplaceAll(sm[1], ",", ""))
			starsToday = n
		}

		entries = append(entries, trendingEntry{
			FullName:   fullName,
			StarsToday: starsToday,
			Language:   lang,
		})
	}
	return entries
}

func handleTrending(serverToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		language := q.Get("language")
		since := q.Get("since")
		if since == "" {
			since = "daily"
		}
		token := q.Get("token")
		if token == "" {
			token = serverToken
		}
		workers := 5
		if v := q.Get("workers"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				workers = min(n, 20)
			}
		}
		// Default maxScore=10 to show all trending repos; user can filter client-side
		maxScore := 10.0
		checkFilter := q.Get("check_filter")
		minMaintained := 0
		if v := q.Get("min_maintained"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				minMaintained = n
			}
		}

		entries, err := scrapeTrending(language, since, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("scraping error: %v", err), http.StatusBadGateway)
			return
		}

		// Build stars-today map
		starsToday := make(map[string]int, len(entries))
		for _, e := range entries {
			starsToday[e.FullName] = e.StarsToday
		}

		// Fetch GitHub metadata concurrently
		var mu sync.Mutex
		var repos []ghRepo
		var wg sync.WaitGroup
		for _, e := range entries {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				body, err := ghGet("https://api.github.com/repos/"+name, token)
				if err != nil {
					mu.Lock()
					repos = append(repos, ghRepo{FullName: name, HTMLURL: "https://github.com/" + name})
					mu.Unlock()
					return
				}
				var repo ghRepo
				if json.Unmarshal(body, &repo) != nil {
					repo = ghRepo{FullName: name, HTMLURL: "https://github.com/" + name}
				}
				mu.Lock()
				repos = append(repos, repo)
				mu.Unlock()
			}(e.FullName)
		}
		wg.Wait()

		cfg := config{
			workers:       workers,
			token:         token,
			maxScore:      maxScore,
			minMaintained: minMaintained,
			checkFilter:   checkFilter,
		}
		results := runWorkers(repos, cfg)

		out := make([]trendingResult, 0, len(results))
		for _, res := range results {
			wc := res.WeakChecks
			if wc == nil {
				wc = []string{}
			}
			repoURL := res.Repo.HTMLURL
			if repoURL == "" {
				repoURL = "https://github.com/" + res.Repo.FullName
			}
			out = append(out, trendingResult{
				Repo:         res.Repo.FullName,
				RepoURL:      repoURL,
				Stars:        res.Repo.Stars,
				StarsToday:   starsToday[res.Repo.FullName],
				OpenIssues:   res.Repo.OpenIssuesCount,
				Score:        res.Score,
				ScorecardURL: res.ScorecardURL,
				WeakChecks:   wc,
				Description:  res.Repo.Description,
				Language:     res.Repo.Language,
			})
		}

		writeJSON(w, http.StatusOK, out)
	}
}

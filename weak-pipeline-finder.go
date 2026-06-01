package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── GitHub Search API types ──────────────────────────────────────────────────

type ghSearchResult struct {
	Items []ghRepo `json:"items"`
}

type ghRepo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	HTMLURL     string `json:"html_url"`
	Language    string `json:"language"`
	Topics      []string `json:"topics"`
}

// ── OpenSSF Scorecard API types ──────────────────────────────────────────────

type scorecardResponse struct {
	Score  float64        `json:"score"`
	Checks []scorecardCheck `json:"checks"`
}

type scorecardCheck struct {
	Name  string `json:"name"`
	Score int    `json:"score"` // -1 = N/A, 0–10
}

// ── Result ───────────────────────────────────────────────────────────────────

type result struct {
	Repo        ghRepo
	Score       float64
	WeakChecks  []string
	ScorecardURL string
}

// ── Config ───────────────────────────────────────────────────────────────────

type config struct {
	language    string
	minStars    int
	maxScore    float64
	limit       int
	workers     int
	jsonOut     bool
	token       string
	checkFilter string // comma-separated check names to highlight
}

// ── HTTP clients ─────────────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 15 * time.Second}

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

func scorecardGet(owner, repo string) (*scorecardResponse, error) {
	apiURL := fmt.Sprintf("https://api.securityscorecards.dev/projects/github.com/%s/%s", owner, repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, nil // not scanned yet
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scorecard %d", resp.StatusCode)
	}
	var sc scorecardResponse
	if err := json.NewDecoder(resp.Body).Decode(&sc); err != nil {
		return nil, err
	}
	return &sc, nil
}

// ── GitHub search ─────────────────────────────────────────────────────────────

func searchGitHub(cfg config) ([]ghRepo, error) {
	query := fmt.Sprintf("stars:>%d", cfg.minStars)
	if cfg.language != "" {
		query += " language:" + cfg.language
	}
	query += " is:public archived:false"

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
		time.Sleep(500 * time.Millisecond) // be polite
	}
	if len(all) > cfg.limit {
		all = all[:cfg.limit]
	}
	return all, nil
}

// ── Worker pool ───────────────────────────────────────────────────────────────

func runWorkers(repos []ghRepo, cfg config) []result {
	wantChecks := parseChecks(cfg.checkFilter)

	jobs := make(chan ghRepo, len(repos))
	results := make(chan result, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				parts := strings.SplitN(repo.FullName, "/", 2)
				if len(parts) != 2 {
					continue
				}
				sc, err := scorecardGet(parts[0], parts[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "  [warn] %s: %v\n", repo.FullName, err)
					continue
				}
				if sc == nil {
					// Not in scorecard dataset — still interesting (no pipeline at all)
					results <- result{
						Repo:        repo,
						Score:       -1,
						WeakChecks:  []string{"NOT_SCANNED"},
						ScorecardURL: fmt.Sprintf("https://scorecard.dev/viewer/?uri=github.com/%s", repo.FullName),
					}
					continue
				}
				if sc.Score > cfg.maxScore && len(wantChecks) == 0 {
					continue
				}
				weak := weakChecks(sc.Checks, wantChecks)
				if sc.Score <= cfg.maxScore || len(weak) > 0 {
					results <- result{
						Repo:        repo,
						Score:       sc.Score,
						WeakChecks:  weak,
						ScorecardURL: fmt.Sprintf("https://scorecard.dev/viewer/?uri=github.com/%s", repo.FullName),
					}
				}
				time.Sleep(200 * time.Millisecond) // rate-limit courtesy
			}
		}()
	}

	for _, r := range repos {
		jobs <- r
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []result
	for r := range results {
		out = append(out, r)
	}
	// sort by score asc (worst first), NOT_SCANNED (-1) at bottom
	sort.Slice(out, func(i, j int) bool {
		si, sj := out[i].Score, out[j].Score
		if si == -1 { return false }
		if sj == -1 { return true }
		return si < sj
	})
	return out
}

func parseChecks(s string) map[string]bool {
	m := make(map[string]bool)
	if s == "" {
		return m
	}
	for _, c := range strings.Split(s, ",") {
		m[strings.TrimSpace(c)] = true
	}
	return m
}

func weakChecks(checks []scorecardCheck, want map[string]bool) []string {
	// Security-relevant check names (subset most relevant for DevSecOps PRs)
	securityChecks := map[string]bool{
		"CI-Tests":              true,
		"SAST":                  true,
		"Dependency-Update-Tool": true,
		"Vulnerabilities":       true,
		"Pinned-Dependencies":   true,
		"Branch-Protection":     true,
		"Code-Review":           true,
		"Maintained":            true,
	}

	filter := securityChecks
	if len(want) > 0 {
		filter = want
	}

	var weak []string
	for _, c := range checks {
		if filter[c.Name] && (c.Score == -1 || c.Score < 5) {
			weak = append(weak, fmt.Sprintf("%s(%d)", c.Name, c.Score))
		}
	}
	return weak
}

// ── Output ────────────────────────────────────────────────────────────────────

func printTable(results []result) {
	fmt.Printf("\n%-50s %6s %8s  %-40s  %s\n",
		"REPOSITORY", "STARS", "SCORE", "WEAK CHECKS", "SCORECARD URL")
	fmt.Println(strings.Repeat("─", 160))
	for _, r := range results {
		score := fmt.Sprintf("%.1f", r.Score)
		if r.Score == -1 {
			score = "N/A"
		}
		fmt.Printf("%-50s %6d %8s  %-40s  %s\n",
			r.Repo.FullName,
			r.Repo.Stars,
			score,
			strings.Join(r.WeakChecks, ", "),
			r.ScorecardURL,
		)
	}
	fmt.Printf("\nFound %d repos matching criteria.\n", len(results))
}

type jsonResult struct {
	Repo         string   `json:"repo"`
	Stars        int      `json:"stars"`
	Score        float64  `json:"score"`
	Language     string   `json:"language"`
	Description  string   `json:"description"`
	WeakChecks   []string `json:"weak_checks"`
	ScorecardURL string   `json:"scorecard_url"`
	RepoURL      string   `json:"repo_url"`
}

func printJSON(results []result) {
	var out []jsonResult
	for _, r := range results {
		out = append(out, jsonResult{
			Repo:        r.Repo.FullName,
			Stars:       r.Repo.Stars,
			Score:       r.Score,
			Language:    r.Repo.Language,
			Description: r.Repo.Description,
			WeakChecks:  r.WeakChecks,
			ScorecardURL: r.ScorecardURL,
			RepoURL:     r.Repo.HTMLURL,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	cfg := config{}
	flag.StringVar(&cfg.language, "lang", "", "Filter by language (e.g. go, python, java)")
	flag.IntVar(&cfg.minStars, "min-stars", 500, "Minimum stars")
	flag.Float64Var(&cfg.maxScore, "max-score", 5.0, "Maximum OpenSSF score to include (0-10)")
	flag.IntVar(&cfg.limit, "limit", 100, "Max repos to fetch from GitHub search")
	flag.IntVar(&cfg.workers, "workers", 5, "Concurrent Scorecard API workers")
	flag.BoolVar(&cfg.jsonOut, "json", false, "Output as JSON")
	flag.StringVar(&cfg.token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub PAT (or set GITHUB_TOKEN env)")
	flag.StringVar(&cfg.checkFilter, "checks", "", "Comma-separated Scorecard checks to highlight (default: security set)")
	flag.Parse()

	if cfg.token == "" {
		fmt.Fprintln(os.Stderr, "Warning: no GITHUB_TOKEN set — GitHub rate limit is 60 req/hour unauthenticated")
	}

	fmt.Fprintf(os.Stderr, "Searching GitHub: language=%q minStars=%d limit=%d\n",
		cfg.language, cfg.minStars, cfg.limit)

	repos, err := searchGitHub(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Fetched %d repos — querying Scorecard API (workers=%d)...\n",
		len(repos), cfg.workers)

	results := runWorkers(repos, cfg)

	if cfg.jsonOut {
		printJSON(results)
	} else {
		printTable(results)
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ── Scan handlers ─────────────────────────────────────────────────────────────

func handleCreateScan(db *sql.DB, serverToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Language      string  `json:"language"`
			MinStars      int     `json:"min_stars"`
			MaxScore      float64 `json:"max_score"`
			Limit         int     `json:"limit"`
			Workers       int     `json:"workers"`
			CheckFilter   string  `json:"check_filter"`
			GithubToken   string  `json:"github_token"`
			CliFallback   bool    `json:"use_cli_fallback"`
			PushedAfter   string  `json:"pushed_after"`
			MinMaintained int     `json:"min_maintained"`
			Topic         string  `json:"topic"`
			Keyword       string  `json:"keyword"`
			SingleRepo    string  `json:"single_repo"`
		}
		body.MinStars = defaultMinStars
		body.MaxScore = defaultMaxScore
		body.Limit = defaultLimit
		body.Workers = defaultWorkers

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		token := serverToken
		if body.GithubToken != "" {
			token = body.GithubToken
		}

		cfg := config{
			language:      body.Language,
			minStars:      body.MinStars,
			maxScore:      body.MaxScore,
			limit:         body.Limit,
			workers:       body.Workers,
			checkFilter:   body.CheckFilter,
			token:         token,
			cliFallback:   body.CliFallback,
			pushedAfter:   body.PushedAfter,
			minMaintained: body.MinMaintained,
			topic:         body.Topic,
			keyword:       body.Keyword,
			singleRepo:    body.SingleRepo,
		}

		id, err := dbInsertScan(db, cfg)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		go func() {
			repos, err := fetchRepos(cfg)
			if err != nil {
				if derr := dbUpdateScanError(db, id, err.Error()); derr != nil {
					fmt.Fprintf(os.Stderr, "dbUpdateScanError: %v\n", derr)
				}
				return
			}
			results := runWorkers(repos, cfg)
			if err := dbInsertResults(db, id, results); err != nil {
				if derr := dbUpdateScanError(db, id, err.Error()); derr != nil {
					fmt.Fprintf(os.Stderr, "dbUpdateScanError: %v\n", derr)
				}
				return
			}
			if derr := dbUpdateScanDone(db, id, len(repos), len(results)); derr != nil {
				fmt.Fprintf(os.Stderr, "dbUpdateScanDone: %v\n", derr)
			}
		}()

		scan, err := dbGetScan(db, id)
		if err != nil || scan == nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, scan)
	}
}

func handleListScans(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scans, err := dbListScans(db)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if scans == nil {
			scans = []scanRow{}
		}
		writeJSON(w, http.StatusOK, scans)
	}
}

func handleGetScan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseScanID(w, r)
		if !ok {
			return
		}
		scan, err := dbGetScan(db, id)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if scan == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, scan)
	}
}

func handleGetResults(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseScanID(w, r)
		if !ok {
			return
		}
		results, err := dbGetResults(db, id)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if results == nil {
			results = []scanResultRow{}
		}
		writeJSON(w, http.StatusOK, results)
	}
}

func handleDeleteScan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseScanID(w, r)
		if !ok {
			return
		}
		if err := dbDeleteScan(db, id); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ── Issues/PR index handler ───────────────────────────────────────────────────

// handleGetIssuesPRs returns a cached or freshly fetched issues/PR summary
// for a given repo. Add ?refresh=true to force a re-fetch.
func handleGetIssuesPRs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := r.PathValue("owner")
		repoName := r.PathValue("repo")
		if owner == "" || repoName == "" {
			http.Error(w, "owner and repo path params required", http.StatusBadRequest)
			return
		}
		repo := owner + "/" + repoName
		refresh := r.URL.Query().Get("refresh") == "true"
		ghToken := os.Getenv("GITHUB_TOKEN")

		if !refresh {
			if summary, err := dbGetRecentIssuesPRsSummary(db, repo, 24*time.Hour); err == nil && summary != nil {
				writeJSON(w, http.StatusOK, map[string]string{"repo": repo, "summary": *summary, "cached": "true"})
				return
			}
		}

		summary, dataJSON, err := fetchIssuesPRsSummary(repo, ghToken)
		if err != nil {
			http.Error(w, "fetch failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		_ = dbStoreIssuesPRs(db, repo, dataJSON, summary)
		writeJSON(w, http.StatusOK, map[string]string{"repo": repo, "summary": summary, "cached": "false"})
	}
}

// ── Score trend handler ───────────────────────────────────────────────────────

// handleGetScoreTrend returns score history for a repo.
// Path: GET /api/stats/trend?repo=owner/name&limit=N
func handleGetScoreTrend(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			http.Error(w, "repo param required", http.StatusBadRequest)
			return
		}
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}
		points, err := dbGetScoreTrend(db, repo, limit)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if points == nil {
			points = []scoreTrendPoint{}
		}
		writeJSON(w, http.StatusOK, points)
	}
}

// ── Portfolio handler ─────────────────────────────────────────────────────────

// handleGetPortfolio returns aggregated stats for multiple repos.
// Query param: ?repos=owner/a,owner/b (optional — omit to get top 20 by activity)
func handleGetPortfolio(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var repos []string
		if raw := r.URL.Query().Get("repos"); raw != "" {
			for _, r := range strings.Split(raw, ",") {
				if trimmed := strings.TrimSpace(r); trimmed != "" {
					repos = append(repos, trimmed)
				}
			}
		}
		data, err := dbGetPortfolio(db, repos)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if data == nil {
			data = []portfolioRepo{}
		}
		writeJSON(w, http.StatusOK, data)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseScanID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"embed"
)

//go:embed frontend/dist
var staticFiles embed.FS

func startServer(port int, dbPath string, serverToken string) {
	db, err := openDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/scans", handleCreateScan(db, serverToken))
	mux.HandleFunc("GET /api/scans", handleListScans(db))
	mux.HandleFunc("GET /api/scans/{id}", handleGetScan(db))
	mux.HandleFunc("GET /api/scans/{id}/results", handleGetResults(db))
	mux.HandleFunc("DELETE /api/scans/{id}", handleDeleteScan(db))

	sub, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load embedded frontend: %v\n", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file; fall back to index.html for SPA routing
		path := r.URL.Path
		if path != "/" {
			if _, err := fs.Stat(sub, path[1:]); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Fprintf(os.Stderr, "ossf-scout web UI running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// ── handlers ──────────────────────────────────────────────────────────────────

func handleCreateScan(db *sql.DB, serverToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Language    string  `json:"language"`
			MinStars    int     `json:"min_stars"`
			MaxScore    float64 `json:"max_score"`
			Limit       int     `json:"limit"`
			Workers     int     `json:"workers"`
			CheckFilter string  `json:"check_filter"`
			GithubToken string  `json:"github_token"`
			CliFallback bool    `json:"use_cli_fallback"`
		}
		body.MinStars = 500
		body.MaxScore = 5.0
		body.Limit = 100
		body.Workers = 5

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		token := serverToken
		if body.GithubToken != "" {
			token = body.GithubToken
		}

		cfg := config{
			language:    body.Language,
			minStars:    body.MinStars,
			maxScore:    body.MaxScore,
			limit:       body.Limit,
			workers:     body.Workers,
			checkFilter: body.CheckFilter,
			token:       token,
			cliFallback: body.CliFallback,
		}

		id, err := dbInsertScan(db, cfg)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		go func() {
			repos, err := searchGitHub(cfg)
			if err != nil {
				_ = dbUpdateScanError(db, id, err.Error())
				return
			}
			results := runWorkers(repos, cfg)
			if err := dbInsertResults(db, id, results); err != nil {
				_ = dbUpdateScanError(db, id, err.Error())
				return
			}
			_ = dbUpdateScanDone(db, id, len(repos), len(results))
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

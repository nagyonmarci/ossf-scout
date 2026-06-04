package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
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

	mux.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"user": currentUser(r)})
	})

	mux.HandleFunc("POST /api/scans", handleCreateScan(db, serverToken))
	mux.HandleFunc("GET /api/scans", handleListScans(db))
	mux.HandleFunc("GET /api/scans/{id}", handleGetScan(db))
	mux.HandleFunc("GET /api/scans/{id}/results", handleGetResults(db))
	mux.HandleFunc("DELETE /api/scans/{id}", handleDeleteScan(db))
	mux.HandleFunc("GET /api/trending", handleTrending(serverToken))

	mux.HandleFunc("POST /api/audits", handleCreateAudit(db))
	mux.HandleFunc("POST /api/audits/{id}/generate", handleGenerateAudit(db))
	mux.HandleFunc("POST /api/audits/{id}/extract-findings", handleExtractFindings(db))
	mux.HandleFunc("POST /api/audits/{id}/compare", handleCompareAudit(db))
	mux.HandleFunc("GET /api/audits", handleListAudits(db))
	mux.HandleFunc("GET /api/audits/{id}", handleGetAudit(db))
	mux.HandleFunc("GET /api/audits/{id}/context.md", handleDownloadAuditContext(db))
	mux.HandleFunc("GET /api/audits/{id}/supply-chain", handleSupplyChainGraph(db))
	mux.HandleFunc("GET /api/audits/{id}/export.json", handleExportAuditJSON(db))
	mux.HandleFunc("GET /api/audits/{id}/export.sarif", handleExportAuditSARIF(db))
	mux.HandleFunc("DELETE /api/audits/{id}", handleDeleteAudit(db))

	mux.HandleFunc("GET /api/schedules", handleListSchedules(db))
	mux.HandleFunc("POST /api/schedules", handleCreateSchedule(db))
	mux.HandleFunc("PUT /api/schedules/{id}", handleUpdateSchedule(db))
	mux.HandleFunc("DELETE /api/schedules/{id}", handleDeleteSchedule(db))
	mux.HandleFunc("POST /api/schedules/{id}/run", handleTriggerSchedule(db))

	mux.HandleFunc("GET /api/remediation", handleListRemediation(db))
	mux.HandleFunc("POST /api/remediation", handleCreateRemediation(db))
	mux.HandleFunc("PUT /api/remediation/{id}", handleUpdateRemediation(db))
	mux.HandleFunc("DELETE /api/remediation/{id}", handleDeleteRemediation(db))

	mux.HandleFunc("GET /api/stats/costs", handleGetCostStats(db))
	mux.HandleFunc("GET /api/stats/portfolio", handleGetPortfolio(db))
	mux.HandleFunc("GET /api/stats/trend", handleGetScoreTrend(db))
	mux.HandleFunc("GET /api/issues-prs/{owner}/{repo}", handleGetIssuesPRs(db))
	mux.HandleFunc("GET /api/orgs/{org}/repos", handleListOrgRepos())

	mux.HandleFunc("POST /api/webhooks/github", handleGitHubWebhook(db))

	startScheduler(db)

	sub, err := fs.Sub(staticFiles, "frontend/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load embedded frontend: %v\n", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	if err := http.ListenAndServe(addr, authMiddleware(mux)); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

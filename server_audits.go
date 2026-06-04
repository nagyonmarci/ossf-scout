package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ── Audit handlers ────────────────────────────────────────────────────────────

type createAuditRequest struct {
	Repo            string `json:"repo"`
	GithubToken     string `json:"github_token"`
	AnthropicKey    string `json:"anthropic_key"`
	OpenAIKey       string `json:"openai_key"`
	GeminiKey       string `json:"gemini_key"`
	Model           string `json:"model"`
	AnalysisModel   string `json:"analysis_model"`
	SplitGeneration bool   `json:"split_generation"`
	Provider        string `json:"provider"` // "anthropic" | "openai" | "gemini" | "ollama" | "" (template)
	OllamaURL       string `json:"ollama_url"`
}

func buildAuditParams(req createAuditRequest) auditParams {
	p := auditParams{
		Provider:        req.Provider,
		Model:           req.Model,
		AnalysisModel:   req.AnalysisModel,
		SplitGeneration: req.SplitGeneration,
	}

	p.AnthropicKey = req.AnthropicKey
	if p.AnthropicKey == "" && p.Provider == "anthropic" {
		p.AnthropicKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	p.OpenAIKey = req.OpenAIKey
	if p.OpenAIKey == "" && p.Provider == "openai" {
		p.OpenAIKey = os.Getenv("OPENAI_API_KEY")
	}

	p.GeminiKey = req.GeminiKey
	if p.GeminiKey == "" && p.Provider == "gemini" {
		p.GeminiKey = os.Getenv("GEMINI_API_KEY")
	}

	p.OllamaURL = req.OllamaURL
	if p.OllamaURL == "" {
		p.OllamaURL = os.Getenv("OLLAMA_BASE_URL")
	}

	if p.Model == "" {
		switch p.Provider {
		case "anthropic":
			if p.AnthropicKey != "" {
				p.Model = defaultModel
			}
		case "openai":
			p.Model = defaultOpenAIModel
		case "gemini":
			p.Model = defaultGeminiModel
		}
	}

	return p
}

func handleCreateAudit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createAuditRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Repo == "" {
			http.Error(w, "invalid request body — repo is required", http.StatusBadRequest)
			return
		}

		p := buildAuditParams(req)

		id := uuid.New().String()
		if err := dbCreateAudit(db, id, req.Repo, p.Model, p.Provider); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		go runAudit(db, id, req.Repo, req.GithubToken, p)

		a, err := dbGetAudit(db, id)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, a)
	}
}

func handleGenerateAudit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		contextJSON, err := dbGetAuditContext(db, id)
		if err != nil || contextJSON == nil {
			http.Error(w, "no saved context for this audit", http.StatusNotFound)
			return
		}
		var req createAuditRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		p := buildAuditParams(req)
		if err := dbRestartAudit(db, id, p.Model, p.Provider); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		go runAuditFromContext(db, id, *contextJSON, p)
		a, err := dbGetAudit(db, id)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusAccepted, a)
	}
}

func handleListAudits(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		audits, err := dbListAudits(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if audits == nil {
			audits = []auditRow{}
		}
		writeJSON(w, http.StatusOK, audits)
	}
}

func handleGetAudit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		a, err := dbGetAudit(db, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, a)
	}
}

func handleDownloadAuditContext(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		contextJSON, err := dbGetAuditContext(db, id)
		if err != nil || contextJSON == nil {
			http.Error(w, "context not found", http.StatusNotFound)
			return
		}
		var ctx auditContext
		if err := json.Unmarshal([]byte(*contextJSON), &ctx); err != nil {
			http.Error(w, "invalid context", http.StatusInternalServerError)
			return
		}
		md := buildContextMarkdown(&ctx)
		filename := fmt.Sprintf("context-%s-%s.md",
			strings.ReplaceAll(ctx.Meta.Repo, "/", "-"),
			ctx.Meta.Date[:10])
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		_, _ = fmt.Fprint(w, md)
	}
}

func handleExportAuditSARIF(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		audit, err := dbGetAudit(db, id)
		if err != nil || audit == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if audit.Report == nil || *audit.Report == "" {
			http.Error(w, "no report available", http.StatusConflict)
			return
		}
		sarif := buildSARIF(audit.Repo, *audit.Report)
		filename := fmt.Sprintf("audit-%s.sarif", strings.ReplaceAll(audit.Repo, "/", "-"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		writeJSON(w, http.StatusOK, sarif)
	}
}

// buildSARIF converts a markdown audit report into a minimal SARIF 2.1.0 document
// suitable for upload to GitHub Code Scanning.
func buildSARIF(repo, reportMD string) map[string]any {
	type sarifResult struct {
		RuleID  string         `json:"ruleId"`
		Level   string         `json:"level"`
		Message map[string]any `json:"message"`
		Locs    []any          `json:"locations,omitempty"`
	}

	var results []sarifResult
	seen := map[string]bool{}

	for _, line := range strings.Split(reportMD, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cols := strings.Split(line, "|")
		// Filter separator rows and header rows
		if len(cols) < 4 || strings.Contains(line, "---") {
			continue
		}
		cells := make([]string, 0, len(cols))
		for _, c := range cols {
			cells = append(cells, strings.TrimSpace(c))
		}
		// Look for rows that have a non-empty first column and a severity-like second or third column
		severity := ""
		title := ""
		for i, c := range cells {
			cl := strings.ToLower(c)
			if cl == "critical" || cl == "high" || cl == "medium" || cl == "low" || cl == "info" || cl == "informational" {
				severity = cl
				if i > 1 && cells[i-1] != "" {
					title = cells[i-1]
				} else if i < len(cells)-1 && cells[i+1] != "" {
					title = cells[i+1]
				}
				break
			}
		}
		if severity == "" || title == "" || seen[title] {
			continue
		}
		seen[title] = true
		level := "warning"
		switch severity {
		case "critical", "high":
			level = "error"
		case "low", "info", "informational":
			level = "note"
		}
		ruleID := "ossf-scout/" + strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(title, " ", "-"), "/", "-"))
		if len(ruleID) > 80 {
			ruleID = ruleID[:80]
		}
		results = append(results, sarifResult{
			RuleID:  ruleID,
			Level:   level,
			Message: map[string]any{"text": title},
		})
	}

	return map[string]any{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]any{
			{
				"tool": map[string]any{
					"driver": map[string]any{
						"name":            "ossf-scout",
						"informationUri":  "https://github.com/" + repo,
						"semanticVersion": "1.0.0",
					},
				},
				"results": results,
			},
		},
	}
}

func handleDeleteAudit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := dbDeleteAudit(db, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ── Cost stats handler ────────────────────────────────────────────────────────

func handleGetCostStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		days := 30
		if d := r.URL.Query().Get("days"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				days = n
			}
		}
		stats, err := dbGetCostStats(db, days)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, stats)
	}
}

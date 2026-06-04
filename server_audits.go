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

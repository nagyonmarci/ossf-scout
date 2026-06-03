package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// ── Audit handlers ────────────────────────────────────────────────────────────

// ── audit handlers ────────────────────────────────────────────────────────────

type createAuditRequest struct {
	Repo            string `json:"repo"`
	GithubToken     string `json:"github_token"`
	AnthropicKey    string `json:"anthropic_key"`
	Model           string `json:"model"`
	AnalysisModel   string `json:"analysis_model"`
	SplitGeneration bool   `json:"split_generation"`
	Provider        string `json:"provider"`   // "anthropic" | "ollama" | "" (template)
	OllamaURL       string `json:"ollama_url"` // default: http://localhost:11434
}

func handleCreateAudit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createAuditRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Repo == "" {
			http.Error(w, "invalid request body — repo is required", http.StatusBadRequest)
			return
		}

		provider := req.Provider

		apiKey := req.AnthropicKey
		if apiKey == "" && provider == "anthropic" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}

		ollamaURL := req.OllamaURL
		if ollamaURL == "" {
			ollamaURL = os.Getenv("OLLAMA_BASE_URL")
		}

		model := req.Model
		if model == "" && provider == "anthropic" && apiKey != "" {
			model = defaultModel
		}

		id := uuid.New().String()
		if err := dbCreateAudit(db, id, req.Repo, model, provider); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		go runAudit(db, id, req.Repo, req.GithubToken, provider, apiKey, ollamaURL, model, req.AnalysisModel, req.SplitGeneration)

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
		provider := req.Provider
		apiKey := req.AnthropicKey
		if apiKey == "" && provider == "anthropic" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		ollamaURL := req.OllamaURL
		if ollamaURL == "" {
			ollamaURL = os.Getenv("OLLAMA_BASE_URL")
		}
		model := req.Model
		if model == "" && provider == "anthropic" && apiKey != "" {
			model = defaultModel
		}
		if err := dbRestartAudit(db, id, model, provider); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		go runAuditFromContext(db, id, *contextJSON, provider, apiKey, ollamaURL, model, req.AnalysisModel, req.SplitGeneration)
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

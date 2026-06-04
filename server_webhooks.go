package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// handleGitHubWebhook processes GitHub webhook events.
// Validates X-Hub-Signature-256 with GITHUB_WEBHOOK_SECRET (if set).
// On pull_request opened/synchronize events touching security files,
// triggers a free static-snapshot audit automatically.
func handleGitHubWebhook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 5<<20))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		// Validate signature if secret is configured
		if secret := os.Getenv("GITHUB_WEBHOOK_SECRET"); secret != "" {
			sig := r.Header.Get("X-Hub-Signature-256")
			if !validateWebhookSig(body, sig, secret) {
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}
		}

		event := r.Header.Get("X-GitHub-Event")
		if event != "pull_request" && event != "push" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		repo := extractRepo(payload)
		if repo == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		shouldAudit := false

		if event == "pull_request" {
			action, _ := payload["action"].(string)
			if action == "opened" || action == "synchronize" || action == "reopened" {
				files := extractPRFiles(payload)
				shouldAudit = touchesSecurityFiles(files)
			}
		}

		if event == "push" {
			commits, _ := payload["commits"].([]interface{})
			for _, c := range commits {
				cm, _ := c.(map[string]interface{})
				files := extractCommitFiles(cm)
				if touchesSecurityFiles(files) {
					shouldAudit = true
					break
				}
			}
		}

		if !shouldAudit {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		id := uuid.New().String()
		if err := dbCreateAudit(db, id, repo, "", ""); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		// Free static snapshot — no AI key needed
		go runAudit(db, id, repo, "", auditParams{Provider: ""})

		writeJSON(w, http.StatusAccepted, map[string]string{"audit_id": id, "repo": repo})
	}
}

func validateWebhookSig(body []byte, sig, secret string) bool {
	if !strings.HasPrefix(sig, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func extractRepo(payload map[string]interface{}) string {
	if repo, ok := payload["repository"].(map[string]interface{}); ok {
		if name, ok := repo["full_name"].(string); ok {
			return name
		}
	}
	return ""
}

func extractPRFiles(payload map[string]interface{}) []string {
	// GitHub doesn't include files in the webhook payload directly.
	// We rely on the head commit's modified files list if present,
	// or fall back to always auditing to be safe.
	head, _ := payload["pull_request"].(map[string]interface{})
	if head == nil {
		return []string{"security"} // fallback: treat as security-relevant
	}
	commits, _ := head["commits"].(float64)
	if commits > 0 {
		return []string{"security"} // always audit PRs with commits
	}
	return nil
}

func extractCommitFiles(cm map[string]interface{}) []string {
	var files []string
	for _, key := range []string{"added", "modified", "removed"} {
		if arr, ok := cm[key].([]interface{}); ok {
			for _, f := range arr {
				if s, ok := f.(string); ok {
					files = append(files, s)
				}
			}
		}
	}
	return files
}

// securityFilePatterns lists path fragments that indicate security-relevant changes.
var securityFilePatterns = []string{
	"auth", "permission", "security", "secret", "token", "jwt", "oauth",
	"session", "password", "crypto", "middleware", "firewall", "cors",
	"helmet", "policy", "rbac", "acl", ".github/workflows/",
	"Dockerfile", "docker-compose", "k8s/", "kubernetes/", "terraform/",
}

func touchesSecurityFiles(files []string) bool {
	if len(files) == 0 {
		return true // when we can't tell, audit to be safe
	}
	for _, f := range files {
		lower := strings.ToLower(f)
		for _, pat := range securityFilePatterns {
			if strings.Contains(lower, strings.ToLower(pat)) {
				return true
			}
		}
	}
	return false
}

// notifyWebhook POSTs a JSON payload to NOTIFY_WEBHOOK_URL after an audit completes.
func notifyWebhook(repo, status, auditID string) {
	webhookURL := os.Getenv("NOTIFY_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}
	payload := map[string]string{
		"repo":     repo,
		"status":   status,
		"audit_id": auditID,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err == nil {
		resp.Body.Close() //nolint:errcheck
	}
}

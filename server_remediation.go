package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// ── Remediation item handlers ─────────────────────────────────────────────────

// GET  /api/remediation?audit_id=X  — items for a specific audit
// GET  /api/remediation?repo=owner/repo — items for a repo across all audits
// GET  /api/remediation?status=open|in_progress|resolved — filter by status
// POST /api/remediation — create item
// PUT  /api/remediation/{id} — update item (status, notes, severity, due_date)
// DELETE /api/remediation/{id} — delete item

func handleListRemediation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auditID := r.URL.Query().Get("audit_id")
		repo := r.URL.Query().Get("repo")
		status := r.URL.Query().Get("status")

		var items []remediationItem
		var err error
		switch {
		case auditID != "":
			items, err = dbListRemediationByAudit(db, auditID)
		case repo != "":
			items, err = dbListRemediationByRepo(db, repo)
		default:
			items, err = dbListAllRemediation(db, status)
		}
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if items == nil {
			items = []remediationItem{}
		}
		writeJSON(w, http.StatusOK, items)
	}
}

func handleCreateRemediation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			AuditID  string `json:"audit_id"`
			Repo     string `json:"repo"`
			Title    string `json:"title"`
			Severity string `json:"severity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" || body.AuditID == "" {
			http.Error(w, "audit_id and title required", http.StatusBadRequest)
			return
		}
		if body.Severity == "" {
			body.Severity = "medium"
		}
		id := uuid.New().String()
		if err := dbCreateRemediationItem(db, id, body.AuditID, body.Repo, body.Title, body.Severity); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		items, _ := dbListRemediationByAudit(db, body.AuditID)
		for _, it := range items {
			if it.ID == id {
				writeJSON(w, http.StatusCreated, it)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func handleUpdateRemediation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		var body struct {
			Title    string `json:"title"`
			Severity string `json:"severity"`
			Status   string `json:"status"`
			Notes    string `json:"notes"`
			DueDate  string `json:"due_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := dbUpdateRemediationItem(db, id, body.Title, body.Severity, body.Status, body.Notes, body.DueDate); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleDeleteRemediation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		if err := dbDeleteRemediationItem(db, id); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

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
	SkipSecrets     bool   `json:"skip_secrets"`
	Provider        string `json:"provider"` // "anthropic" | "openai" | "gemini" | "ollama" | "" (template)
	OllamaURL       string `json:"ollama_url"`
}

func buildAuditParams(req createAuditRequest) auditParams {
	p := auditParams{
		Provider:        req.Provider,
		Model:           req.Model,
		AnalysisModel:   req.AnalysisModel,
		SplitGeneration: req.SplitGeneration,
		SkipSecrets:     req.SkipSecrets,
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

	// Ollama base URL is always taken from the server environment variable.
	// Accepting it from the request body would create an SSRF vector where
	// any authenticated caller could redirect the server to arbitrary HTTP targets.
	p.OllamaURL = os.Getenv("OLLAMA_BASE_URL")

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
		if err := validateRepo(req.Repo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		p := buildAuditParams(req)

		id := uuid.New().String()
		if err := dbCreateAudit(db, id, req.Repo, p.Model, p.Provider); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		go runAudit(db, id, req.GithubToken, p)

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

// handleExtractFindings parses the audit report's finding/remediation table and
// bulk-creates remediation_items. Safe to call multiple times: duplicate titles
// for the same audit are skipped via INSERT OR IGNORE.
func handleExtractFindings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		audit, err := dbGetAudit(db, id)
		if err != nil || audit == nil {
			http.Error(w, "audit not found", http.StatusNotFound)
			return
		}
		if audit.Report == nil || *audit.Report == "" {
			http.Error(w, "no report available", http.StatusConflict)
			return
		}
		created := extractAndCreateFindings(db, id, audit.Repo, *audit.Report)
		writeJSON(w, http.StatusOK, map[string]int{"created": created})
	}
}

// extractAndCreateFindings parses markdown table rows from the report and
// inserts a remediationItem for each row that has a recognisable severity column.
func extractAndCreateFindings(db *sql.DB, auditID, repo, reportMD string) int {
	created := 0
	knownSev := map[string]bool{
		"critical": true, "high": true, "medium": true, "low": true,
		"info": true, "informational": true,
	}
	sevNorm := map[string]string{"informational": "info"}

	// Only parse rows inside the "Findings Summary" section to avoid false
	// positives from other tables (P2 Recommendations, Sprint Backlog, etc.)
	// that also contain severity words in their Risk Reduction / Story Points columns.
	inFindings := false
	for _, line := range strings.Split(reportMD, "\n") {
		trimmed := strings.TrimSpace(line)

		// Track section boundaries
		if strings.HasPrefix(trimmed, "#") {
			lower := strings.ToLower(trimmed)
			inFindings = strings.Contains(lower, "findings summary")
			continue
		}
		if !inFindings {
			continue
		}
		if !strings.HasPrefix(trimmed, "|") || strings.Contains(trimmed, "---") {
			continue
		}

		cols := strings.Split(trimmed, "|")
		if len(cols) < 6 {
			continue
		}
		cells := make([]string, 0, len(cols))
		for _, c := range cols {
			cells = append(cells, strings.TrimSpace(c))
		}

		// Expected column order: "" | ID | Priority | Severity | Title | OWASP | Status | ""
		// Find severity column index, then title is at index+1.
		sev, title := "", ""
		for i, c := range cells {
			// Normalise: strip markdown, take first word only (handles "Medium (5.3)" → "medium")
			cleaned := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(c, "**", ""), "`", ""))
			parts := strings.Fields(cleaned)
			if len(parts) == 0 {
				continue
			}
			cl := parts[0]
			if !knownSev[cl] {
				continue
			}
			sev = cl
			// Title is the next non-empty, non-severity, non-FIND-ID, non-P-priority cell
			for _, offset := range []int{1, 2, -1} {
				j := i + offset
				if j < 0 || j >= len(cells) {
					continue
				}
				cand := cells[j]
				candLower := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(cand, "**", ""), "`", ""))
				if cand == "" || knownSev[candLower] || len(cand) <= 3 {
					continue
				}
				// Skip priority (P0–P3) and FIND-ID cells
				if strings.HasPrefix(candLower, "p") && len(cand) <= 3 {
					continue
				}
				if strings.HasPrefix(strings.ToUpper(cand), "FIND-") {
					continue
				}
				title = cand
				break
			}
			break
		}

		if sev == "" || title == "" {
			continue
		}
		title = strings.ReplaceAll(strings.ReplaceAll(title, "**", ""), "`", "")
		if norm, ok := sevNorm[sev]; ok {
			sev = norm
		}
		if len(title) > 200 {
			title = title[:200]
		}

		itemID := uuid.New().String()
		if err := dbCreateRemediationItem(db, itemID, auditID, repo, title, sev); err == nil {
			created++
		}
	}
	return created
}

// handleCompareAudit runs two provider configurations in parallel and returns
// both reports as JSON for side-by-side comparison. Does NOT persist results.
func handleCompareAudit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		contextJSON, err := dbGetAuditContext(db, id)
		if err != nil || contextJSON == nil {
			http.Error(w, "no saved context for this audit", http.StatusNotFound)
			return
		}
		var req struct {
			A createAuditRequest `json:"a"`
			B createAuditRequest `json:"b"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		pA, pB := buildAuditParams(req.A), buildAuditParams(req.B)
		ctx := mustUnmarshalContext(*contextJSON)

		type result struct {
			Report       string `json:"report"`
			InputTokens  int    `json:"input_tokens"`
			OutputTokens int    `json:"output_tokens"`
			Error        string `json:"error,omitempty"`
			Provider     string `json:"provider"`
			Model        string `json:"model"`
		}

		runProvider := func(p auditParams) result {
			res := result{Provider: p.Provider, Model: p.Model}
			var report string
			var in, out int
			var e error
			switch p.Provider {
			case "anthropic":
				report, in, out, e = generateReport(ctx, p.AnthropicKey, p.Model)
			case "openai":
				report, in, out, e = generateOpenAIReport(ctx, p.OpenAIKey, p.Model)
			case "gemini":
				report, in, out, e = generateGeminiReport(ctx, p.GeminiKey, p.Model)
			case "ollama":
				report, in, out, e = generateOllamaReport(ctx, p.OllamaURL, p.Model)
			default:
				report = generateTemplateReport(ctx)
			}
			if e != nil {
				res.Error = e.Error()
				return res
			}
			res.Report = verifyReport(stripThinkBlocks(report), ctx)
			res.InputTokens = in
			res.OutputTokens = out
			return res
		}

		type pair struct {
			A result `json:"a"`
			B result `json:"b"`
		}
		chA := make(chan result, 1)
		chB := make(chan result, 1)
		go func() { chA <- runProvider(pA) }()
		go func() { chB <- runProvider(pB) }()
		writeJSON(w, http.StatusOK, pair{A: <-chA, B: <-chB})
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

// handleSupplyChainGraph parses the saved context's PinnedSuggestions field and
// returns a JSON graph of {repo} → [action nodes] suitable for tree rendering.
func handleSupplyChainGraph(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		contextJSON, err := dbGetAuditContext(db, id)
		if err != nil || contextJSON == nil {
			http.Error(w, "context not found", http.StatusNotFound)
			return
		}
		ctx := mustUnmarshalContext(*contextJSON)

		type graphNode struct {
			Action   string   `json:"action"`
			Tag      string   `json:"tag"`
			SHA      string   `json:"sha"`
			Resolved bool     `json:"resolved"`
			Files    []string `json:"files"`
		}
		var nodes []graphNode

		for _, line := range strings.Split(ctx.CICD.PinnedSuggestions, "\n") {
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" || line == "none" {
				continue
			}
			// format: action@tag -> sha | file:line, file:line
			arrowIdx := strings.Index(line, " -> ")
			pipeIdx := strings.Index(line, " | ")
			if arrowIdx < 0 {
				continue
			}
			actionTag := strings.TrimSpace(line[:arrowIdx])
			var sha, filesStr string
			if pipeIdx > arrowIdx {
				sha = strings.TrimSpace(line[arrowIdx+4 : pipeIdx])
				filesStr = strings.TrimSpace(line[pipeIdx+3:])
			} else {
				sha = strings.TrimSpace(line[arrowIdx+4:])
			}
			parts := strings.SplitN(actionTag, "@", 2)
			action, tag := parts[0], ""
			if len(parts) == 2 {
				tag = parts[1]
			}
			var files []string
			for _, f := range strings.Split(filesStr, ",") {
				if f = strings.TrimSpace(f); f != "" {
					// strip line number, keep file path
					if colonIdx := strings.LastIndex(f, ":"); colonIdx > 0 {
						f = f[:colonIdx]
					}
					files = append(files, f)
				}
			}
			resolved := len(sha) == 40 && !strings.Contains(sha, "unresolved")
			nodes = append(nodes, graphNode{
				Action: action, Tag: tag, SHA: sha, Resolved: resolved, Files: files,
			})
		}

		// Also collect unpinned (not in PinnedSuggestions but in UnpinnedActions)
		for _, line := range strings.Split(ctx.CICD.UnpinnedActions, "\n") {
			// grep output: file:line:  uses: action@tag
			colonCount := strings.Count(line, ":")
			if colonCount < 2 {
				continue
			}
			parts := strings.SplitN(line, ":", 3)
			usesLine := strings.TrimSpace(parts[2])
			if !strings.HasPrefix(usesLine, "uses:") {
				continue
			}
			actionTag := strings.TrimSpace(strings.TrimPrefix(usesLine, "uses:"))
			// check if already in nodes
			already := false
			for _, n := range nodes {
				if n.Action+"@"+n.Tag == actionTag {
					already = true
					break
				}
			}
			if !already && actionTag != "" {
				ap := strings.SplitN(actionTag, "@", 2)
				action, tag := ap[0], ""
				if len(ap) == 2 {
					tag = ap[1]
				}
				nodes = append(nodes, graphNode{
					Action: action, Tag: tag, SHA: "", Resolved: false,
					Files: []string{parts[0]},
				})
			}
		}

		if nodes == nil {
			nodes = []graphNode{}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"repo":  ctx.Meta.Repo,
			"nodes": nodes,
		})
	}
}

func handleExportAuditPDF(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		audit, err := dbGetAudit(db, id)
		if err != nil || audit == nil || audit.Report == nil || *audit.Report == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		pdfBytes, err := renderMarkdownToPDF(*audit.Report, audit.Repo)
		if err != nil {
			http.Error(w, "pdf render failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		filename := "audit-" + strings.ReplaceAll(audit.Repo, "/", "-") + ".pdf"
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.Write(pdfBytes) //nolint:errcheck
	}
}

func handleExportAuditJSON(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		audit, err := dbGetAudit(db, id)
		if err != nil || audit == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		contextJSON, _ := dbGetAuditContext(db, id)
		var ctx *auditContext
		if contextJSON != nil {
			c := mustUnmarshalContext(*contextJSON)
			ctx = c
		}
		export := map[string]any{
			"audit":   audit,
			"context": ctx,
		}
		filename := fmt.Sprintf("audit-%s-%s.json",
			strings.ReplaceAll(audit.Repo, "/", "-"),
			audit.CreatedAt.UTC().Format("2006-01-02"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		writeJSON(w, http.StatusOK, export)
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

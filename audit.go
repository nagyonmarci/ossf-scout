package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ── Context structs ───────────────────────────────────────────────────────────

type auditMeta struct {
	Date string `json:"date"`
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
}

type auditCICD struct {
	UnpinnedActions string `json:"unpinnedActions"`
	Zizmor          string `json:"zizmor"`
	WorkflowList    string `json:"workflowList"`
}

type auditCode struct {
	EvalUsage            string `json:"evalUsage"`
	MathRandom           string `json:"mathRandom"`
	RawSqlCalls          string `json:"rawSqlCalls"`
	XPoweredByHeader     string `json:"xPoweredByHeader"`
	HardcodedSecretHints string `json:"hardcodedSecretHints"`
	WeakCrypto           string `json:"weakCrypto"`
	ProcessExitCalls     string `json:"processExitCalls"`
}

type auditInfra struct {
	HelmLint           string `json:"helmLint"`
	HelmSecretTemplate string `json:"helmSecretTemplate"`
	HelmValues         string `json:"helmValues"`
	Dockerfile         string `json:"dockerfile"`
}

type auditDeps struct {
	PnpmAudit          string `json:"pnpmAudit"`
	WorkspaceOverrides string `json:"workspaceOverrides"`
}

type auditGit struct {
	RecentCommits        string `json:"recentCommits"`
	RecentlyChangedFiles string `json:"recentlyChangedFiles"`
}

type auditGitHub struct {
	OpenIssues     interface{} `json:"openIssues"`
	OpenPRs        interface{} `json:"openPRs"`
	SecurityAlerts string      `json:"securityAlerts"`
}

type auditContext struct {
	Meta         auditMeta   `json:"meta"`
	CICD         auditCICD   `json:"cicd"`
	Code         auditCode   `json:"code"`
	Infra        auditInfra  `json:"infra"`
	Dependencies auditDeps   `json:"dependencies"`
	Git          auditGit    `json:"git"`
	GitHub       auditGitHub `json:"github"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func runIn(dir, fallback string, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			combined := strings.TrimSpace(string(out) + string(ee.Stderr))
			if combined != "" {
				return combined
			}
		}
		return fallback
	}
	return strings.TrimSpace(string(out))
}

func shIn(dir, fallback, script string) string {
	return runIn(dir, fallback, "/bin/sh", "-c", script)
}

// ── Collect ───────────────────────────────────────────────────────────────────

func collectContext(repo, ghToken string) (*auditContext, string, error) {
	tmpDir, err := os.MkdirTemp("", "ossf-audit-*")
	if err != nil {
		return nil, "", fmt.Errorf("mktemp: %w", err)
	}

	cloneURL := fmt.Sprintf("https://github.com/%s.git", repo)
	cloneCmd := exec.Command("git", "clone", "--depth=50", "--quiet", cloneURL, tmpDir)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, "", fmt.Errorf("git clone failed: %s", strings.TrimSpace(string(out)))
	}

	ref := shIn(tmpDir, "unknown", "git rev-parse --short HEAD")

	ctx := &auditContext{
		Meta: auditMeta{
			Date: time.Now().UTC().Format(time.RFC3339),
			Repo: repo,
			Ref:  ref,
		},
		CICD: auditCICD{
			UnpinnedActions: shIn(tmpDir, "none",
				"grep -rn 'uses:.*@v[0-9]' .github/workflows/ 2>/dev/null || echo 'none'"),
			Zizmor: shIn(tmpDir, "zizmor not installed — skipped",
				"zizmor --format json .github/workflows/ 2>&1 || echo 'zizmor not installed — skipped'"),
			WorkflowList: shIn(tmpDir, "(none)",
				"ls .github/workflows/ 2>/dev/null || echo '(none)'"),
		},
		Code: auditCode{
			EvalUsage: shIn(tmpDir, "none",
				"grep -rn 'eval(' --include='*.ts' --include='*.js' . | grep -v node_modules | grep -v '\\.test\\.' | head -40 || echo 'none'"),
			MathRandom: shIn(tmpDir, "none",
				"grep -rn 'Math\\.random()' --include='*.ts' --include='*.js' . | grep -v node_modules | grep -v '\\.test\\.' | head -20 || echo 'none'"),
			RawSqlCalls: shIn(tmpDir, "none",
				"grep -rn '\\.raw(' --include='*.ts' . | grep -v node_modules | grep -v '\\.test\\.' | head -40 || echo 'none'"),
			XPoweredByHeader: shIn(tmpDir, "none",
				"grep -rn 'X-Powered-By\\|x-powered-by' --include='*.ts' --include='*.go' . | grep -v node_modules | head -20 || echo 'none'"),
			HardcodedSecretHints: shIn(tmpDir, "none",
				`grep -rEn "(password|secret|api_key)\s*=\s*[\"'][^\"']{4,}[\"']" --include='*.ts' --include='*.go' . | grep -v node_modules | grep -v '\.test\.' | head -20 || echo 'none'`),
			WeakCrypto: shIn(tmpDir, "none",
				"grep -rn 'createHash.*md5\\|createHash.*sha1\\|md5\\.New\\|sha1\\.New' --include='*.ts' --include='*.go' . | grep -v node_modules | head -20 || echo 'none'"),
			ProcessExitCalls: shIn(tmpDir, "none",
				"grep -rn 'process\\.exit\\|os\\.Exit' --include='*.ts' --include='*.go' . | grep -v node_modules | grep -v '\\.test\\.' | head -20 || echo 'none'"),
		},
		Infra: auditInfra{
			HelmLint: shIn(tmpDir, "helm not installed — skipped",
				"helm lint helm/*/ 2>&1 || echo 'no helm chart or helm not installed'"),
			HelmSecretTemplate: shIn(tmpDir, "(not found)",
				"find . -path '*/helm/*/templates/secret.yaml' | head -1 | xargs cat 2>/dev/null || echo '(not found)'"),
			HelmValues: shIn(tmpDir, "(not found)",
				"find . -path '*/helm/*/values.yaml' | head -1 | xargs cat 2>/dev/null || echo '(not found)'"),
			Dockerfile: shIn(tmpDir, "(not found)",
				"cat Dockerfile 2>/dev/null || echo '(not found)'"),
		},
		Dependencies: auditDeps{
			PnpmAudit: shIn(tmpDir, "pnpm not available — skipped",
				"pnpm audit --json 2>&1 | head -300 || npm audit --json 2>&1 | head -300 || echo 'no package manager available'"),
			WorkspaceOverrides: shIn(tmpDir, "none",
				"grep -A 40 '^overrides:' pnpm-workspace.yaml 2>/dev/null || echo 'none'"),
		},
		Git: auditGit{
			RecentCommits: shIn(tmpDir, "(unavailable)",
				"git log --oneline -30 2>/dev/null"),
			RecentlyChangedFiles: shIn(tmpDir, "(unavailable)",
				"git diff HEAD~10..HEAD --name-only 2>/dev/null | head -60 || echo '(unavailable)'"),
		},
	}

	ctx.GitHub = fetchGitHubContext(repo, ghToken)

	return ctx, tmpDir, nil
}

func fetchGitHubContext(repo, ghToken string) auditGitHub {
	fetch := func(url string) (interface{}, error) {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if ghToken != "" {
			req.Header.Set("Authorization", "Bearer "+ghToken)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close() //nolint:errcheck
		var v interface{}
		json.NewDecoder(resp.Body).Decode(&v) //nolint:errcheck
		return v, nil
	}

	base := "https://api.github.com/repos/" + repo

	issues, _ := fetch(base + "/issues?state=open&per_page=50")
	prs, _ := fetch(base + "/pulls?state=open&per_page=20")

	alerts := "(no token or insufficient permissions)"
	if ghToken != "" {
		req, _ := http.NewRequest("GET", base+"/secret-scanning/alerts?per_page=20", nil)
		req.Header.Set("Authorization", "Bearer "+ghToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		if resp, err := http.DefaultClient.Do(req); err == nil {
			defer resp.Body.Close() //nolint:errcheck
			body, _ := io.ReadAll(resp.Body)
			alerts = string(body)
		}
	}

	return auditGitHub{
		OpenIssues:     issues,
		OpenPRs:        prs,
		SecurityAlerts: alerts,
	}
}

// ── Prompts ───────────────────────────────────────────────────────────────────

const auditSystemPrompt = `You are a senior DevSecOps engineer producing a formal, peer-reviewable security audit report ` +
	`for a GitHub repository. Your output is a single, complete Markdown document with no surrounding commentary.

Non-negotiable principles:

1. ROOT CAUSE — explain WHY the issue exists, not just what it is.
2. IMPACT CHAIN — trace the realistic path from finding to harm.
3. CALIBRATED SEVERITY — Critical = RCE / auth bypass / privilege escalation / secret disclosure. ` +
	`An informational header exposure is Low, not High. Over-rating severity destroys reviewer trust.
4. METHODOLOGY TRANSPARENCY — for categories with no findings, write: ` +
	`"No <X> paths identified during static review. Method: <what grep pattern, which paths were searched>." ` +
	`Never write "X — LOW RISK" — auditors cannot prove a negative.
5. VERIFICATION COMMANDS — every actionable finding must include concrete, copy-paste shell commands ` +
	`(curl, grep, kubectl, helm) that a reviewer can run to confirm the fix.
6. SYNTHESISE, do not dump — a finding is: root cause + impact chain + fix + verification. ` +
	`Raw tool output belongs in a raw log, not in an audit report.
7. PRIORITY vs SEVERITY — P0/P1/P2 labels denote fix urgency, not CVSS severity bands. ` +
	`State CVSS severity (Critical / High / Medium / Low / Informational) separately per finding.
8. OPEN ISSUES & PRS — surface any security-relevant open GitHub issues or PRs as a dedicated section. ` +
	`Assess their risk and flag any that may introduce new vulnerabilities before merge.
9. SHIFT-LEFT — close with a table mapping each manual verification step to an automated CI guardrail.`

func buildUserPrompt(ctx *auditContext) string {
	dateShort := ctx.Meta.Date[:10]
	ctxJSON, _ := json.MarshalIndent(ctx, "", "  ")

	return fmt.Sprintf(`Generate a complete DevSecOps audit report for the scan results below.

Repository: %s
Commit: %s
Scan date: %s

## Collected security context

`+"```json\n%s\n```"+`

## Required document structure

Produce the following sections in order. Do not omit any.

1. **Metadata table** — date, repository, commit, auditor ("Automated — Claude Opus"), status
2. **Scope** — what was checked (files, tools, GitHub API calls)
3. **Methodology** — tools used, static vs dynamic distinction, known limitations
4. **Findings Summary** — table: ID | Priority | Severity | Title | OWASP 2021 | Status
5. **Per-finding sections** (one H3 per finding) — each must contain:
   - OWASP, CWE, Severity metadata
   - Description
   - Root Cause
   - Impact Chain
   - Fix (code or config snippet where applicable)
   - Verification (shell commands)
6. **Open GitHub Issues & PRs** — security-relevant items with risk assessment
7. **P2 Recommendations** — backlog items not immediately critical
8. **Remediation Status table** — all findings with commit or PR reference where fixed
9. **Verification Checklist** — numbered list of copy-paste commands, one per finding
10. **Shift-left guardrails** — table: Finding | Manual check | Automated CI gate
11. **Appendix: Full Application Security Assessment** — one H3 subsection per category `+
		`(SQL Injection, Authentication, Authorisation, Deserialization, SSRF, XXE, Path Traversal, `+
		`Cryptography, Rate Limiting, Dependencies, HTTP Headers, Container, Kubernetes/Helm). `+
		`Each subsection must start with the methodology note before listing observations.`,
		ctx.Meta.Repo, ctx.Meta.Ref, dateShort, string(ctxJSON))
}

// ── Claude API ────────────────────────────────────────────────────────────────

type claudeSystemBlock struct {
	Type         string              `json:"type"`
	Text         string              `json:"text"`
	CacheControl *claudeCacheControl `json:"cache_control,omitempty"`
}

type claudeCacheControl struct {
	Type string `json:"type"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    []claudeSystemBlock `json:"system"`
	Messages  []claudeMessage     `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens          int `json:"input_tokens"`
		CacheReadInputTokens int `json:"cache_read_input_tokens"`
		OutputTokens         int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func generateReport(ctx *auditContext, apiKey string) (report string, inputTokens, outputTokens int, err error) {
	payload := claudeRequest{
		Model:     "claude-opus-4-8",
		MaxTokens: 8192,
		System: []claudeSystemBlock{
			{
				Type:         "text",
				Text:         auditSystemPrompt,
				CacheControl: &claudeCacheControl{Type: "ephemeral"},
			},
		},
		Messages: []claudeMessage{
			{Role: "user", Content: buildUserPrompt(ctx)},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("claude API request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	var cr claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", 0, 0, fmt.Errorf("claude API decode failed: %w", err)
	}
	if cr.Error != nil {
		return "", 0, 0, fmt.Errorf("claude API error %s: %s", cr.Error.Type, cr.Error.Message)
	}
	if len(cr.Content) == 0 {
		return "", 0, 0, fmt.Errorf("claude API returned empty content")
	}

	return cr.Content[0].Text, cr.Usage.InputTokens, cr.Usage.OutputTokens, nil
}

// ── Runner ────────────────────────────────────────────────────────────────────

func runAudit(db *sql.DB, id, repo, ghToken, anthropicKey string) {
	_ = dbUpdateAuditRunning(db, id)

	auditCtx, tmpDir, err := collectContext(repo, ghToken)
	if err != nil {
		_ = dbUpdateAuditError(db, id, fmt.Sprintf("collect failed: %v", err))
		return
	}
	defer os.RemoveAll(tmpDir)

	report, inputTokens, outputTokens, err := generateReport(auditCtx, anthropicKey)
	if err != nil {
		_ = dbUpdateAuditError(db, id, fmt.Sprintf("generate failed: %v", err))
		return
	}

	_ = dbUpdateAuditDone(db, id, report, inputTokens, outputTokens)
}

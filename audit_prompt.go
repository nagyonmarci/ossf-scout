package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var reThinkBlock = regexp.MustCompile(`(?s)<think>.*?</think>\s*`)

// stripThinkBlocks removes <think>…</think> reasoning blocks that the model
// emits as part of severity calibration (per system prompt rule 13).
func stripThinkBlocks(s string) string {
	return strings.TrimSpace(reThinkBlock.ReplaceAllString(s, ""))
}

// ── Prompts ───────────────────────────────────────────────────────────────────

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
9. SHIFT-LEFT — close with a table: Finding | Manual check | Automated CI gate | CI YAML snippet. ` +
	`The CI YAML snippet column must contain a runnable GitHub Actions step (≤10 lines) that implements the gate.
10. NO FABRICATION (most important) — every file:line reference, commit SHA, action pin SHA, ` +
	`PR number (#NNN), and CVE you cite MUST appear verbatim in the Collected Context. ` +
	`For action pin SHAs, copy them from the "Resolved pin SHAs (AUTHORITATIVE)" block — never invent one. ` +
	`If a specific is not present in the context, write "(not captured)" instead. ` +
	`Inventing a SHA, line number, PR, or CVE is the single worst error you can make and destroys the report. ` +
	`This applies equally to historical CVEs cited as examples — if a CVE number is not in the collected evidence, do not cite it; describe the incident in general terms instead (e.g. "the 2023 tj-actions supply-chain compromise" not "CVE-2023-XXXXX").
11. CVSS — for every finding state a full CVSS v3.1 vector string (e.g. CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N) ` +
	`alongside the score, and use these exact severity bands: 0.0 None · 0.1–3.9 Low · 4.0–6.9 Medium · ` +
	`7.0–8.9 High · 9.0–10.0 Critical. A score of 8.7 is HIGH, not Critical. The score must match the vector.
12. SUPPLY-CHAIN SEVERITY RUBRIC — an unpinned GitHub Action is a hardening gap (OpenSSF Scorecard ` +
	`Pinned-Dependencies), NOT inherently Critical/P0. Rate it Medium by default. Rate it High only with a ` +
	`concrete escalation path: the workflow grants write permissions (contents/packages/id-token: write) or ` +
	`handles secrets, AND runs on an attacker-influenced trigger (pull_request_target, issue_comment, workflow_run, ` +
	`or pull_request from forks). A third-party action on a mutable tag in a read-only-token workflow triggered by ` +
	`push or ordinary pull_request is Medium. Reserve Critical/P0 for demonstrated RCE, auth bypass, privilege ` +
	`escalation, or secret disclosure — never for tag-pinning alone.
13. SEVERITY REASONING — before finalising the severity label for each finding, reason through it explicitly ` +
	`in a <think> block (stripped from final output by post-processing): ` +
	`<think>Evidence: [cite exact file:line or tool output]. Attacker preconditions: [what must be true]. ` +
	`Impact path: [step-by-step]. CVSS vector components: AV:[?] AC:[?] PR:[?] UI:[?] S:[?] C:[?] I:[?] A:[?]. ` +
	`Base score: [computed]. Band: [band]. Is this calibrated? [yes/no + why]</think> ` +
	`The think block ensures you do not skip the impact chain analysis.
14. ADMIN-BY-DESIGN — If a capability (e.g. a "Run Script" / eval-based executor, raw SQL admin panel, ` +
	`arbitrary command runner) is explicitly documented as admin-only, do NOT list it as a standalone security ` +
	`finding. The attack path requires defeating access control first — that is a separate, prior finding. ` +
	`Instead, document it in the Appendix as a P3 Architectural Risk note: ` +
	`"Exploitability depends on a prior RBAC bypass not observed in this audit." ` +
	`Elevate to P1/P2 only if evidence shows a non-admin code path directly reaches the dangerous function.
15. CONFIRMED vs SUSPECTED — A static grep match (fetch(), axios.get(), URL construction) identifies an ` +
	`attack surface, not a confirmed vulnerability. If you cannot trace the data flow from a user-controlled ` +
	`input to the dangerous call using evidence in the context, label the finding ` +
	`"Potential Attack Surface — Requires Confirmation", assign CVSS ≤ 5.0, and add: ` +
	`"Requires dynamic validation or manual code-path review to confirm exploitability." ` +
	`Do not assign CVSS 7.0+ to an unconfirmed pattern match.
16. RELIABILITY vs SECURITY — Workflow configuration errors that cause pipeline failure ` +
	`(invalid/unknown permission scopes, syntax errors, actionlint warnings with no exploitable path) ` +
	`without a realistic security impact path are NOT security findings. Do not include them in the ` +
	`Findings Summary table. Place them in a "CI/CD Reliability" H3 subsection inside the Appendix.
17. UNTRIAGED SECRETS — If a secrets scanner (gitleaks, trufflehog) returns only an aggregate count ` +
	`with no individual findings itemized in the collected evidence, do NOT assign a CVSS score. ` +
	`List it in the Findings Summary as: "[TOOL]: N aggregate findings — manual triage required", ` +
	`Priority P2, Severity "Unrated — pending triage". A finding cannot be rated without knowing what was found.

## Calibration examples (few-shot)

CORRECT rating — SQL injection with direct DB access:
<think>Evidence: src/db/query.go:142 raw string concat. Attacker: unauthenticated HTTP param. Impact: full DB read/write → exfiltrate users, insert admin. CVSS: AV:N AC:L PR:N UI:N S:U C:H I:H A:L → 9.4 Critical.</think>
→ **Finding CRITICAL-001 · SQL Injection · CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:L · 9.4 Critical**

CORRECT rating — unpinned action in a read-only workflow:
<think>Evidence: .github/workflows/ci.yml:12 uses actions/checkout@v4 (mutable tag). Attacker: supply chain compromise of that action. Precondition: action must be compromised, workflow has no write perms (no id-token/contents:write), triggered by push. Impact: code execution in CI but no secret exfil without extra env vars. CVSS: AV:N AC:H PR:N UI:R S:U C:L I:L A:N → 3.7 Low.</think>
→ **Finding LOW-001 · Unpinned Action (read-only context) · CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:L/I:L/A:N · 3.7 Low**

INCORRECT (over-rated — do not do this):
→ Finding CRITICAL-001 · Unpinned GitHub Action · 9.0 Critical ← WRONG: no write perms, no secret access`

func buildUserPrompt(ctx *auditContext) string {
	dateShort := ctx.Meta.Date[:10]
	ctxMD := buildContextMarkdown(ctx)

	return fmt.Sprintf(`Generate a complete DevSecOps audit report for the scan results below.

Repository: %s
Commit: %s
Scan date: %s

## Collected security context

`+"```\n%s\n```"+`

## Required document structure

Produce the following sections in order. Do not omit any.

1. **Metadata table** — date, repository, commit, auditor ("Automated — Claude Opus"), status
2. **Executive Summary** — 3–5 sentence non-technical overview for engineering managers and CISOs.
   Cover: one-sentence security posture verdict, the most severe finding in plain English,
   the overall Security Posture Score, and the single most important recommended action.
   Audience: non-technical stakeholders. Avoid jargon. Do not omit even if all findings are low severity.
3. **Scope** — what was checked (files, tools, GitHub API calls)
4. **Methodology** — tools used, static vs dynamic distinction, known limitations
5. **Security Strengths** — list security controls that are correctly implemented and provide genuine protection.
   Examples: CodeQL enabled, secret scanning active, multi-stage Docker build, IP blocklist present, Dependabot configured.
   For each item cite the evidence (file, workflow name, or tool output).
   This section demonstrates audit balance — do not omit it even if findings are severe.
6. **Findings Summary** — table: ID | Priority | Severity | Title | OWASP 2021 | Status
7. **Security Posture Summary** — a table with area scores (0–10) derived from the findings above.
   Scores must be grounded in evidence — do not fabricate.
   Columns: Area | Score | Rationale.
   Rows: CI/CD Pipeline Security · Dependency Management · Secrets Management · Supply Chain Integrity · Container Security · Application Code (SAST) · **Overall** (weighted average).
8. **Per-finding sections** (one H3 per finding) — each must contain:
   - OWASP, CWE, Severity metadata
   - Description
   - Root Cause
   - Impact Chain
   - Fix (code or config snippet where applicable)
   - Verification (shell commands)
9. **Open GitHub Issues & PRs** — security-relevant items with risk assessment. If the GitHub context marks issues/PRs as unavailable or rate-limited, state that plainly and DO NOT cite any specific issue or PR numbers.
10. **P2 Recommendations** — table: ID | Title | Effort (hours/days) | Risk Reduction (Low/Medium/High) | Notes
    Estimate effort as engineer-hours for a mid-level contributor. Risk reduction is the severity of the gap being closed.
11. **Remediation Roadmap** — a 30/60/90-day plan derived from the findings above.
    Table columns: Horizon | Finding IDs | Actions.
    Rows: 30 days (P1 findings + high-exploitability P2) · 60 days (remaining P2, CVE upgrades, CI hardening) · 90 days (P3, SBOM, provenance, supply-chain maturity).
    Base urgency on CVSS score and exploitability evidence — do not invent timelines.
12. **Sprint Backlog (P2 items)** — translate every P2 finding into a sprint-ready user story.
    Table columns: Story | Acceptance Criteria | Story Points | Sprint.
    Rules: one row per P2 finding; story format "As a security engineer, I want to [fix X] so that [risk Y is eliminated]";
    acceptance criteria = the Verification Checklist command for that finding;
    story points: 1–2 h → 1, 3–4 h → 2, 5–8 h → 3, 1–2 d → 5, 3–5 d → 8;
    sprint: Sprint 1 = highest CVSS or easiest fix, Sprint 3 = lowest CVSS or most complex.
    Only use P2 findings from the Findings Summary — do not invent items.
13. **Remediation Status table** — list all findings; cite a commit or PR reference ONLY if it appears in the collected git log / GitHub evidence, otherwise write "not yet fixed (no commit in scanned history)". Never invent a fix reference.
14. **Verification Checklist** — numbered list of copy-paste commands, one per finding
15. **Shift-left guardrails** — table: Finding | Manual check | Automated CI gate | CI YAML snippet
    The CI YAML snippet column must contain a runnable GitHub Actions step (≤10 lines) that implements the gate.
16. **Appendix: Full Application Security Assessment** — one H3 subsection per category `+
		`(SQL Injection, Authentication, Authorisation, Deserialization, SSRF, XXE, Path Traversal, `+
		`Cryptography, Rate Limiting, CORS, Dependencies, HTTP Headers, Container, Kubernetes/Helm, `+
		`Secrets / Credential Hygiene, IaC Security, Policy as Code, SLSA / Supply Chain, CI/CD Reliability). `+
		`Each subsection must start with the methodology note before listing observations. `+
		`For SQL Injection: use keyFiles.entryPoint and code.sqlInjection; ORM raw call with string concatenation = High. `+
		`For Authentication: read keyFiles.authMiddleware carefully; missing jwt.verify or session fixation = Critical. `+
		`For Authorisation: read keyFiles.permissionSystem; route handler without permission check = High. `+
		`For SSRF: evaluate code.ssrf against keyFiles; user-controlled URL passed to fetch/axios/got = Critical. `+
		`For Rate Limiting: if code.rateLimiting is empty or 'none', flag missing rate limiting as Medium/High on auth endpoints. `+
		`For CORS: if code.corsConfig shows origin:'*' with credentials, flag as Medium. `+
		`For Secrets: assess Gitleaks output or grep results, estimate false positive rate, list any confirmed leaks as Critical findings. `+
		`For IaC: synthesise OSV-Scanner and Trivy findings, flag unencrypted storage / overly-permissive IAM / missing network policies. `+
		`For Policy as Code: assess OPA/Kyverno coverage gaps — what admission controls are missing. `+
		`For SLSA / Supply Chain: determine the current SLSA level (L0–L3) based on the evidence, `+
		`state exactly what is needed to reach the next level, and flag any unsigned artifacts or missing provenance. `+
		`For CI/CD gate enforcement: check cicd.workflowContents for each security workflow (codeql, trivy, scorecard). `+
		`If any lacks an 'on: pull_request' trigger, flag as High — the scan is scheduled-only and does not block merges. `+
		`For startup validation: check keyFiles.startupValidation for advisory-only (logger.warn without process.exit) `+
		`enforcement of critical config (SECRET, tokens, signing keys). If production can start with insecure config `+
		`without hard-failing, flag as High — misconfiguration silently reaches production. `+
		`For error handling: check keyFiles.errorHandler for stack trace or internal error details in production responses `+
		`(res.json(err) / err.stack exposure without NODE_ENV guard). Flag as Medium. `+
		`For HTTP security headers: use keyFiles.helmetConfig to assess CSP directives, HSTS enforcement, and `+
		`X-Powered-By suppression with full context — distinguish deliberate upstream defaults from oversights. `+
		`For CI/CD Reliability: place workflow configuration errors (invalid permission scopes, syntax/lint issues) `+
		`that have no exploitable security impact here — NOT in the Findings Summary. `+
		`17. **Threat Model (STRIDE)** — table: Threat | STRIDE category `+
		`(Spoofing / Tampering / Repudiation / Information Disclosure / Denial of Service / Elevation of Privilege) | `+
		`Affected component | Existing mitigation | Residual risk (High/Med/Low). `+
		`Cover at least: authentication flows, CI/CD pipeline, supply-chain, secrets storage, API inputs.`,
		ctx.Meta.Repo, ctx.Meta.Ref, dateShort, ctxMD)
}

func buildSummarizedUserPrompt(ctx *auditContext, summaries []auditSectionSummary) string {
	dateShort := ctx.Meta.Date[:10]

	var sb strings.Builder
	for _, s := range summaries {
		fmt.Fprintf(&sb, "### %s\n\n%s\n\n---\n\n", s.Section, s.Summary)
	}
	summaryMD := sb.String()

	return fmt.Sprintf(`Generate a complete DevSecOps audit report for the scan results below.

Repository: %s
Commit: %s
Scan date: %s

## Collected security context

The raw audit context has already been reduced into focused evidence summaries. Treat these summaries as the source of truth. Preserve concrete file paths, commands, tool names, and finding evidence. Do not invent findings not supported by the evidence.

%s

## Required document structure

Produce the following sections in order. Do not omit any.

1. **Metadata table** — date, repository, commit, auditor ("Automated — split model workflow"), status
2. **Executive Summary** — 3–5 sentence non-technical overview for engineering managers and CISOs.
   Cover: one-sentence security posture verdict, the most severe finding in plain English, the overall Security Posture Score, and the single most important recommended action.
   Audience: non-technical stakeholders. Avoid jargon. Do not omit even if all findings are low severity.
3. **Scope** — what was checked (files, tools, GitHub API calls)
4. **Methodology** — tools used, static vs dynamic distinction, known limitations
5. **Security Strengths** — list security controls that are correctly implemented and provide genuine protection.
   For each item cite the evidence. Do not omit this section even if findings are severe.
6. **Findings Summary** — table: ID | Priority | Severity | Title | OWASP 2021 | Status
7. **Security Posture Summary** — area scores (0–10) table: Area | Score | Rationale.
   Rows: CI/CD Pipeline Security · Dependency Management · Secrets Management · Supply Chain Integrity · Container Security · Application Code (SAST) · Overall (weighted average).
   Scores must be grounded in evidence — do not fabricate.
8. **Per-finding sections** (one H3 per finding) — each must contain:
   - OWASP, CWE, Severity metadata
   - Description
   - Root Cause
   - Impact Chain
   - Fix (code or config snippet where applicable)
   - Verification (shell commands)
9. **Open GitHub Issues & PRs** — security-relevant items with risk assessment. If the GitHub context marks issues/PRs as unavailable or rate-limited, state that plainly and DO NOT cite any specific issue or PR numbers.
10. **P2 Recommendations** — table: ID | Title | Effort (hours/days) | Risk Reduction (Low/Medium/High) | Notes
    Estimate effort as engineer-hours for a mid-level contributor. Risk reduction is the severity of the gap being closed.
11. **Remediation Roadmap** — 30/60/90-day plan. Table: Horizon | Finding IDs | Actions.
    30 days: P1 + high-exploitability P2. 60 days: remaining P2, CVE upgrades, CI hardening. 90 days: P3, SBOM, provenance.
    Base urgency on CVSS score and exploitability evidence.
12. **Sprint Backlog (P2 items)** — translate every P2 finding into a sprint-ready user story.
    Table columns: Story | Acceptance Criteria | Story Points | Sprint.
    Story format: "As a security engineer, I want to [fix X] so that [risk Y is eliminated]".
    Story points: 1–2 h → 1, 3–4 h → 2, 5–8 h → 3, 1–2 d → 5, 3–5 d → 8.
    Sprint: Sprint 1 = highest CVSS or easiest fix, Sprint 3 = lowest CVSS or most complex.
    Only use P2 findings from the Findings Summary — do not invent items.
13. **Remediation Status table** — list all findings; cite a commit or PR reference ONLY if it appears in the collected git log / GitHub evidence, otherwise write "not yet fixed (no commit in scanned history)". Never invent a fix reference.
14. **Verification Checklist** — numbered list of copy-paste commands, one per finding
15. **Shift-left guardrails** — table: Finding | Manual check | Automated CI gate | CI YAML snippet
    The CI YAML snippet column must contain a runnable GitHub Actions step (≤10 lines) that implements the gate.
16. **Appendix: Full Application Security Assessment** — one H3 subsection per category.
Each appendix subsection must start with the methodology note before listing observations.
Include a "CI/CD Reliability" subsection for workflow config errors with no exploitable security path.
17. **Threat Model (STRIDE)** — table: Threat | STRIDE category (Spoofing / Tampering / Repudiation / Information Disclosure / Denial of Service / Elevation of Privilege) | Affected component | Existing mitigation | Residual risk (High/Med/Low).
Cover at least: authentication flows, CI/CD pipeline, supply-chain, secrets storage, API inputs.`,
		ctx.Meta.Repo, ctx.Meta.Ref, dateShort, summaryMD)
}

// ── Claude API ────────────────────────────────────────────────────────────────

// ── Shared summary types and utilities ──────────────────────────────────────

type auditSectionSummary struct {
	Section string `json:"section"`
	Model   string `json:"model"`
	Summary string `json:"summary"`
}

// truncateField caps s at maxBytes and appends a note if truncated.
func truncateField(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "\n... [truncated]"
}

const auditSummarySystemPrompt = `You are a careful DevSecOps evidence summarizer. ` +
	`Your job is to reduce one section of raw audit data into compact, accurate evidence for another model. ` +
	`Never invent findings. Preserve concrete evidence and uncertainty.`

type auditSummarySection struct {
	Name    string
	Content string // Markdown
}

func sectionMD(parts ...func(w func(string, ...any))) string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	for _, part := range parts {
		part(w)
	}
	return b.String()
}

func buildAuditSummarySections(ctx *auditContext) []auditSummarySection {
	return []auditSummarySection{
		{Name: "CI/CD and policy gates", Content: sectionMD(func(w func(string, ...any)) {
			w("## CI/CD")
			w("### Unpinned GitHub Actions")
			w("```\n%s\n```", ctx.CICD.UnpinnedActions)
			w("### Actionlint")
			w("```\n%s\n```", ctx.CICD.Actionlint)
			w("### Workflow files")
			w("```\n%s\n```", ctx.CICD.WorkflowList)
			w("### Zizmor findings")
			w("%s", extractZizmortFindings(ctx.CICD.Zizmor))
			w("### Security workflow contents")
			w("```yaml\n%s\n```", truncateField(ctx.CICD.WorkflowContents, truncSummaryWorkflow))
		})},
		{Name: "Application code and key files", Content: sectionMD(func(w func(string, ...any)) {
			w("## Code Patterns")
			w("### eval()\n```\n%s\n```", ctx.Code.EvalUsage)
			w("### Math.random()\n```\n%s\n```", ctx.Code.MathRandom)
			w("### Raw SQL\n```\n%s\n```", ctx.Code.RawSqlCalls)
			w("### X-Powered-By\n```\n%s\n```", ctx.Code.XPoweredByHeader)
			w("### Hardcoded secrets\n```\n%s\n```", ctx.Code.HardcodedSecretHints)
			w("### Weak crypto\n```\n%s\n```", ctx.Code.WeakCrypto)
			w("### process.exit/os.Exit\n```\n%s\n```", ctx.Code.ProcessExitCalls)
			w("### SQL injection\n```\n%s\n```", ctx.Code.SqlInjection)
			w("### SSRF\n```\n%s\n```", ctx.Code.SSRF)
			w("### Path traversal\n```\n%s\n```", ctx.Code.PathTraversal)
			w("### XXE\n```\n%s\n```", ctx.Code.XXE)
			w("### Deserialization\n```\n%s\n```", ctx.Code.Deserialization)
			w("### Rate limiting\n```\n%s\n```", ctx.Code.RateLimiting)
			w("### CORS\n```\n%s\n```", ctx.Code.CORSConfig)
			w("### Semgrep findings\n```json\n%s\n```", truncateField(ctx.Code.SemgrepFindings, truncSemgrep))
			w("## Key Security Files")
			w("### Entry point\n```\n%s\n```", ctx.KeyFiles.EntryPoint)
			w("### Auth middleware\n```\n%s\n```", ctx.KeyFiles.AuthMiddleware)
			w("### Permission system\n```\n%s\n```", ctx.KeyFiles.PermissionSystem)
			w("### Security config\n```\n%s\n```", ctx.KeyFiles.SecurityConfig)
			w("### Startup validation\n```\n%s\n```", ctx.KeyFiles.StartupValidation)
			w("### Error handler\n```\n%s\n```", ctx.KeyFiles.ErrorHandler)
			w("### Helmet config\n```\n%s\n```", ctx.KeyFiles.HelmetConfig)
		})},
		{Name: "Dependencies and secrets", Content: sectionMD(func(w func(string, ...any)) {
			w("## Dependencies")
			w("### Dependency audit (pnpm/npm/yarn)\n```\n%s\n```", ctx.Dependencies.PnpmAudit)
			w("### Workspace overrides\n```\n%s\n```", ctx.Dependencies.WorkspaceOverrides)
			w("## Secrets Scanning")
			w("### Gitleaks\n```\n%s\n```", ctx.Secrets.Gitleaks)
			w("### TruffleHog\n```\n%s\n```", ctx.Secrets.TruffleHog)
			w("### Private key headers\n```\n%s\n```", ctx.Secrets.PrivateKeyHeaders)
			w("### .env files\n```\n%s\n```", ctx.Secrets.EnvFiles)
			w("### Token patterns\n```\n%s\n```", truncateField(ctx.Secrets.TokenPatterns, truncSummaryTokenPat))
		})},
		{Name: "Infrastructure, containers, Kubernetes, and SLSA", Content: sectionMD(func(w func(string, ...any)) {
			w("## Infrastructure")
			w("### Dockerfile\n```dockerfile\n%s\n```", ctx.Infra.Dockerfile)
			w("### Helm lint\n```\n%s\n```", ctx.Infra.HelmLint)
			w("### Helm secret template\n```yaml\n%s\n```", ctx.Infra.HelmSecretTemplate)
			w("### Helm values\n```yaml\n%s\n```", ctx.Infra.HelmValues)
			w("## IaC")
			w("### Terraform files\n```\n%s\n```", ctx.IaC.TerraformFiles)
			w("### OSV-Scanner\n```\n%s\n```", ctx.IaC.OSVScanner)
			w("### Trivy config\n```\n%s\n```", ctx.IaC.Trivy)
			w("### Kubernetes manifests\n```\n%s\n```", ctx.IaC.KubeManifests)
			w("### kube-linter\n```\n%s\n```", ctx.IaC.KubeLinter)
			w("## Policy as Code")
			w("### OPA\n```\n%s\n```", ctx.Policy.OPAFiles)
			w("### Kyverno\n```\n%s\n```", ctx.Policy.KyvernoFiles)
			w("### Falco\n```\n%s\n```", ctx.Policy.FalcoRules)
			w("## SLSA / Supply Chain")
			w("### Provenance files\n```\n%s\n```", ctx.SLSA.ProvenanceFiles)
			w("### SBOM files\n```\n%s\n```", ctx.SLSA.SBOMFiles)
			w("### Cosign keys\n```\n%s\n```", ctx.SLSA.CosignFiles)
			w("### SLSA workflow\n```\n%s\n```", ctx.SLSA.SLSAWorkflow)
			w("### Signed commit\n```\n%s\n```", ctx.SLSA.SignedCommit)
		})},
		{Name: "Git and GitHub signals", Content: sectionMD(func(w func(string, ...any)) {
			w("## Git History")
			w("### Recent commits\n```\n%s\n```", ctx.Git.RecentCommits)
			w("### Recently changed files\n```\n%s\n```", ctx.Git.RecentlyChangedFiles)
			w("## GitHub")
			issuesJSON, _ := json.MarshalIndent(ctx.GitHub.OpenIssues, "", "  ")
			prsJSON, _ := json.MarshalIndent(ctx.GitHub.OpenPRs, "", "  ")
			if ghListAvailable(ctx.GitHub.IssuesStatus, ctx.GitHub.OpenIssues) {
				w("### Open issues\n```json\n%s\n```", truncateField(string(issuesJSON), truncSummaryIssues))
			} else {
				w("### Open issues\n_Data %s — do NOT cite specific issue numbers._", ghStatusLabel(ctx.GitHub.IssuesStatus))
			}
			if ghListAvailable(ctx.GitHub.PRsStatus, ctx.GitHub.OpenPRs) {
				w("### Open PRs\n```json\n%s\n```", truncateField(string(prsJSON), truncSummaryPRs))
			} else {
				w("### Open PRs\n_Data %s — do NOT cite specific PR numbers._", ghStatusLabel(ctx.GitHub.PRsStatus))
			}
			w("### Secret-scanning alerts\n```\n%s\n```", ctx.GitHub.SecurityAlerts)
			w("### Branch protection\n```json\n%s\n```", ctx.GitHub.BranchProtection)
			w("### Dependabot alerts\n```json\n%s\n```", truncateField(ctx.GitHub.DependabotAlerts, 4000))
			w("### Release history (last 10)\n```json\n%s\n```", truncateField(ctx.GitHub.ReleaseHistory, 2000))
			w("### CODEOWNERS\n```\n%s\n```", ctx.KeyFiles.CodeOwners)
		})},
	}
}


// ── Context markdown and SARIF extraction ───────────────────────────────────

func extractZizmortFindings(sarifJSON string) string {
	var sarif struct {
		Runs []struct {
			Results []struct {
				RuleID    string `json:"ruleId"`
				Level     string `json:"level"`
				Message   struct{ Text string `json:"text"` } `json:"message"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct{ URI string `json:"uri"` } `json:"artifactLocation"`
						Region           struct{ StartLine int `json:"startLine"` } `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal([]byte(sarifJSON), &sarif); err != nil || len(sarif.Runs) == 0 {
		return truncateField(sarifJSON, truncSarif)
	}
	results := sarif.Runs[0].Results
	if len(results) == 0 {
		return "(no findings)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "| Rule ID | Level | File | Line | Message |\n")
	fmt.Fprintf(&b, "|---------|-------|------|------|---------|\n")
	limit := 100
	if len(results) < limit {
		limit = len(results)
	}
	for _, r := range results[:limit] {
		file, line := "", 0
		if len(r.Locations) > 0 {
			file = r.Locations[0].PhysicalLocation.ArtifactLocation.URI
			line = r.Locations[0].PhysicalLocation.Region.StartLine
		}
		msg := strings.ReplaceAll(r.Message.Text, "\n", " ")
		if len(msg) > 120 {
			msg = msg[:120] + "…"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %d | %s |\n", r.RuleID, r.Level, file, line, msg)
	}
	if len(sarif.Runs[0].Results) > 100 {
		fmt.Fprintf(&b, "\n... and %d more findings (truncated at 100)\n", len(sarif.Runs[0].Results)-100)
	}
	return b.String()
}

// buildContextMarkdown converts an auditContext into a compact, human-readable Markdown
// string for use as AI prompt context. Unlike generateTemplateReport (which is a verbose
// static snapshot), this applies minimal truncation only where needed and replaces the
// raw Zizmor SARIF JSON with an extracted findings table.
func buildContextMarkdown(ctx *auditContext) string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }

	w("## CI/CD")
	w("")
	w("### Unpinned GitHub Actions")
	w("```")
	w("%s", ctx.CICD.UnpinnedActions)
	w("```")
	w("")
	w("### Resolved pin SHAs (AUTHORITATIVE — cite these exact SHAs in fixes; do not invent)")
	w("```")
	w("%s", ctx.CICD.PinnedSuggestions)
	w("```")
	w("")
	w("### Actionlint")
	w("```")
	w("%s", ctx.CICD.Actionlint)
	w("```")
	w("")
	w("### Workflow files")
	w("```")
	w("%s", ctx.CICD.WorkflowList)
	w("```")
	w("")
	w("### Zizmor findings")
	w("")
	w("%s", extractZizmortFindings(ctx.CICD.Zizmor))
	w("")
	w("### Security workflow contents (codeql / trivy / scorecard triggers)")
	w("```yaml")
	w("%s", truncateField(ctx.CICD.WorkflowContents, truncSummaryWorkflow))
	w("```")
	w("")
	w("## Code Patterns")
	w("")
	w("### eval()")
	w("```")
	w("%s", ctx.Code.EvalUsage)
	w("```")
	w("### Math.random()")
	w("```")
	w("%s", ctx.Code.MathRandom)
	w("```")
	w("### Raw SQL")
	w("```")
	w("%s", ctx.Code.RawSqlCalls)
	w("```")
	w("### X-Powered-By")
	w("```")
	w("%s", ctx.Code.XPoweredByHeader)
	w("```")
	w("### Hardcoded secrets")
	w("```")
	w("%s", ctx.Code.HardcodedSecretHints)
	w("```")
	w("### Weak crypto (MD5/SHA1)")
	w("```")
	w("%s", ctx.Code.WeakCrypto)
	w("```")
	w("### process.exit / os.Exit")
	w("```")
	w("%s", ctx.Code.ProcessExitCalls)
	w("```")
	w("### SQL injection")
	w("```")
	w("%s", ctx.Code.SqlInjection)
	w("```")
	w("### SSRF")
	w("```")
	w("%s", ctx.Code.SSRF)
	w("```")
	w("### Path traversal")
	w("```")
	w("%s", ctx.Code.PathTraversal)
	w("```")
	w("### XXE")
	w("```")
	w("%s", ctx.Code.XXE)
	w("```")
	w("### Deserialization")
	w("```")
	w("%s", ctx.Code.Deserialization)
	w("```")
	w("### Rate limiting")
	w("```")
	w("%s", ctx.Code.RateLimiting)
	w("```")
	w("### CORS config")
	w("```")
	w("%s", ctx.Code.CORSConfig)
	w("```")
	w("### Semgrep findings")
	w("```json")
	w("%s", truncateField(ctx.Code.SemgrepFindings, truncSemgrep))
	w("```")
	w("")
	w("## Key Security Files")
	w("")
	w("### Entry point")
	w("```")
	w("%s", ctx.KeyFiles.EntryPoint)
	w("```")
	w("### Auth middleware")
	w("```")
	w("%s", ctx.KeyFiles.AuthMiddleware)
	w("```")
	w("### Permission system")
	w("```")
	w("%s", ctx.KeyFiles.PermissionSystem)
	w("```")
	w("### Security config")
	w("```")
	w("%s", ctx.KeyFiles.SecurityConfig)
	w("```")
	w("### Startup validation")
	w("```")
	w("%s", ctx.KeyFiles.StartupValidation)
	w("```")
	w("### Error handler")
	w("```")
	w("%s", ctx.KeyFiles.ErrorHandler)
	w("```")
	w("### Helmet config")
	w("```")
	w("%s", ctx.KeyFiles.HelmetConfig)
	w("```")
	w("### CODEOWNERS")
	w("```")
	w("%s", ctx.KeyFiles.CodeOwners)
	w("```")
	w("")
	w("## Infrastructure")
	w("")
	w("### Dockerfile")
	w("```dockerfile")
	w("%s", ctx.Infra.Dockerfile)
	w("```")
	w("### Helm lint")
	w("```")
	w("%s", ctx.Infra.HelmLint)
	w("```")
	w("### Helm secret template")
	w("```yaml")
	w("%s", ctx.Infra.HelmSecretTemplate)
	w("```")
	w("### Helm values")
	w("```yaml")
	w("%s", ctx.Infra.HelmValues)
	w("```")
	w("")
	w("## Dependencies")
	w("")
	w("### Dependency audit (pnpm/npm/yarn)")
	w("```")
	w("%s", ctx.Dependencies.PnpmAudit)
	w("```")
	w("### Workspace overrides")
	w("```")
	w("%s", ctx.Dependencies.WorkspaceOverrides)
	w("```")
	w("")
	w("## Git History")
	w("")
	w("### Recent commits")
	w("```")
	w("%s", ctx.Git.RecentCommits)
	w("```")
	w("### Recently changed files")
	w("```")
	w("%s", ctx.Git.RecentlyChangedFiles)
	w("```")
	w("")
	w("## GitHub")
	w("")
	issuesJSON, _ := json.MarshalIndent(ctx.GitHub.OpenIssues, "", "  ")
	prsJSON, _ := json.MarshalIndent(ctx.GitHub.OpenPRs, "", "  ")
	w("### Open issues")
	if ghListAvailable(ctx.GitHub.IssuesStatus, ctx.GitHub.OpenIssues) {
		w("```json")
		w("%s", truncateField(string(issuesJSON), truncSummaryIssues))
		w("```")
	} else {
		w("_Data %s — do NOT cite specific issue numbers._", ghStatusLabel(ctx.GitHub.IssuesStatus))
	}
	w("### Open PRs")
	if ghListAvailable(ctx.GitHub.PRsStatus, ctx.GitHub.OpenPRs) {
		w("```json")
		w("%s", truncateField(string(prsJSON), truncSummaryPRs))
		w("```")
	} else {
		w("_Data %s — do NOT cite specific PR numbers._", ghStatusLabel(ctx.GitHub.PRsStatus))
	}
	w("### Secret-scanning alerts")
	w("```")
	w("%s", ctx.GitHub.SecurityAlerts)
	w("```")
	w("### Branch protection (main)")
	w("```json")
	w("%s", ctx.GitHub.BranchProtection)
	w("```")
	w("### Dependabot alerts")
	w("```json")
	w("%s", truncateField(ctx.GitHub.DependabotAlerts, 4000))
	w("```")
	w("### Release history (last 10)")
	w("```json")
	w("%s", truncateField(ctx.GitHub.ReleaseHistory, 2000))
	w("```")
	w("")
	w("## Secrets Scanning")
	w("")
	w("### Gitleaks")
	w("```")
	w("%s", ctx.Secrets.Gitleaks)
	w("```")
	w("### TruffleHog")
	w("```")
	w("%s", ctx.Secrets.TruffleHog)
	w("```")
	w("### Private key headers")
	w("```")
	w("%s", ctx.Secrets.PrivateKeyHeaders)
	w("```")
	w("### .env files")
	w("```")
	w("%s", ctx.Secrets.EnvFiles)
	w("```")
	w("### Token patterns (AWS/JWT/GH)")
	w("```")
	w("%s", ctx.Secrets.TokenPatterns)
	w("```")
	w("")
	w("## Infrastructure as Code (IaC)")
	w("")
	w("### Terraform files")
	w("```")
	w("%s", ctx.IaC.TerraformFiles)
	w("```")
	w("### OSV-Scanner")
	w("```")
	w("%s", ctx.IaC.OSVScanner)
	w("```")
	w("### Trivy config")
	w("```")
	w("%s", ctx.IaC.Trivy)
	w("```")
	w("### Kubernetes manifests")
	w("```")
	w("%s", ctx.IaC.KubeManifests)
	w("```")
	w("### kube-linter")
	w("```")
	w("%s", ctx.IaC.KubeLinter)
	w("```")
	w("")
	w("## Policy as Code")
	w("")
	w("### OPA (.rego files)")
	w("```")
	w("%s", ctx.Policy.OPAFiles)
	w("```")
	w("### Kyverno policies")
	w("```")
	w("%s", ctx.Policy.KyvernoFiles)
	w("```")
	w("### Falco rules")
	w("```")
	w("%s", ctx.Policy.FalcoRules)
	w("```")
	w("")
	w("## SLSA / Supply Chain")
	w("")
	w("### Provenance files")
	w("```")
	w("%s", ctx.SLSA.ProvenanceFiles)
	w("```")
	w("### SBOM files")
	w("```")
	w("%s", ctx.SLSA.SBOMFiles)
	w("```")
	w("### Cosign / signing keys")
	w("```")
	w("%s", ctx.SLSA.CosignFiles)
	w("```")
	w("### SLSA / attestation workflow usage")
	w("```")
	w("%s", ctx.SLSA.SLSAWorkflow)
	w("```")
	w("### Signed commit (latest)")
	w("```")
	w("%s", ctx.SLSA.SignedCommit)
	w("```")

	return b.String()
}

// generateTemplateReport formats the collected context as a structured Markdown
// snapshot without any AI synthesis. Used when no Anthropic API key is available.

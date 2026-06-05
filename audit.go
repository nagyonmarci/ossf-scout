package main

import (
	"bufio"
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
	UnpinnedActions  string `json:"unpinnedActions"`
	Zizmor           string `json:"zizmor"`
	Actionlint       string `json:"actionlint"`
	WorkflowList     string `json:"workflowList"`
	WorkflowContents string `json:"workflowContents"`
}

type auditCode struct {
	EvalUsage            string `json:"evalUsage"`
	MathRandom           string `json:"mathRandom"`
	RawSqlCalls          string `json:"rawSqlCalls"`
	XPoweredByHeader     string `json:"xPoweredByHeader"`
	HardcodedSecretHints string `json:"hardcodedSecretHints"`
	WeakCrypto           string `json:"weakCrypto"`
	ProcessExitCalls     string `json:"processExitCalls"`
	SqlInjection         string `json:"sqlInjection"`
	SSRF                 string `json:"ssrf"`
	PathTraversal        string `json:"pathTraversal"`
	XXE                  string `json:"xxe"`
	Deserialization      string `json:"deserialization"`
	RateLimiting         string `json:"rateLimiting"`
	CORSConfig           string `json:"corsConfig"`
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
	OpenIssues       interface{} `json:"openIssues"`
	OpenPRs          interface{} `json:"openPRs"`
	SecurityAlerts   string      `json:"securityAlerts"`
	BranchProtection string      `json:"branchProtection"`
}

type auditSecrets struct {
	Gitleaks          string `json:"gitleaks"`
	TruffleHog        string `json:"truffleHog"`
	PrivateKeyHeaders string `json:"privateKeyHeaders"`
	EnvFiles          string `json:"envFiles"`
	TokenPatterns     string `json:"tokenPatterns"`
}

type auditIaC struct {
	TerraformFiles string `json:"terraformFiles"`
	Checkov        string `json:"checkov"`
	Trivy          string `json:"trivy"`
	KubeManifests  string `json:"kubeManifests"`
	KubeLinter     string `json:"kubeLinter"`
}

type auditKeyFiles struct {
	EntryPoint        string `json:"entryPoint"`
	AuthMiddleware    string `json:"authMiddleware"`
	PermissionSystem  string `json:"permissionSystem"`
	SecurityConfig    string `json:"securityConfig"`
	StartupValidation string `json:"startupValidation"`
	ErrorHandler      string `json:"errorHandler"`
	HelmetConfig      string `json:"helmetConfig"`
}

type auditPolicy struct {
	OPAFiles     string `json:"opaFiles"`
	KyvernoFiles string `json:"kyvernoFiles"`
	FalcoRules   string `json:"falcoRules"`
}

type auditSLSA struct {
	ProvenanceFiles string `json:"provenanceFiles"`
	SBOMFiles       string `json:"sbomFiles"`
	CosignFiles     string `json:"cosignFiles"`
	SLSAWorkflow    string `json:"slsaWorkflow"`
	SignedCommit    string `json:"signedCommit"`
}

type auditContext struct {
	Meta         auditMeta     `json:"meta"`
	CICD         auditCICD     `json:"cicd"`
	Code         auditCode     `json:"code"`
	KeyFiles     auditKeyFiles `json:"keyFiles"`
	Infra        auditInfra    `json:"infra"`
	Dependencies auditDeps     `json:"dependencies"`
	Git          auditGit      `json:"git"`
	GitHub       auditGitHub   `json:"github"`
	Secrets      auditSecrets  `json:"secrets"`
	IaC          auditIaC      `json:"iac"`
	Policy       auditPolicy   `json:"policy"`
	SLSA         auditSLSA     `json:"slsa"`
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
	if ghToken != "" {
		cloneURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", ghToken, repo)
	}
	cloneCmd := exec.Command("git", "clone", "--depth=50", cloneURL, tmpDir)
	cloneCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir) //nolint:errcheck
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, "", fmt.Errorf("git clone failed: %s", msg)
	}

	ref := shIn(tmpDir, "unknown", "git rev-parse --short HEAD")

	zizmorCmd := "zizmor --no-online --format sarif .github/workflows/ 2>&1 || echo 'zizmor not installed — skipped'"
	if ghToken != "" {
		zizmorCmd = fmt.Sprintf("zizmor --github-token %s --format sarif .github/workflows/ 2>&1 || echo 'zizmor not installed — skipped'", ghToken)
	}

	ctx := &auditContext{
		Meta: auditMeta{
			Date: time.Now().UTC().Format(time.RFC3339),
			Repo: repo,
			Ref:  ref,
		},
		CICD: auditCICD{
			UnpinnedActions: shIn(tmpDir, "none",
				"grep -rn 'uses:.*@v[0-9]' .github/workflows/ 2>/dev/null || echo 'none'"),
			Zizmor: shIn(tmpDir, "zizmor not installed — skipped", zizmorCmd),
			Actionlint: shIn(tmpDir, "actionlint not installed — skipped",
				"actionlint -format '{{range $e := .}}{{$e.Filepath}}:{{$e.Line}}: [{{$e.Kind}}] {{$e.Message}}\n{{end}}' .github/workflows/*.yml 2>&1 | head -100 || echo 'actionlint not installed — skipped'"),
			WorkflowList: shIn(tmpDir, "(none)",
				"ls .github/workflows/ 2>/dev/null || echo '(none)'"),
			WorkflowContents: shIn(tmpDir, "(none)",
				"for f in $(find .github/workflows/ -name '*.yml' | xargs grep -l 'codeql\\|trivy\\|scorecard\\|dependency' 2>/dev/null | head -5); do echo \"=== $f ===\"; cat \"$f\"; echo; done 2>/dev/null || echo '(none)'"),
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
			SqlInjection: shIn(tmpDir, "none",
				"grep -rn 'knex\\.raw\\|whereRaw\\|sequelize\\.query\\|\\.db\\.query' --include='*.ts' . | grep -v node_modules | grep -v '\\.test\\.' | head -40 || echo 'none'"),
			SSRF: shIn(tmpDir, "none",
				"grep -rn 'fetch(\\|axios\\.get\\|axios\\.post\\|got(\\|http\\.get\\|https\\.get' --include='*.ts' --include='*.go' . | grep -v node_modules | grep -v '\\.test\\.' | head -40 || echo 'none'"),
			PathTraversal: shIn(tmpDir, "none",
				"grep -rn 'readFile\\|readFileSync\\|createReadStream\\|path\\.join' --include='*.ts' . | grep -v node_modules | grep -v '\\.test\\.' | head -30 || echo 'none'"),
			XXE: shIn(tmpDir, "none",
				"grep -rn 'xml2js\\|fast-xml-parser\\|DOMParser\\|XMLParser\\|parseString' --include='*.ts' . | grep -v node_modules | head -20 || echo 'none'"),
			Deserialization: shIn(tmpDir, "none",
				"grep -rn 'yaml\\.load\\|yaml\\.safeLoad\\|unserialize\\|JSON\\.parse(' --include='*.ts' . | grep -v node_modules | grep -v '\\.test\\.' | head -20 || echo 'none'"),
			RateLimiting: shIn(tmpDir, "none",
				"grep -rn 'rateLimit\\|express-rate-limit\\|rate_limit\\|rateLimiter' --include='*.ts' . | grep -v node_modules | head -20 || echo 'none'"),
			CORSConfig: shIn(tmpDir, "none",
				"grep -rn 'cors(' --include='*.ts' . | grep -v node_modules | head -10 || echo 'none'"),
		},
		KeyFiles: auditKeyFiles{
			EntryPoint: shIn(tmpDir, "(not found)",
				"for f in app.ts server.ts index.ts main.ts src/app.ts src/server.ts src/index.ts src/main.ts; do [ -f \"$f\" ] && head -150 \"$f\" && break; done 2>/dev/null || echo '(not found)'"),
			AuthMiddleware: shIn(tmpDir, "(not found)",
				"find . -not -path '*/node_modules/*' -not -path '*test*' -name '*.ts' | xargs grep -l 'authenticate\\|passport\\|jwt\\.verify\\|session' 2>/dev/null | head -2 | xargs -I{} sh -c 'echo \"=== {} ===\"; head -150 \"{}\"' 2>/dev/null || echo '(not found)'"),
			PermissionSystem: shIn(tmpDir, "(not found)",
				"find . -not -path '*/node_modules/*' -not -path '*test*' -name '*.ts' | xargs grep -l 'permission\\|authorize\\|rbac\\|\\bACL\\b' 2>/dev/null | head -2 | xargs -I{} sh -c 'echo \"=== {} ===\"; head -150 \"{}\"' 2>/dev/null || echo '(not found)'"),
			SecurityConfig: shIn(tmpDir, "(not found)",
				"grep -rn 'helmet\\|cors(\\|session(\\|cookieParser' --include='*.ts' . | grep -v node_modules | head -30 || echo '(not found)'"),
			StartupValidation: shIn(tmpDir, "(not found)",
				"find . -not -path '*/node_modules/*' -not -path '*test*' -name '*.ts' | xargs grep -l 'process\\.exit\\|logger\\.warn\\|logger\\.error' 2>/dev/null | xargs grep -l 'SECRET\\|NODE_ENV\\|env\\[' 2>/dev/null | head -2 | xargs -I{} sh -c 'echo \"=== {} ===\"; head -200 \"{}\"' 2>/dev/null || echo '(not found)'"),
			ErrorHandler: shIn(tmpDir, "(not found)",
				"find . -not -path '*/node_modules/*' -not -path '*test*' -name '*.ts' | xargs grep -l 'error.*handler\\|ErrorHandler\\|err.*Request.*Response' 2>/dev/null | head -2 | xargs -I{} sh -c 'echo \"=== {} ===\"; head -150 \"{}\"' 2>/dev/null || echo '(not found)'"),
			HelmetConfig: shIn(tmpDir, "(not found)",
				"find . -not -path '*/node_modules/*' -not -path '*test*' -name '*.ts' | xargs grep -l 'helmet(' 2>/dev/null | head -2 | xargs -I{} sh -c 'echo \"=== {} ===\"; grep -n -A 40 \"helmet(\" \"{}\"' 2>/dev/null || echo '(not found)'"),
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
			PnpmAudit: shIn(tmpDir, "no package manager available",
				"npm audit --json 2>&1 | head -300 || pnpm audit --json 2>&1 | head -300 || echo 'no package manager available'"),
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

	ctx.Secrets = auditSecrets{
		Gitleaks: shIn(tmpDir, "gitleaks not installed — skipped",
			"gitleaks detect --source . --no-git --report-format json 2>&1 | head -200 || echo 'gitleaks not installed — skipped'"),
		TruffleHog: shIn(tmpDir, "trufflehog not installed — skipped",
			"trufflehog filesystem . --json --no-update 2>&1 | head -200 || echo 'trufflehog not installed — skipped'"),
		PrivateKeyHeaders: shIn(tmpDir, "none",
			"grep -rn '-----BEGIN.*PRIVATE KEY-----' . --include='*.pem' --include='*.key' --include='*.env' | grep -v node_modules | head -20 || echo 'none'"),
		EnvFiles: shIn(tmpDir, "(none found)",
			"find . -name '.env*' -not -path '*/node_modules/*' | head -5 | xargs grep -v '^#' 2>/dev/null | head -50 || echo '(none found)'"),
		TokenPatterns: shIn(tmpDir, "none",
			`grep -rEn "AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+" . | grep -v node_modules | grep -v '\.git' | cut -c-300 | head -20 || echo 'none'`),
	}

	ctx.IaC = auditIaC{
		TerraformFiles: shIn(tmpDir, "(none found)",
			"find . -name '*.tf' -not -path '*/node_modules/*' | head -20 || echo '(none found)'"),
		Checkov: shIn(tmpDir, "checkov not installed — skipped",
			"checkov -d . --quiet --compact --output json 2>&1 | head -300 || echo 'checkov not installed — skipped'"),
		Trivy: shIn(tmpDir, "trivy not installed — skipped",
			"trivy config . --format json --quiet 2>&1 | head -300 || echo 'trivy not installed — skipped'"),
		KubeManifests: shIn(tmpDir, "(none found)",
			"grep -rl 'kind:' --include='*.yaml' --include='*.yml' . | grep -v node_modules | head -20 || echo '(none found)'"),
		KubeLinter: shIn(tmpDir, "kube-linter not installed — skipped",
			"kube-linter lint . 2>&1 | head -100 || echo 'kube-linter not installed — skipped'"),
	}

	ctx.Policy = auditPolicy{
		OPAFiles: shIn(tmpDir, "(none found)",
			"find . -name '*.rego' -not -path '*/node_modules/*' | head -10 || echo '(none found)'"),
		KyvernoFiles: shIn(tmpDir, "(none found)",
			"grep -rl 'kind: ClusterPolicy\\|kind: Policy' --include='*.yaml' . | grep -v node_modules | head -10 || echo '(none found)'"),
		FalcoRules: shIn(tmpDir, "(none found)",
			"grep -rl 'rule:' --include='*.yaml' . | xargs grep -l 'condition:\\|output:' 2>/dev/null | head -10 || echo '(none found)'"),
	}

	ctx.SLSA = auditSLSA{
		ProvenanceFiles: shIn(tmpDir, "(none found)",
			"find . -name '*.intoto.jsonl' -o -name 'provenance.json' | head -5 || echo '(none found)'"),
		SBOMFiles: shIn(tmpDir, "(none found)",
			"find . \\( -name '*.spdx' -o -name '*.spdx.json' -o -name '*.cyclonedx.json' -o -name 'sbom*.json' \\) | head -5 || echo '(none found)'"),
		CosignFiles: shIn(tmpDir, "(none found)",
			"find . -name 'cosign.pub' -o -name '*.pem' | grep -v node_modules | head -5 || echo '(none found)'"),
		SLSAWorkflow: shIn(tmpDir, "none",
			"grep -rn 'slsa-framework/slsa-github-generator\\|sigstore/cosign-action\\|actions/attest-build-provenance' .github/workflows/ 2>/dev/null || echo 'none'"),
		SignedCommit: shIn(tmpDir, "(unavailable)",
			"git log --show-signature -1 2>/dev/null | head -30 || echo '(unavailable)'"),
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

	bp, _ := fetch(base + "/branches/main/protection")
	bpJSON, _ := json.Marshal(bp)

	return auditGitHub{
		OpenIssues:       issues,
		OpenPRs:          prs,
		SecurityAlerts:   alerts,
		BranchProtection: string(bpJSON),
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
9. SHIFT-LEFT — close with a table: Finding | Manual check | Automated CI gate | CI YAML snippet. ` +
	`The CI YAML snippet column must contain a runnable GitHub Actions step (≤10 lines) that implements the gate.`

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
7. **P2 Recommendations** — table: ID | Title | Effort (hours/days) | Risk Reduction (Low/Medium/High) | Notes
   Estimate effort as engineer-hours for a mid-level contributor. Risk reduction is the severity of the gap being closed.
8. **Remediation Status table** — all findings with commit or PR reference where fixed
9. **Verification Checklist** — numbered list of copy-paste commands, one per finding
10. **Shift-left guardrails** — table: Finding | Manual check | Automated CI gate | CI YAML snippet
    The CI YAML snippet column must contain a runnable GitHub Actions step (≤10 lines) that implements the gate.
11. **Appendix: Full Application Security Assessment** — one H3 subsection per category `+
		`(SQL Injection, Authentication, Authorisation, Deserialization, SSRF, XXE, Path Traversal, `+
		`Cryptography, Rate Limiting, CORS, Dependencies, HTTP Headers, Container, Kubernetes/Helm, `+
		`Secrets / Credential Hygiene, IaC Security, Policy as Code, SLSA / Supply Chain). `+
		`Each subsection must start with the methodology note before listing observations. `+
		`For SQL Injection: use keyFiles.entryPoint and code.sqlInjection; ORM raw call with string concatenation = High. `+
		`For Authentication: read keyFiles.authMiddleware carefully; missing jwt.verify or session fixation = Critical. `+
		`For Authorisation: read keyFiles.permissionSystem; route handler without permission check = High. `+
		`For SSRF: evaluate code.ssrf against keyFiles; user-controlled URL passed to fetch/axios/got = Critical. `+
		`For Rate Limiting: if code.rateLimiting is empty or 'none', flag missing rate limiting as Medium/High on auth endpoints. `+
		`For CORS: if code.corsConfig shows origin:'*' with credentials, flag as Medium. `+
		`For Secrets: assess Gitleaks output or grep results, estimate false positive rate, list any confirmed leaks as Critical findings. `+
		`For IaC: synthesise Checkov and Trivy findings, flag unencrypted storage / overly-permissive IAM / missing network policies. `+
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
		`12. **Threat Model (STRIDE)** — table: Threat | STRIDE category `+
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
7. **P2 Recommendations** — table: ID | Title | Effort (hours/days) | Risk Reduction (Low/Medium/High) | Notes
   Estimate effort as engineer-hours for a mid-level contributor. Risk reduction is the severity of the gap being closed.
8. **Remediation Status table** — all findings with commit or PR reference where fixed
9. **Verification Checklist** — numbered list of copy-paste commands, one per finding
10. **Shift-left guardrails** — table: Finding | Manual check | Automated CI gate | CI YAML snippet
    The CI YAML snippet column must contain a runnable GitHub Actions step (≤10 lines) that implements the gate.
11. **Appendix: Full Application Security Assessment** — one H3 subsection per category.
Each appendix subsection must start with the methodology note before listing observations.
12. **Threat Model (STRIDE)** — table: Threat | STRIDE category (Spoofing / Tampering / Repudiation / Information Disclosure / Denial of Service / Elevation of Privilege) | Affected component | Existing mitigation | Residual risk (High/Med/Low).
Cover at least: authentication flows, CI/CD pipeline, supply-chain, secrets storage, API inputs.`,
		ctx.Meta.Repo, ctx.Meta.Ref, dateShort, summaryMD)
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

const defaultModel = "claude-opus-4-8"

// ── Ollama API ────────────────────────────────────────────────────────────────

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type ollamaStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

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

// compactForOllama returns a shallow copy of ctx with verbose tool outputs
// truncated so the total JSON fits within typical Ollama context windows.
func compactForOllama(ctx *auditContext) *auditContext {
	c := *ctx
	c.CICD.WorkflowContents = truncateField(ctx.CICD.WorkflowContents, 8_000)
	c.CICD.Zizmor = truncateField(ctx.CICD.Zizmor, 3_000)
	c.KeyFiles.AuthMiddleware = truncateField(ctx.KeyFiles.AuthMiddleware, 5_000)
	c.KeyFiles.PermissionSystem = truncateField(ctx.KeyFiles.PermissionSystem, 3_000)
	c.KeyFiles.StartupValidation = truncateField(ctx.KeyFiles.StartupValidation, 5_000)
	c.KeyFiles.ErrorHandler = truncateField(ctx.KeyFiles.ErrorHandler, 3_000)
	c.KeyFiles.HelmetConfig = truncateField(ctx.KeyFiles.HelmetConfig, 3_000)
	c.Dependencies.PnpmAudit = truncateField(ctx.Dependencies.PnpmAudit, 3_000)
	c.Secrets.Gitleaks = truncateField(ctx.Secrets.Gitleaks, 2_000)
	c.Secrets.TruffleHog = truncateField(ctx.Secrets.TruffleHog, 2_000)
	c.IaC.Checkov = truncateField(ctx.IaC.Checkov, 4_000)
	c.IaC.Trivy = truncateField(ctx.IaC.Trivy, 3_000)
	c.IaC.KubeLinter = truncateField(ctx.IaC.KubeLinter, 2_000)
	return &c
}

func generateOllamaReport(ctx *auditContext, ollamaURL, model string) (report string, inputTokens, outputTokens int, err error) {
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	return generateOllamaReportWith(ctx, ollamaURL, model, false)
}

func generateSplitOllamaReport(ctx *auditContext, ollamaURL, analysisModel, finalModel string) (report string, inputTokens, outputTokens int, err error) {
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	if analysisModel == "" {
		analysisModel = finalModel
	}

	sections := buildAuditSummarySections(ctx)
	summaries := make([]auditSectionSummary, 0, len(sections))
	for _, section := range sections {
		evidence := section.Content
		if len(evidence) > ollamaMaxSummaryPromptChars {
			evidence = evidence[:ollamaMaxSummaryPromptChars] + "\n\n[Section evidence truncated to fit the analysis model context window]"
		}
		prompt := fmt.Sprintf(`Summarize this DevSecOps audit evidence section for a later final report writer.

Section: %s

Rules:
- Keep concrete file paths, tool outputs, commands, package names, workflow names, and API evidence.
- Identify actionable findings, likely false positives, and clear "no issue found" categories.
- Calibrate severity. Do not inflate weak evidence.
- Output concise Markdown with headings: Evidence, Findings, No-Issue Notes, Open Questions.
- Do not write the final report.

Evidence:

%s`, section.Name, evidence)

		summary, in, out, serr := generateOllamaChat(ollamaURL, analysisModel, auditSummarySystemPrompt, prompt)
		inputTokens += in
		outputTokens += out
		if serr != nil {
			return "", inputTokens, outputTokens, fmt.Errorf("summarize %s: %w", section.Name, serr)
		}
		summaries = append(summaries, auditSectionSummary{
			Section: section.Name,
			Model:   analysisModel,
			Summary: summary,
		})
	}

	finalPrompt := buildSummarizedUserPrompt(ctx, summaries)
	report, in, out, err := generateOllamaChat(ollamaURL, finalModel, auditSystemPrompt, finalPrompt)
	inputTokens += in
	outputTokens += out
	if err != nil {
		return "", inputTokens, outputTokens, fmt.Errorf("final report: %w", err)
	}
	return report, inputTokens, outputTokens, nil
}

const auditSummarySystemPrompt = `You are a careful DevSecOps evidence summarizer. ` +
	`Your job is to reduce one section of raw audit data into compact, accurate evidence for another model. ` +
	`Never invent findings. Preserve concrete evidence and uncertainty.`

const ollamaMaxSummaryPromptChars = 250_000

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
			w("```yaml\n%s\n```", truncateField(ctx.CICD.WorkflowContents, 15_000))
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
			w("### pnpm / npm audit\n```\n%s\n```", ctx.Dependencies.PnpmAudit)
			w("### Workspace overrides\n```\n%s\n```", ctx.Dependencies.WorkspaceOverrides)
			w("## Secrets Scanning")
			w("### Gitleaks\n```\n%s\n```", ctx.Secrets.Gitleaks)
			w("### TruffleHog\n```\n%s\n```", ctx.Secrets.TruffleHog)
			w("### Private key headers\n```\n%s\n```", ctx.Secrets.PrivateKeyHeaders)
			w("### .env files\n```\n%s\n```", ctx.Secrets.EnvFiles)
			w("### Token patterns\n```\n%s\n```", truncateField(ctx.Secrets.TokenPatterns, 3_000))
		})},
		{Name: "Infrastructure, containers, Kubernetes, and SLSA", Content: sectionMD(func(w func(string, ...any)) {
			w("## Infrastructure")
			w("### Dockerfile\n```dockerfile\n%s\n```", ctx.Infra.Dockerfile)
			w("### Helm lint\n```\n%s\n```", ctx.Infra.HelmLint)
			w("### Helm secret template\n```yaml\n%s\n```", ctx.Infra.HelmSecretTemplate)
			w("### Helm values\n```yaml\n%s\n```", ctx.Infra.HelmValues)
			w("## IaC")
			w("### Terraform files\n```\n%s\n```", ctx.IaC.TerraformFiles)
			w("### Checkov\n```\n%s\n```", ctx.IaC.Checkov)
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
			w("### Open issues\n```json\n%s\n```", truncateField(string(issuesJSON), 20_000))
			w("### Open PRs\n```json\n%s\n```", truncateField(string(prsJSON), 10_000))
			w("### Secret-scanning alerts\n```\n%s\n```", ctx.GitHub.SecurityAlerts)
			w("### Branch protection\n```json\n%s\n```", ctx.GitHub.BranchProtection)
		})},
	}
}

// ollamaMaxPromptChars targets ~200k tokens (3 chars/token for code-heavy content),
// leaving headroom for the system prompt under 262k context windows.
const ollamaMaxPromptChars = 600_000

func generateOllamaReportWith(ctx *auditContext, ollamaURL, model string, compact bool) (report string, inputTokens, outputTokens int, err error) {
	sendCtx := ctx
	if compact {
		sendCtx = compactForOllama(ctx)
	}
	userPrompt := buildUserPrompt(sendCtx)
	if compact && len(userPrompt) > ollamaMaxPromptChars {
		userPrompt = userPrompt[:ollamaMaxPromptChars] + "\n\n[Context truncated to fit model context window]"
	}
	report, inputTokens, outputTokens, err = generateOllamaChat(ollamaURL, model, auditSystemPrompt, userPrompt)
	if err != nil {
		if !compact && isOllamaCompactRetryError(err) {
			return generateOllamaReportWith(ctx, ollamaURL, model, true)
		}
		return "", 0, 0, err
	}

	return report, inputTokens, outputTokens, nil
}

func generateOllamaChat(ollamaURL, model, systemPrompt, userPrompt string) (report string, inputTokens, outputTokens int, err error) {
	payload := ollamaRequest{
		Model:  model,
		Stream: true,
		Messages: []ollamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Hour}
	resp, err := client.Do(req)
	if err != nil {
		hint := ""
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "refused") {
			hint = " (is Ollama running and reachable? if inside Docker, set OLLAMA_HOST=0.0.0.0 on the host)"
		}
		return "", 0, 0, fmt.Errorf("ollama request failed: %w%s", err, hint)
	}
	defer resp.Body.Close() //nolint:errcheck

	report, inputTokens, outputTokens, err = readOllamaStream(resp)
	if err != nil {
		return "", 0, 0, err
	}

	return report, inputTokens, outputTokens, nil
}

func isOllamaCompactRetryError(err error) bool {
	msg := err.Error()
	contextTooLong := strings.Contains(msg, "exceeds") && strings.Contains(msg, "context")
	backendCrash := strings.Contains(msg, "EOF") || strings.Contains(msg, "api_error")
	return contextTooLong || backendCrash
}

func readOllamaStream(resp *http.Response) (report string, inputTokens, outputTokens int, err error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var or_ ollamaResponse
		if derr := json.Unmarshal(body, &or_); derr == nil && or_.Error != nil {
			return "", 0, 0, fmt.Errorf("ollama error %s: %s", or_.Error.Type, or_.Error.Message)
		}
		return "", 0, 0, fmt.Errorf("ollama HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var b strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		if line == "[DONE]" {
			break
		}

		var chunk ollamaStreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return "", 0, 0, fmt.Errorf("ollama stream decode failed: %w", err)
		}
		if chunk.Error != nil {
			return "", 0, 0, fmt.Errorf("ollama error %s: %s", chunk.Error.Type, chunk.Error.Message)
		}
		if chunk.Usage.PromptTokens > 0 {
			inputTokens = chunk.Usage.PromptTokens
		}
		if chunk.Usage.CompletionTokens > 0 {
			outputTokens = chunk.Usage.CompletionTokens
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				b.WriteString(choice.Delta.Content)
			} else if choice.Message.Content != "" {
				b.WriteString(choice.Message.Content)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", 0, 0, fmt.Errorf("ollama stream failed: %w", err)
	}
	if b.Len() == 0 {
		return "", 0, 0, fmt.Errorf("ollama returned empty content")
	}
	return b.String(), inputTokens, outputTokens, nil
}

// extractZizmortFindings parses a zizmor SARIF JSON and returns a compact Markdown
// findings table. Non-parseable input (e.g. "zizmor not installed") is returned as-is.
func extractZizmortFindings(sarifJSON string) string {
	var sarif struct {
		Runs []struct {
			Results []struct {
				RuleID  string `json:"ruleId"`
				Level   string `json:"level"`
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal([]byte(sarifJSON), &sarif); err != nil || len(sarif.Runs) == 0 {
		return truncateField(sarifJSON, 3_000)
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
	w("%s", truncateField(ctx.CICD.WorkflowContents, 15_000))
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
	w("### pnpm / npm audit")
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
	w("```json")
	w("%s", truncateField(string(issuesJSON), 20_000))
	w("```")
	w("### Open PRs")
	w("```json")
	w("%s", truncateField(string(prsJSON), 10_000))
	w("```")
	w("### Secret-scanning alerts")
	w("```")
	w("%s", ctx.GitHub.SecurityAlerts)
	w("```")
	w("### Branch protection (main)")
	w("```json")
	w("%s", ctx.GitHub.BranchProtection)
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
	w("### Checkov")
	w("```")
	w("%s", ctx.IaC.Checkov)
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
func generateTemplateReport(ctx *auditContext) string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }

	w("# Security Data Snapshot: %s", ctx.Meta.Repo)
	w("")
	w("**Date:** %s | **Commit:** %s | **Mode:** Static analysis only (no AI synthesis)", ctx.Meta.Date[:10], ctx.Meta.Ref)
	w("")
	w("---")
	w("")
	w("## CI/CD")
	w("")
	w("### Unpinned GitHub Actions (`uses: action@vX` tags)")
	w("")
	w("```")
	w("%s", ctx.CICD.UnpinnedActions)
	w("```")
	w("")
	w("### Actionlint")
	w("")
	w("```")
	w("%s", ctx.CICD.Actionlint)
	w("```")
	w("")
	w("### Workflow files")
	w("")
	w("```")
	w("%s", ctx.CICD.WorkflowList)
	w("```")
	w("")
	w("### Zizmor analysis")
	w("")
	w("```")
	w("%s", ctx.CICD.Zizmor)
	w("```")
	w("")
	w("### Security workflow contents (codeql / trivy / scorecard triggers)")
	w("")
	w("```yaml")
	w("%s", ctx.CICD.WorkflowContents)
	w("```")
	w("")
	w("---")
	w("")
	w("## Code Patterns")
	w("")
	w("### eval() usage")
	w("")
	w("```")
	w("%s", ctx.Code.EvalUsage)
	w("```")
	w("")
	w("### Math.random() usage")
	w("")
	w("```")
	w("%s", ctx.Code.MathRandom)
	w("```")
	w("")
	w("### Raw SQL calls")
	w("")
	w("```")
	w("%s", ctx.Code.RawSqlCalls)
	w("```")
	w("")
	w("### X-Powered-By header exposure")
	w("")
	w("```")
	w("%s", ctx.Code.XPoweredByHeader)
	w("```")
	w("")
	w("### Hardcoded secret hints")
	w("")
	w("```")
	w("%s", ctx.Code.HardcodedSecretHints)
	w("```")
	w("")
	w("### Weak crypto (MD5/SHA1)")
	w("")
	w("```")
	w("%s", ctx.Code.WeakCrypto)
	w("```")
	w("")
	w("### process.exit / os.Exit calls")
	w("")
	w("```")
	w("%s", ctx.Code.ProcessExitCalls)
	w("```")
	w("")
	w("### SQL injection patterns")
	w("")
	w("```")
	w("%s", ctx.Code.SqlInjection)
	w("```")
	w("")
	w("### SSRF (fetch / axios / got)")
	w("")
	w("```")
	w("%s", ctx.Code.SSRF)
	w("```")
	w("")
	w("### Path traversal (readFile / path.join)")
	w("")
	w("```")
	w("%s", ctx.Code.PathTraversal)
	w("```")
	w("")
	w("### XXE (xml2js / DOMParser)")
	w("")
	w("```")
	w("%s", ctx.Code.XXE)
	w("```")
	w("")
	w("### Deserialization (yaml.load / JSON.parse)")
	w("")
	w("```")
	w("%s", ctx.Code.Deserialization)
	w("```")
	w("")
	w("### Rate limiting")
	w("")
	w("```")
	w("%s", ctx.Code.RateLimiting)
	w("```")
	w("")
	w("### CORS config")
	w("")
	w("```")
	w("%s", ctx.Code.CORSConfig)
	w("```")
	w("")
	w("---")
	w("")
	w("## Key Security Files")
	w("")
	w("### Entry point")
	w("")
	w("```typescript")
	w("%s", ctx.KeyFiles.EntryPoint)
	w("```")
	w("")
	w("### Auth middleware")
	w("")
	w("```typescript")
	w("%s", ctx.KeyFiles.AuthMiddleware)
	w("```")
	w("")
	w("### Permission system")
	w("")
	w("```typescript")
	w("%s", ctx.KeyFiles.PermissionSystem)
	w("```")
	w("")
	w("### Security config (helmet / cors / session)")
	w("")
	w("```")
	w("%s", ctx.KeyFiles.SecurityConfig)
	w("```")
	w("")
	w("### Startup validation (SECRET / env enforcement)")
	w("")
	w("```typescript")
	w("%s", ctx.KeyFiles.StartupValidation)
	w("```")
	w("")
	w("### Error handler")
	w("")
	w("```typescript")
	w("%s", ctx.KeyFiles.ErrorHandler)
	w("```")
	w("")
	w("### Helmet config")
	w("")
	w("```typescript")
	w("%s", ctx.KeyFiles.HelmetConfig)
	w("```")
	w("")
	w("---")
	w("")
	w("## Infrastructure")
	w("")
	w("### Dockerfile")
	w("")
	w("```dockerfile")
	w("%s", ctx.Infra.Dockerfile)
	w("```")
	w("")
	w("### Helm lint")
	w("")
	w("```")
	w("%s", ctx.Infra.HelmLint)
	w("```")
	w("")
	w("### Helm secret template")
	w("")
	w("```yaml")
	w("%s", ctx.Infra.HelmSecretTemplate)
	w("```")
	w("")
	w("### Helm values")
	w("")
	w("```yaml")
	w("%s", ctx.Infra.HelmValues)
	w("```")
	w("")
	w("---")
	w("")
	w("## Dependencies")
	w("")
	w("### pnpm / npm audit")
	w("")
	w("```")
	w("%s", ctx.Dependencies.PnpmAudit)
	w("```")
	w("")
	w("### Workspace overrides")
	w("")
	w("```")
	w("%s", ctx.Dependencies.WorkspaceOverrides)
	w("```")
	w("")
	w("---")
	w("")
	w("## Git History")
	w("")
	w("### Recent commits")
	w("")
	w("```")
	w("%s", ctx.Git.RecentCommits)
	w("```")
	w("")
	w("### Recently changed files (last 10 commits)")
	w("")
	w("```")
	w("%s", ctx.Git.RecentlyChangedFiles)
	w("```")
	w("")
	w("---")
	w("")
	w("## GitHub")
	w("")

	if issues, ok := ctx.GitHub.OpenIssues.([]interface{}); ok {
		w("**Open issues:** %d", len(issues))
	} else {
		w("**Open issues:** (unavailable)")
	}

	if prs, ok := ctx.GitHub.OpenPRs.([]interface{}); ok {
		w("**Open PRs:** %d", len(prs))
	} else {
		w("**Open PRs:** (unavailable)")
	}

	w("")
	w("### Secret-scanning alerts")
	w("")
	w("```")
	w("%s", ctx.GitHub.SecurityAlerts)
	w("```")
	w("")
	w("### Branch protection (main)")
	w("")
	w("```json")
	w("%s", ctx.GitHub.BranchProtection)
	w("```")
	w("")
	w("---")
	w("")
	w("## Secrets Scanning")
	w("")
	w("### Gitleaks")
	w("")
	w("```")
	w("%s", ctx.Secrets.Gitleaks)
	w("```")
	w("")
	w("### TruffleHog")
	w("")
	w("```")
	w("%s", ctx.Secrets.TruffleHog)
	w("```")
	w("")
	w("### Private key headers")
	w("")
	w("```")
	w("%s", ctx.Secrets.PrivateKeyHeaders)
	w("```")
	w("")
	w("### .env files")
	w("")
	w("```")
	w("%s", ctx.Secrets.EnvFiles)
	w("```")
	w("")
	w("### Token patterns (AWS/JWT/GH)")
	w("")
	w("```")
	w("%s", ctx.Secrets.TokenPatterns)
	w("```")
	w("")
	w("---")
	w("")
	w("## Infrastructure as Code (IaC)")
	w("")
	w("### Terraform files")
	w("")
	w("```")
	w("%s", ctx.IaC.TerraformFiles)
	w("```")
	w("")
	w("### Checkov")
	w("")
	w("```")
	w("%s", ctx.IaC.Checkov)
	w("```")
	w("")
	w("### Trivy config")
	w("")
	w("```")
	w("%s", ctx.IaC.Trivy)
	w("```")
	w("")
	w("### Kubernetes manifests")
	w("")
	w("```")
	w("%s", ctx.IaC.KubeManifests)
	w("```")
	w("")
	w("### kube-linter")
	w("")
	w("```")
	w("%s", ctx.IaC.KubeLinter)
	w("```")
	w("")
	w("---")
	w("")
	w("## Policy as Code")
	w("")
	w("### OPA (.rego files)")
	w("")
	w("```")
	w("%s", ctx.Policy.OPAFiles)
	w("```")
	w("")
	w("### Kyverno policies")
	w("")
	w("```")
	w("%s", ctx.Policy.KyvernoFiles)
	w("```")
	w("")
	w("### Falco rules")
	w("")
	w("```")
	w("%s", ctx.Policy.FalcoRules)
	w("```")
	w("")
	w("---")
	w("")
	w("## SLSA / Supply Chain")
	w("")
	w("### Provenance files")
	w("")
	w("```")
	w("%s", ctx.SLSA.ProvenanceFiles)
	w("```")
	w("")
	w("### SBOM files")
	w("")
	w("```")
	w("%s", ctx.SLSA.SBOMFiles)
	w("```")
	w("")
	w("### Cosign / signing keys")
	w("")
	w("```")
	w("%s", ctx.SLSA.CosignFiles)
	w("```")
	w("")
	w("### SLSA / attestation workflow usage")
	w("")
	w("```")
	w("%s", ctx.SLSA.SLSAWorkflow)
	w("```")
	w("")
	w("### Signed commit (latest)")
	w("")
	w("```")
	w("%s", ctx.SLSA.SignedCommit)
	w("```")

	return b.String()
}

func callClaude(systemPrompt, userPrompt, apiKey, model string, maxTokens int) (text string, inputTokens, outputTokens int, err error) {
	payload := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System: []claudeSystemBlock{
			{
				Type:         "text",
				Text:         systemPrompt,
				CacheControl: &claudeCacheControl{Type: "ephemeral"},
			},
		},
		Messages: []claudeMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Minute}

	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")

		resp, rerr := client.Do(req)
		if rerr != nil {
			return "", 0, 0, fmt.Errorf("claude API request failed: %w", rerr)
		}

		var cr claudeResponse
		decErr := json.NewDecoder(resp.Body).Decode(&cr)
		resp.Body.Close() //nolint:errcheck
		if decErr != nil {
			return "", 0, 0, fmt.Errorf("claude API decode failed: %w", decErr)
		}

		if cr.Error != nil {
			if cr.Error.Type == "rate_limit_error" && attempt == 0 {
				time.Sleep(65 * time.Second)
				continue
			}
			return "", 0, 0, fmt.Errorf("claude API error %s: %s", cr.Error.Type, cr.Error.Message)
		}
		if len(cr.Content) == 0 {
			return "", 0, 0, fmt.Errorf("claude API returned empty content")
		}
		return cr.Content[0].Text, cr.Usage.InputTokens, cr.Usage.OutputTokens, nil
	}
	return "", 0, 0, fmt.Errorf("claude API rate limit not resolved after retry")
}

func generateReport(ctx *auditContext, apiKey, model string) (report string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultModel
	}
	return callClaude(auditSystemPrompt, buildUserPrompt(ctx), apiKey, model, 8192)
}

func generateSplitClaudeReport(ctx *auditContext, apiKey, analysisModel, finalModel string) (report string, inputTokens, outputTokens int, err error) {
	if finalModel == "" {
		finalModel = defaultModel
	}
	if analysisModel == "" {
		analysisModel = "claude-haiku-4-5-20251001"
	}

	sections := buildAuditSummarySections(ctx)
	summaries := make([]auditSectionSummary, 0, len(sections))
	for _, section := range sections {
		prompt := fmt.Sprintf(`Summarize this DevSecOps audit evidence section for a later final report writer.

Section: %s

Rules:
- Keep concrete file paths, tool outputs, commands, package names, workflow names, and API evidence.
- Identify actionable findings, likely false positives, and clear "no issue found" categories.
- Calibrate severity. Do not inflate weak evidence.
- Output concise Markdown with headings: Evidence, Findings, No-Issue Notes, Open Questions.
- Do not write the final report.

Evidence:

%s`, section.Name, section.Content)

		summary, in, out, serr := callClaude(auditSummarySystemPrompt, prompt, apiKey, analysisModel, 2048)
		inputTokens += in
		outputTokens += out
		if serr != nil {
			return "", inputTokens, outputTokens, fmt.Errorf("summarize %s: %w", section.Name, serr)
		}
		summaries = append(summaries, auditSectionSummary{
			Section: section.Name,
			Model:   analysisModel,
			Summary: summary,
		})
	}

	report, in, out, err := callClaude(auditSystemPrompt, buildSummarizedUserPrompt(ctx, summaries), apiKey, finalModel, 8192)
	inputTokens += in
	outputTokens += out
	return report, inputTokens, outputTokens, err
}

// ── Runner ────────────────────────────────────────────────────────────────────

func runAudit(db *sql.DB, id, repo, ghToken, provider, anthropicKey, ollamaURL, model, analysisModel string, splitGeneration bool) {
	_ = dbUpdateAuditRunning(db, id)

	auditCtx, tmpDir, err := collectContext(repo, ghToken)
	if err != nil {
		_ = dbUpdateAuditError(db, id, fmt.Sprintf("collect failed: %v", err))
		return
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	if raw, merr := json.Marshal(auditCtx); merr == nil {
		_ = dbUpdateAuditContext(db, id, string(raw))
	}

	runAuditGenerate(db, id, auditCtx, provider, anthropicKey, ollamaURL, model, analysisModel, splitGeneration)
}

func runAuditFromContext(db *sql.DB, id, contextJSON, provider, anthropicKey, ollamaURL, model, analysisModel string, splitGeneration bool) {
	var auditCtx auditContext
	if err := json.Unmarshal([]byte(contextJSON), &auditCtx); err != nil {
		_ = dbUpdateAuditError(db, id, "context unmarshal failed: "+err.Error())
		return
	}
	runAuditGenerate(db, id, &auditCtx, provider, anthropicKey, ollamaURL, model, analysisModel, splitGeneration)
}

func runAuditGenerate(db *sql.DB, id string, auditCtx *auditContext, provider, anthropicKey, ollamaURL, model, analysisModel string, splitGeneration bool) {
	var report string
	var inputTokens, outputTokens int
	var err error

	switch provider {
	case "anthropic":
		if splitGeneration {
			report, inputTokens, outputTokens, err = generateSplitClaudeReport(auditCtx, anthropicKey, analysisModel, model)
		} else {
			report, inputTokens, outputTokens, err = generateReport(auditCtx, anthropicKey, model)
		}
	case "ollama":
		if splitGeneration {
			report, inputTokens, outputTokens, err = generateSplitOllamaReport(auditCtx, ollamaURL, analysisModel, model)
		} else {
			report, inputTokens, outputTokens, err = generateOllamaReport(auditCtx, ollamaURL, model)
		}
	default:
		_ = dbUpdateAuditDone(db, id, generateTemplateReport(auditCtx), 0, 0)
		return
	}

	if err != nil {
		// Save the static snapshot as fallback so the user has downloadable data
		_ = dbUpdateAuditErrorWithReport(db, id,
			fmt.Sprintf("generate failed: %v", err),
			generateTemplateReport(auditCtx))
		return
	}
	// Ground every concrete claim against the collected evidence and append the
	// verification appendix; unverifiable claims flip the report to DRAFT.
	report = verifyReport(report, auditCtx)
	_ = dbUpdateAuditDone(db, id, report, inputTokens, outputTokens)
}

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
			`grep -rEn "AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+" . | grep -v node_modules | grep -v '\.git' | head -20 || echo 'none'`),
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
		`X-Powered-By suppression with full context — distinguish deliberate upstream defaults from oversights.`,
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

func generateOllamaReport(ctx *auditContext, ollamaURL, model string) (report string, inputTokens, outputTokens int, err error) {
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	payload := ollamaRequest{
		Model:  model,
		Stream: false,
		Messages: []ollamaMessage{
			{Role: "system", Content: auditSystemPrompt},
			{Role: "user", Content: buildUserPrompt(ctx)},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	var or_ ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&or_); err != nil {
		return "", 0, 0, fmt.Errorf("ollama decode failed: %w", err)
	}
	if or_.Error != nil {
		return "", 0, 0, fmt.Errorf("ollama error %s: %s", or_.Error.Type, or_.Error.Message)
	}
	if len(or_.Choices) == 0 || or_.Choices[0].Message.Content == "" {
		return "", 0, 0, fmt.Errorf("ollama returned empty content")
	}

	return or_.Choices[0].Message.Content, or_.Usage.PromptTokens, or_.Usage.CompletionTokens, nil
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

func generateReport(ctx *auditContext, apiKey, model string) (report string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultModel
	}
	payload := claudeRequest{
		Model:     model,
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

func runAudit(db *sql.DB, id, repo, ghToken, provider, anthropicKey, ollamaURL, model string) {
	_ = dbUpdateAuditRunning(db, id)

	auditCtx, tmpDir, err := collectContext(repo, ghToken)
	if err != nil {
		_ = dbUpdateAuditError(db, id, fmt.Sprintf("collect failed: %v", err))
		return
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	var report string
	var inputTokens, outputTokens int

	switch provider {
	case "anthropic":
		report, inputTokens, outputTokens, err = generateReport(auditCtx, anthropicKey, model)
	case "ollama":
		report, inputTokens, outputTokens, err = generateOllamaReport(auditCtx, ollamaURL, model)
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
	_ = dbUpdateAuditDone(db, id, report, inputTokens, outputTokens)
}

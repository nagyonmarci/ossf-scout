package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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
	// PinnedSuggestions maps each unpinned action to its resolved commit SHA so the
	// report writer (and the verifier) use ground-truth SHAs instead of invented ones.
	PinnedSuggestions string `json:"pinnedSuggestions"`
	Zizmor            string `json:"zizmor"`
	Actionlint        string `json:"actionlint"`
	WorkflowList      string `json:"workflowList"`
	WorkflowContents  string `json:"workflowContents"`
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
	SemgrepFindings      string `json:"semgrepFindings"`
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
	IssuesStatus     string      `json:"issuesStatus"`
	PRsStatus        string      `json:"prsStatus"`
	SecurityAlerts   string      `json:"securityAlerts"`
	BranchProtection string      `json:"branchProtection"`
	DependabotAlerts string      `json:"dependabotAlerts"`
	ReleaseHistory   string      `json:"releaseHistory"`
	DefaultBranch    string      `json:"defaultBranch"`
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
	OSVScanner     string `json:"osvScanner"`
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
	CodeOwners        string `json:"codeOwners"`
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

// usesLineRe parses a `grep -rn` line of an unpinned action use:
//
//	.github/workflows/ci.yml:42:      uses: actions/checkout@v4
var usesLineRe = regexp.MustCompile(`^(\S+?):(\d+):.*\buses:\s*['"]?([\w.-]+/[\w./-]+)@(v[0-9][^\s'"#]*)`)

func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// resolveTagSHA resolves owner/repo@tag to the commit SHA the tag points at,
// dereferencing annotated tags. Returns "unresolved" on any failure.
func resolveTagSHA(action, tag string) string {
	parts := strings.Split(action, "/")
	if len(parts) < 2 {
		return "unresolved"
	}
	repo := parts[0] + "/" + parts[1]
	cctx, cancel := context.WithTimeout(context.Background(), resolveTagTimeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "git", "ls-remote",
		"https://github.com/"+repo+".git",
		"refs/tags/"+tag, "refs/tags/"+tag+"^{}")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		return "unresolved"
	}
	var plain, deref string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.Fields(l)
		if len(f) != 2 {
			continue
		}
		if strings.HasSuffix(f[1], "^{}") {
			deref = f[0]
		} else {
			plain = f[0]
		}
	}
	if deref != "" {
		return deref
	}
	if plain != "" {
		return plain
	}
	return "unresolved"
}

// resolvePinnedActions turns the raw UnpinnedActions grep output into an
// authoritative `action@tag -> SHA | locations` table. The resolved SHAs feed
// both the report writer (real fix snippets) and the verifier (allow-set).
func resolvePinnedActions(unpinned string) string {
	t := strings.TrimSpace(unpinned)
	if t == "" || t == "none" {
		return "none"
	}
	type entry struct{ locs []string }
	order := []string{}
	seen := map[string]*entry{}
	for _, line := range strings.Split(unpinned, "\n") {
		m := usesLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		file, lineNo, action, tag := m[1], m[2], m[3], m[4]
		key := action + "@" + tag
		loc := file + ":" + lineNo
		if e, ok := seen[key]; ok {
			e.locs = append(e.locs, loc)
			continue
		}
		seen[key] = &entry{locs: []string{loc}}
		order = append(order, key)
	}
	if len(order) == 0 {
		return "none"
	}
	var b strings.Builder
	b.WriteString("# action@tag -> resolved commit SHA (authoritative; use in fixes) | used at\n")
	for i, key := range order {
		sha := "unresolved (resolve manually before citing)"
		if i < maxPinnedActionResolve {
			at := strings.SplitN(key, "@", 2)
			sha = resolveTagSHA(at[0], at[1])
		}
		fmt.Fprintf(&b, "%s -> %s | %s\n", key, sha, strings.Join(dedupeStrings(seen[key].locs), ", "))
	}
	return b.String()
}

// collectOptions controls which expensive tool categories are run during context collection.
type collectOptions struct {
	SkipSecrets bool // skip gitleaks and trufflehog (slow; set true when speed matters)
}

func collectContext(repo, ghToken string, opts collectOptions) (*auditContext, string, error) {
	owner, repoName, ok := splitValidRepo(repo)
	if !ok {
		return nil, "", fmt.Errorf("invalid repository name")
	}

	tmpDir, err := os.MkdirTemp("", "ossf-audit-*")
	if err != nil {
		return nil, "", fmt.Errorf("mktemp: %w", err)
	}

	// Build URL without credentials; auth is injected via GIT_CONFIG env vars instead.
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repoName)
	cloneCmd := exec.Command("git", "clone", "--depth=50", cloneURL, tmpDir)
	cloneCmd.Env = gitAuthEnv(ghToken)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir) //nolint:errcheck
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, "", fmt.Errorf("git clone failed: %s", msg)
	}

	ref := shIn(tmpDir, "unknown", "git rev-parse --short HEAD")

	// Detect IaC presence to skip expensive tools on non-IaC repos
	hasDockerfile := shIn(tmpDir, "", "test -f Dockerfile && echo 1") == "1"
	hasTerraform  := shIn(tmpDir, "", "find . -maxdepth 5 -name '*.tf' -not -path '*/node_modules/*' | head -1") != ""
	hasHelmChart  := shIn(tmpDir, "", "find . -maxdepth 6 -name 'Chart.yaml' | head -1") != ""
	hasK8s        := shIn(tmpDir, "", "find . -name '*.yaml' -not -path '*/node_modules/*' | xargs grep -l 'kind:' 2>/dev/null | head -1") != ""
	runIaC        := hasDockerfile || hasTerraform || hasHelmChart || hasK8s

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
			SemgrepFindings: shIn(tmpDir, "semgrep not installed — skipped",
				"semgrep --config=auto --json --timeout 60 --quiet . 2>&1 | head -500 || echo 'semgrep not installed — skipped'"),
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
			CodeOwners: shIn(tmpDir, "(not found)",
				"cat CODEOWNERS .github/CODEOWNERS docs/CODEOWNERS 2>/dev/null | head -80 || echo '(not found)'"),
		},
		Infra: auditInfra{
			HelmLint: func() string {
				if !hasHelmChart {
					return "skipped (no Helm chart detected)"
				}
				return shIn(tmpDir, "helm not installed — skipped", "helm lint helm/*/ 2>&1 || echo 'no helm chart or helm not installed'")
			}(),
			HelmSecretTemplate: shIn(tmpDir, "(not found)",
				"find . -path '*/helm/*/templates/secret.yaml' | head -1 | xargs cat 2>/dev/null || echo '(not found)'"),
			HelmValues: shIn(tmpDir, "(not found)",
				"find . -path '*/helm/*/values.yaml' | head -1 | xargs cat 2>/dev/null || echo '(not found)'"),
			Dockerfile: shIn(tmpDir, "(not found)",
				"cat Dockerfile 2>/dev/null || echo '(not found)'"),
		},
		Dependencies: auditDeps{
			// Pick the right tool by lockfile. The previous one-liner piped npm
			// audit into `head` (exit 0), so the `|| pnpm` fallback never fired and
			// pnpm workspaces returned the npm ENOLOCK error instead of real data.
			PnpmAudit: shIn(tmpDir, "no package manager available", `
if [ -f pnpm-lock.yaml ]; then pnpm audit --json 2>&1 | head -400;
elif [ -f package-lock.json ] || [ -f npm-shrinkwrap.json ]; then npm audit --json 2>&1 | head -400;
elif [ -f yarn.lock ]; then yarn audit --json 2>&1 | head -400;
elif [ -f package.json ]; then pnpm audit --json 2>&1 | head -400 || npm audit --json 2>&1 | head -400;
else echo 'no package manager available'; fi`),
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

	ctx.CICD.PinnedSuggestions = resolvePinnedActions(ctx.CICD.UnpinnedActions)

	ctx.Secrets = auditSecrets{
		Gitleaks: func() string {
			if opts.SkipSecrets {
				return "skipped (SkipSecrets=true)"
			}
			return shIn(tmpDir, "gitleaks not installed — skipped",
				"gitleaks detect --source . --no-git --report-format json 2>&1 | head -200 || echo 'gitleaks not installed — skipped'")
		}(),
		TruffleHog: func() string {
			if opts.SkipSecrets {
				return "skipped (SkipSecrets=true)"
			}
			return shIn(tmpDir, "trufflehog not installed — skipped",
				"trufflehog filesystem . --json --no-update 2>&1 | head -200 || echo 'trufflehog not installed — skipped'")
		}(),
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
		OSVScanner: shIn(tmpDir, "osv-scanner not installed — skipped",
			"osv-scanner --json . 2>&1 | head -400 || echo 'osv-scanner not installed — skipped'"),
		Trivy: func() string {
			if !runIaC {
				return "skipped (no IaC files detected)"
			}
			return shIn(tmpDir, "trivy not installed — skipped",
				"trivy config . --format json --quiet 2>&1 | head -300 || echo 'trivy not installed — skipped'")
		}(),
		KubeManifests: shIn(tmpDir, "(none found)",
			"grep -rl 'kind:' --include='*.yaml' --include='*.yml' . | grep -v node_modules | head -20 || echo '(none found)'"),
		KubeLinter: func() string {
			if !hasK8s {
				return "skipped (no Kubernetes manifests detected)"
			}
			return shIn(tmpDir, "kube-linter not installed — skipped",
				"kube-linter lint . 2>&1 | head -100 || echo 'kube-linter not installed — skipped'")
		}(),
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
		req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)
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

	// classify turns a decoded GitHub response into (data, status). A successful
	// list endpoint returns a JSON array; a rate-limit or error returns an object
	// with a "message" field. We keep data only when it is a genuine array, so the
	// report writer never sees error JSON it could mistake for real issues/PRs.
	classify := func(v interface{}, err error) (interface{}, string) {
		if err != nil {
			return nil, "unavailable: " + err.Error()
		}
		switch t := v.(type) {
		case []interface{}:
			return t, "ok"
		case map[string]interface{}:
			if msg, ok := t["message"].(string); ok {
				if len(msg) > 120 {
					msg = msg[:120]
				}
				return nil, "unavailable: " + msg
			}
		}
		return nil, "unavailable: unexpected response"
	}

	issuesRaw, issuesErr := fetch(base + "/issues?state=open&per_page=50")
	issues, issuesStatus := classify(issuesRaw, issuesErr)
	prsRaw, prsErr := fetch(base + "/pulls?state=open&per_page=20")
	prs, prsStatus := classify(prsRaw, prsErr)

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

	defaultBranch := "main"
	if metaRaw, err := fetch(base); err == nil {
		if metaMap, ok := metaRaw.(map[string]interface{}); ok {
			if db, ok := metaMap["default_branch"].(string); ok && db != "" {
				defaultBranch = db
			}
		}
	}

	bp, _ := fetch(base + "/branches/" + defaultBranch + "/protection")
	bpJSON, _ := json.Marshal(bp)

	// Dependabot alerts — requires token with security_events scope
	depAlertsStr := "(no token or insufficient permissions)"
	if ghToken != "" {
		depRaw, _ := fetch(base + "/dependabot/alerts?state=open&per_page=20")
		if depJSON, merr := json.Marshal(depRaw); merr == nil {
			depAlertsStr = string(depJSON)
		}
	}

	// Recent releases — helps assess security patch cadence
	relRaw, _ := fetch(base + "/releases?per_page=8")
	relJSON, _ := json.Marshal(relRaw)

	return auditGitHub{
		OpenIssues:       issues,
		OpenPRs:          prs,
		IssuesStatus:     issuesStatus,
		PRsStatus:        prsStatus,
		SecurityAlerts:   alerts,
		BranchProtection: string(bpJSON),
		DependabotAlerts: depAlertsStr,
		ReleaseHistory:   string(relJSON),
		DefaultBranch:    defaultBranch,
	}
}

// ghListAvailable reports whether GitHub list data is genuine (not a rate-limit
// or error). Legacy contexts saved before status tracking have an empty status
// but may still hold a real array, so those are honoured too.
func ghListAvailable(status string, data interface{}) bool {
	if status == "ok" {
		return true
	}
	if status == "" {
		_, ok := data.([]interface{})
		return ok
	}
	return false
}

func ghStatusLabel(status string) string {
	if status == "" {
		return "unavailable"
	}
	return status
}


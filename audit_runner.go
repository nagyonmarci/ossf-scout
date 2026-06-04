package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// auditParams bundles all provider-specific parameters for an audit run.
type auditParams struct {
	Provider        string
	AnthropicKey    string
	OpenAIKey       string
	GeminiKey       string
	OllamaURL       string
	Model           string
	AnalysisModel   string
	SplitGeneration bool
}

// computeAuditCost returns the estimated USD cost for a given model and token counts.
func computeAuditCost(model string, inputTokens, outputTokens int) float64 {
	prices, ok := modelPrices[model]
	if !ok {
		return 0
	}
	return float64(inputTokens)/1_000_000*prices[0] + float64(outputTokens)/1_000_000*prices[1]
}

// getRepoHeadSHA resolves the current HEAD commit SHA for a GitHub repo via git ls-remote.
func getRepoHeadSHA(repo, ghToken string) string {
	repoURL := "https://github.com/" + repo + ".git"
	if ghToken != "" {
		repoURL = "https://x-access-token:" + ghToken + "@github.com/" + repo + ".git"
	}
	cctx, cancel := context.WithTimeout(context.Background(), resolveTagTimeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "git", "ls-remote", repoURL, "HEAD")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 1 {
		return ""
	}
	return fields[0]
}

// ── Template report (static snapshot, no AI) ──────────────────────────────────

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
	w("### OSV-Scanner")
	w("")
	w("```")
	w("%s", ctx.IaC.OSVScanner)
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

// ── Audit orchestration ───────────────────────────────────────────────────────

func runAudit(db *sql.DB, id, repo, ghToken string, p auditParams) {
	_ = dbUpdateAuditRunning(db, id)

	// SHA-cache: resolve head SHA and check for a recent identical context.
	headSHA := getRepoHeadSHA(repo, ghToken)
	if headSHA != "" {
		_ = dbUpdateAuditHeadSHA(db, id, headSHA)
		if cached, cerr := dbFindRecentContext(db, repo, headSHA, 24*time.Hour); cerr == nil && cached != nil {
			_ = dbUpdateAuditContext(db, id, *cached)
			runAuditGenerate(db, id, mustUnmarshalContext(*cached), p)
			return
		}
	}

	auditCtx, tmpDir, err := collectContext(repo, ghToken)
	if err != nil {
		_ = dbUpdateAuditError(db, id, fmt.Sprintf("collect failed: %v", err))
		return
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	if raw, merr := json.Marshal(auditCtx); merr == nil {
		_ = dbUpdateAuditContext(db, id, string(raw))
	}

	runAuditGenerate(db, id, auditCtx, p)
}

func runAuditFromContext(db *sql.DB, id, contextJSON string, p auditParams) {
	auditCtx := mustUnmarshalContext(contextJSON)
	if auditCtx == nil {
		_ = dbUpdateAuditError(db, id, "context unmarshal failed")
		return
	}
	runAuditGenerate(db, id, auditCtx, p)
}

func mustUnmarshalContext(contextJSON string) *auditContext {
	var ctx auditContext
	if err := json.Unmarshal([]byte(contextJSON), &ctx); err != nil {
		return nil
	}
	return &ctx
}

// estimateAuditCost returns a rough pre-run cost estimate in USD based on
// context length and the target model's pricing.
func estimateAuditCost(contextJSON, model string) float64 {
	prices, ok := modelPrices[model]
	if !ok {
		return 0
	}
	estimatedInputTokens := len(contextJSON) / 4
	return float64(estimatedInputTokens)/1_000_000*prices[0] + float64(defaultMaxTokens)/1_000_000*prices[1]
}

func runAuditGenerate(db *sql.DB, id string, auditCtx *auditContext, p auditParams) {
	// Budget guard: reject before making any API call if estimated cost exceeds limit.
	if maxCostStr := os.Getenv("MAX_AUDIT_COST_USD"); maxCostStr != "" && p.Provider != "" && p.Provider != "ollama" {
		if maxCost, err := strconv.ParseFloat(maxCostStr, 64); err == nil && maxCost > 0 {
			raw, _ := json.Marshal(auditCtx)
			model := p.Model
			if model == "" {
				switch p.Provider {
				case "openai":
					model = defaultOpenAIModel
				case "gemini":
					model = defaultGeminiModel
				}
			}
			if est := estimateAuditCost(string(raw), model); est > maxCost {
				_ = dbUpdateAuditError(db, id,
					fmt.Sprintf("estimated cost $%.4f exceeds MAX_AUDIT_COST_USD=$%.2f; adjust limit or choose a cheaper model", est, maxCost))
				notifyWebhook(auditCtx.Meta.Repo, "error", id)
				return
			}
		}
	}

	var report string
	var inputTokens, outputTokens int
	var err error

	switch p.Provider {
	case "anthropic":
		if p.SplitGeneration {
			report, inputTokens, outputTokens, err = generateSplitClaudeReport(auditCtx, p.AnthropicKey, p.AnalysisModel, p.Model)
		} else {
			report, inputTokens, outputTokens, err = generateReport(auditCtx, p.AnthropicKey, p.Model)
		}
	case "openai":
		report, inputTokens, outputTokens, err = generateOpenAIReport(auditCtx, p.OpenAIKey, p.Model)
	case "gemini":
		report, inputTokens, outputTokens, err = generateGeminiReport(auditCtx, p.GeminiKey, p.Model)
	case "ollama":
		if p.SplitGeneration {
			report, inputTokens, outputTokens, err = generateSplitOllamaReport(auditCtx, p.OllamaURL, p.AnalysisModel, p.Model)
		} else {
			report, inputTokens, outputTokens, err = generateOllamaReport(auditCtx, p.OllamaURL, p.Model)
		}
	default:
		_ = dbUpdateAuditDone(db, id, generateTemplateReport(auditCtx), 0, 0)
		return
	}

	if err != nil {
		_ = dbUpdateAuditErrorWithReport(db, id,
			fmt.Sprintf("generate failed: %v", err),
			generateTemplateReport(auditCtx))
		notifyWebhook(auditCtx.Meta.Repo, "error", id)
		return
	}
	report = verifyReport(report, auditCtx)
	_ = dbUpdateAuditDone(db, id, report, inputTokens, outputTokens)
	notifyWebhook(auditCtx.Meta.Repo, "done", id)
}

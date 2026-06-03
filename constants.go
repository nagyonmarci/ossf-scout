package main

import "time"

// CLI / server defaults (mirrored in main.go flags and server.go request defaults)
const (
	defaultMinStars = 500
	defaultMaxScore = 5.0
	defaultLimit    = 100
	defaultWorkers  = 5
	defaultPort     = 7878
	defaultPageSize = 100 // GitHub search per_page
)

// Timeouts — defined as vars because time.Duration is not an untyped constant
var (
	scorecardCLITimeout = 10 * time.Minute
	resolveTagTimeout   = 15 * time.Second
	searchPageDelay     = 500 * time.Millisecond
)

// Audit collection / context limits
const (
	maxPinnedActionResolve = 60 // max actions to resolve SHAs for in resolvePinnedActions

	ollamaMaxSummaryPromptChars = 250_000 // per-section evidence cap for split analysis
	ollamaMaxPromptChars        = 600_000 // full prompt cap for single-stage Ollama

	// compactForOllama per-field byte limits
	truncWorkflowContents  = 8_000
	truncZizmor            = 3_000
	truncAuthMiddleware    = 5_000
	truncPermissionSystem  = 3_000
	truncStartupValidation = 5_000
	truncErrorHandler      = 3_000
	truncHelmetConfig      = 3_000
	truncDepsAudit         = 3_000
	truncGitleaks          = 2_000
	truncTruffleHog        = 2_000
	truncCheckov           = 4_000
	truncTrivy             = 3_000
	truncKubeLinter        = 2_000

	// buildAuditSummarySections / buildContextMarkdown per-field limits
	truncSummaryWorkflow = 15_000
	truncSummaryTokenPat = 3_000
	truncSummaryIssues   = 20_000
	truncSummaryPRs      = 10_000
	truncSarif           = 3_000
)

// Output / formatting
const tableWidth = 160

// GitHub API
const gitHubAPIVersion = "2022-11-28"

// Ollama
const defaultOllamaURL = "http://localhost:11434"

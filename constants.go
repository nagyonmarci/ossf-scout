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
	truncSemgrep           = 5_000

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

// OpenAI / Gemini provider defaults
const (
	defaultOpenAIModel = "gpt-4o"
	defaultGeminiModel = "gemini-2.0-flash"
)

// OpenAI / Gemini API endpoints
const (
	openAIEndpoint    = "https://api.openai.com/v1/chat/completions"
	geminiEndpointFmt = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
)

// modelPrices maps model name → [input $/1M tokens, output $/1M tokens]
var modelPrices = map[string][2]float64{
	// Anthropic — current
	"claude-opus-4-8":           {5.0, 25.0},
	"claude-sonnet-4-6":         {3.0, 15.0},
	"claude-haiku-4-5-20251001": {1.0, 5.0},
	// Anthropic — legacy (still available)
	"claude-opus-4-7":            {5.0, 25.0},
	"claude-opus-4-6":            {5.0, 25.0},
	"claude-sonnet-4-5-20250929": {3.0, 15.0},
	"claude-opus-4-5-20251101":   {5.0, 25.0},
	// OpenAI — GPT-5.x
	"gpt-5.5":      {5.0, 30.0},
	"gpt-5.4":      {2.5, 15.0},
	"gpt-5.4-mini": {0.75, 4.50},
	// OpenAI — GPT-4.x
	"gpt-4.1":      {2.0, 8.0},
	"gpt-4.1-mini": {0.40, 1.60},
	"gpt-4.1-nano": {0.10, 0.40},
	"gpt-4o":       {2.5, 10.0},
	"gpt-4o-mini":  {0.15, 0.60},
	// OpenAI — o-series
	"o1":      {15.0, 60.0},
	"o3":      {10.0, 40.0},
	"o3-mini": {1.10, 4.40},
	"o4-mini": {1.10, 4.40},
	// Gemini — current
	"gemini-3.5-flash":      {1.50, 9.0},
	"gemini-3.1-flash-lite": {0.25, 1.50},
	"gemini-2.5-pro":        {1.25, 10.0},
	"gemini-2.5-flash":      {0.30, 2.50},
	"gemini-2.5-flash-lite": {0.10, 0.40},
	"gemini-2.0-flash":      {0.10, 0.40},
	// Gemini — legacy
	"gemini-1.5-pro":   {1.25, 5.0},
	"gemini-1.5-flash": {0.075, 0.30},
}

// Scheduler
const defaultScheduleIntervalH = 168 // 1 week in hours

// Audit generation
const defaultMaxTokens = 8192 // max output tokens for all AI providers

// PR Guard
const (
	defaultBlastRadiusLimit = 200 // max changed lines (additions+deletions) before hard-wall block

	strikeLabelPrefix = "scout-strike:"
	maxStrikes        = 3
	epicApprovedLabel = "scout-epic-approved"

	guardCommentMarker = "<!-- ossf-scout-pr-guard:do-not-edit -->"

	truncGuardPatchPerFile = 4_000   // per-file patch cap in triage prompt
	guardMaxPromptChars    = 200_000 // overall triage prompt cap
	truncGuardCheckDetail  = 4_000   // per-check detail cap in feedback comment

	// Guard exit codes — consumed by the CI workflow step to decide how to react.
	guardExitApproved       = 0  // hard wall + triage both passed
	guardExitStrikeApplied  = 1  // hard-wall failure, strike 1 or 2, PR still open
	guardExitClosed         = 2  // strike 3 reached — PR force-closed, terminal
	guardExitTriageRejected = 3  // hard wall passed, triage rejected, strike applied
	guardExitInfraError     = 10 // guard itself failed (GitHub/Ollama unreachable, bad args)
)

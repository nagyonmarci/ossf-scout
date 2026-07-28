package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg := config{}
	var serve bool
	var port int
	var dbPath string

	flag.StringVar(&cfg.language, "lang", "", "Filter by language (e.g. go, python, java)")
	flag.IntVar(&cfg.minStars, "min-stars", defaultMinStars, "Minimum stars")
	flag.Float64Var(&cfg.maxScore, "max-score", defaultMaxScore, "Maximum OpenSSF score to include (0-10)")
	flag.IntVar(&cfg.limit, "limit", defaultLimit, "Max repos to fetch from GitHub search")
	flag.IntVar(&cfg.workers, "workers", defaultWorkers, "Concurrent Scorecard API workers")
	flag.BoolVar(&cfg.jsonOut, "json", false, "Output as JSON")
	flag.StringVar(&cfg.token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub PAT (or set GITHUB_TOKEN env)")
	flag.StringVar(&cfg.checkFilter, "checks", "", "Comma-separated Scorecard checks to highlight (default: security set)")
	flag.BoolVar(&cfg.cliFallback, "cli-fallback", false, "Use local scorecard CLI for repos not in the Scorecard database")
	flag.StringVar(&cfg.pushedAfter, "pushed-after", "", "Only include repos pushed after this date (YYYY-MM-DD)")
	flag.IntVar(&cfg.minMaintained, "min-maintained", 0, "Exclude repos where Scorecard Maintained check score is below this (0 = disabled)")
	flag.StringVar(&cfg.topic, "topic", "", "Filter by GitHub topic (e.g. ai, machine-learning)")
	flag.StringVar(&cfg.keyword, "keyword", "", "Keyword search in repo name/description")
	flag.StringVar(&cfg.singleRepo, "repo", "", "Scan a single specific repo (owner/repo), skips GitHub search")
	flag.BoolVar(&serve, "serve", false, "Start web server mode")
	flag.IntVar(&port, "port", defaultPort, "Port for web server mode")
	flag.StringVar(&dbPath, "db", "ossf-scout.db", "SQLite database path")

	var guard bool
	var prNumber int
	var guardRepo string
	var guardToken string
	var guardOllamaURL string
	var guardOllamaModel string
	var blastRadiusLimit int
	var guardDryRun bool
	flag.BoolVar(&guard, "guard", false, "Run AI Agent PR Guard mode (CI gatekeeper)")
	flag.IntVar(&prNumber, "pr", 0, "Pull request number to evaluate (required with -guard)")
	flag.StringVar(&guardRepo, "guard-repo", "", "owner/repo for PR Guard mode (required with -guard)")
	flag.StringVar(&guardToken, "guard-token", os.Getenv("GIT_TOKEN"), "GitHub token for PR Guard mode (or GIT_TOKEN env; falls back to -token/GITHUB_TOKEN)")
	flag.StringVar(&guardOllamaURL, "guard-ollama-url", defaultOllamaURL, "Ollama URL for the PR Guard triage step")
	flag.StringVar(&guardOllamaModel, "guard-ollama-model", "", "Ollama model for the PR Guard triage step (required with -guard)")
	flag.IntVar(&blastRadiusLimit, "guard-blast-radius", defaultBlastRadiusLimit, "Max changed lines before hard-wall block (unless scout-epic-approved)")
	flag.BoolVar(&guardDryRun, "guard-dry-run", false, "Evaluate and log only — no label/comment/close writes to GitHub")
	flag.Parse()

	if serve {
		startServer(port, dbPath, cfg.token)
		return
	}

	if guard {
		token := guardToken
		if token == "" {
			token = cfg.token
		}
		os.Exit(runGuard(guardConfig{
			Repo:             guardRepo,
			PRNumber:         prNumber,
			Token:            token,
			OllamaURL:        guardOllamaURL,
			OllamaModel:      guardOllamaModel,
			BlastRadiusLimit: blastRadiusLimit,
			DryRun:           guardDryRun,
		}))
	}

	if cfg.token == "" {
		fmt.Fprintln(os.Stderr, "Warning: no GITHUB_TOKEN set — GitHub rate limit is 60 req/hour unauthenticated")
	}

	repos, err := fetchRepos(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Fetched %d repos — querying Scorecard API (workers=%d)...\n",
		len(repos), cfg.workers)

	results := runWorkers(repos, cfg)

	if cfg.jsonOut {
		printJSON(results)
	} else {
		printTable(results)
	}
}

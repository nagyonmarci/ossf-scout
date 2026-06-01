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
	flag.IntVar(&cfg.minStars, "min-stars", 500, "Minimum stars")
	flag.Float64Var(&cfg.maxScore, "max-score", 5.0, "Maximum OpenSSF score to include (0-10)")
	flag.IntVar(&cfg.limit, "limit", 100, "Max repos to fetch from GitHub search")
	flag.IntVar(&cfg.workers, "workers", 5, "Concurrent Scorecard API workers")
	flag.BoolVar(&cfg.jsonOut, "json", false, "Output as JSON")
	flag.StringVar(&cfg.token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub PAT (or set GITHUB_TOKEN env)")
	flag.StringVar(&cfg.checkFilter, "checks", "", "Comma-separated Scorecard checks to highlight (default: security set)")
	flag.BoolVar(&cfg.cliFallback, "cli-fallback", false, "Use local scorecard CLI for repos not in the Scorecard database")
	flag.StringVar(&cfg.pushedAfter, "pushed-after", "", "Only include repos pushed after this date (YYYY-MM-DD)")
	flag.IntVar(&cfg.minMaintained, "min-maintained", 0, "Exclude repos where Scorecard Maintained check score is below this (0 = disabled)")
	flag.StringVar(&cfg.topic, "topic", "", "Filter by GitHub topic (e.g. ai, machine-learning)")
	flag.StringVar(&cfg.keyword, "keyword", "", "Keyword search in repo name/description")
	flag.BoolVar(&serve, "serve", false, "Start web server mode")
	flag.IntVar(&port, "port", 7878, "Port for web server mode")
	flag.StringVar(&dbPath, "db", "ossf-scout.db", "SQLite database path")
	flag.Parse()

	if serve {
		startServer(port, dbPath, cfg.token)
		return
	}

	if cfg.token == "" {
		fmt.Fprintln(os.Stderr, "Warning: no GITHUB_TOKEN set — GitHub rate limit is 60 req/hour unauthenticated")
	}

	fmt.Fprintf(os.Stderr, "Searching GitHub: language=%q minStars=%d limit=%d\n",
		cfg.language, cfg.minStars, cfg.limit)

	repos, err := searchGitHub(cfg)
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

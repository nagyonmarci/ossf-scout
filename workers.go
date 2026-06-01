package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

type result struct {
	Repo         ghRepo
	Score        float64
	WeakChecks   []string
	ScorecardURL string
}

func runWorkers(repos []ghRepo, cfg config) []result {
	wantChecks := parseChecks(cfg.checkFilter)

	jobs := make(chan ghRepo, len(repos))
	results := make(chan result, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				parts := strings.SplitN(repo.FullName, "/", 2)
				if len(parts) != 2 {
					continue
				}
				sc, err := scorecardGet(parts[0], parts[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "  [warn] %s: %v\n", repo.FullName, err)
					continue
				}
				fromCLI := false
				if sc == nil {
					if cfg.cliFallback {
						sc, err = scorecardCLI(parts[0], parts[1], cfg.token)
						if err != nil {
							fmt.Fprintf(os.Stderr, "  [warn] CLI %s: %v\n", repo.FullName, err)
						} else if sc != nil {
							fromCLI = true
						}
					}
					if sc == nil {
						results <- result{
							Repo:         repo,
							Score:        -1,
							WeakChecks:   []string{"NOT_SCANNED"},
							ScorecardURL: fmt.Sprintf("https://scorecard.dev/viewer/?uri=github.com/%s", repo.FullName),
						}
						continue
					}
				}
				if cfg.minMaintained > 0 {
					if m := findCheckScore(sc.Checks, "Maintained"); m >= 0 && m < cfg.minMaintained {
						continue
					}
				}
				if cfg.singleRepo == "" && sc.Score > cfg.maxScore && len(wantChecks) == 0 {
					continue
				}
				scorecardURL := fmt.Sprintf("https://scorecard.dev/viewer/?uri=github.com/%s", repo.FullName)
				if fromCLI {
					scorecardURL = "" // repo not indexed in online database
				}
				weak := weakChecks(sc.Checks, wantChecks)
				if cfg.singleRepo != "" || sc.Score <= cfg.maxScore || len(weak) > 0 {
					results <- result{
						Repo:         repo,
						Score:        sc.Score,
						WeakChecks:   weak,
						ScorecardURL: scorecardURL,
					}
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}

	for _, r := range repos {
		jobs <- r
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []result
	for r := range results {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		si, sj := out[i].Score, out[j].Score
		if si == -1 {
			return false
		}
		if sj == -1 {
			return true
		}
		return si < sj
	})
	return out
}

func findCheckScore(checks []scorecardCheck, name string) int {
	for _, c := range checks {
		if c.Name == name {
			return c.Score
		}
	}
	return -1
}

func parseChecks(s string) map[string]bool {
	m := make(map[string]bool)
	if s == "" {
		return m
	}
	for _, c := range strings.Split(s, ",") {
		m[strings.TrimSpace(c)] = true
	}
	return m
}

func weakChecks(checks []scorecardCheck, want map[string]bool) []string {
	securityChecks := map[string]bool{
		"CI-Tests":               true,
		"SAST":                   true,
		"Dependency-Update-Tool": true,
		"Vulnerabilities":        true,
		"Pinned-Dependencies":    true,
		"Branch-Protection":      true,
		"Code-Review":            true,
		"Maintained":             true,
	}

	filter := securityChecks
	if len(want) > 0 {
		filter = want
	}

	var weak []string
	for _, c := range checks {
		if filter[c.Name] && (c.Score == -1 || c.Score < 5) {
			weak = append(weak, fmt.Sprintf("%s(%d)", c.Name, c.Score))
		}
	}
	return weak
}

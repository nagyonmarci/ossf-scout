package scout

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Result is a single repo's Scorecard evaluation outcome.
type Result struct {
	Repo         Repo
	Score        float64
	WeakChecks   []string
	ScorecardURL string
}

// Evaluate runs each repo through the Scorecard API (with optional CLI fallback) and
// returns the ones matching Config's score/check filters.
func (s *Scout) Evaluate(ctx context.Context, repos []Repo) []Result {
	cfg := s.cfg
	wantChecks := parseChecks(cfg.CheckFilter)

	jobs := make(chan Repo, len(repos))
	results := make(chan Result, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				if ctx.Err() != nil {
					continue
				}
				parts := strings.SplitN(repo.FullName, "/", 2)
				if len(parts) != 2 {
					continue
				}
				sc, err := scorecardGet(ctx, parts[0], parts[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "  [warn] %s: %v\n", repo.FullName, err)
					continue
				}
				fromCLI := false
				if sc == nil {
					if cfg.CliFallback {
						sc, err = scorecardCLI(ctx, parts[0], parts[1], cfg.Token)
						if err != nil {
							fmt.Fprintf(os.Stderr, "  [warn] CLI %s: %v\n", repo.FullName, err)
						} else if sc != nil {
							fromCLI = true
						}
					}
					if sc == nil {
						results <- Result{
							Repo:         repo,
							Score:        -1,
							WeakChecks:   []string{"NOT_SCANNED"},
							ScorecardURL: fmt.Sprintf("https://scorecard.dev/viewer/?uri=github.com/%s", repo.FullName),
						}
						continue
					}
				}
				if cfg.MinMaintained > 0 {
					if m := findCheckScore(sc.Checks, "Maintained"); m >= 0 && m < cfg.MinMaintained {
						continue
					}
				}
				if cfg.SingleRepo == "" && sc.Score > cfg.MaxScore && len(wantChecks) == 0 {
					continue
				}
				scorecardURL := fmt.Sprintf("https://scorecard.dev/viewer/?uri=github.com/%s", repo.FullName)
				if fromCLI {
					scorecardURL = "" // repo not indexed in online database
				}
				weak := weakChecks(sc.Checks, wantChecks)
				if cfg.SingleRepo != "" || sc.Score <= cfg.MaxScore || len(weak) > 0 {
					results <- Result{
						Repo:         repo,
						Score:        sc.Score,
						WeakChecks:   weak,
						ScorecardURL: scorecardURL,
					}
				}
				select {
				case <-ctx.Done():
				case <-time.After(200 * time.Millisecond):
				}
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

	var out []Result
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

// Package scout implements ossf-scout's repository discovery and Scorecard evaluation core.
package scout

import (
	"context"
	"time"
)

const (
	gitHubAPIVersion = "2022-11-28"
	defaultPageSize  = 100 // GitHub search per_page
)

var (
	searchPageDelay     = 500 * time.Millisecond
	scorecardCLITimeout = 10 * time.Minute
)

// Config holds the parameters for a single scan run.
type Config struct {
	Language      string
	MinStars      int
	MaxScore      float64
	Limit         int
	Workers       int
	Token         string
	CheckFilter   string
	CliFallback   bool
	PushedAfter   string
	MinMaintained int
	Topic         string
	Keyword       string
	SingleRepo    string
}

// Scout runs repository discovery and Scorecard evaluation for a given Config.
type Scout struct {
	cfg Config
}

// New creates a Scout for the given Config.
func New(cfg Config) *Scout {
	return &Scout{cfg: cfg}
}

// Scan fetches matching repos and evaluates them against Scorecard, in one call.
func (s *Scout) Scan(ctx context.Context) ([]Result, error) {
	repos, err := s.FetchRepos(ctx)
	if err != nil {
		return nil, err
	}
	return s.Evaluate(ctx, repos), nil
}

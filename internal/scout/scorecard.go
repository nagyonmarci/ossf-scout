package scout

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type scorecardResponse struct {
	Score  float64          `json:"score"`
	Checks []scorecardCheck `json:"checks"`
}

type scorecardCheck struct {
	Name  string `json:"name"`
	Score int    `json:"score"` // -1 = N/A, 0–10
}

// scorecardAPIBase is overridden in tests to point at an httptest.Server.
var scorecardAPIBase = "https://api.securityscorecards.dev"

func scorecardGet(ctx context.Context, owner, repo string) (*scorecardResponse, error) {
	apiURL := fmt.Sprintf(scorecardAPIBase+"/projects/github.com/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scorecard %d", resp.StatusCode)
	}
	var sc scorecardResponse
	if err := json.NewDecoder(resp.Body).Decode(&sc); err != nil {
		return nil, err
	}
	return &sc, nil
}

func scorecardCLI(ctx context.Context, owner, repo, token string) (*scorecardResponse, error) {
	cctx, cancel := context.WithTimeout(ctx, scorecardCLITimeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, "scorecard",
		"--repo=github.com/"+owner+"/"+repo,
		"--format=json",
	)
	// scorecard needs a GitHub token and nothing else from our environment.
	// Fall back to the env tokens it would otherwise have inherited.
	cmd.Env = MinimalEnv()
	if token = cmp.Or(token, os.Getenv("GITHUB_AUTH_TOKEN"), os.Getenv("GITHUB_TOKEN")); token != "" {
		cmd.Env = append(cmd.Env, "GITHUB_TOKEN="+token, "GITHUB_AUTH_TOKEN="+token)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	if runErr != nil && errors.Is(runErr, exec.ErrNotFound) {
		return nil, nil
	}

	// scorecard exits with status 1 when any check fails (e.g. token scope issues),
	// but still writes valid JSON to stdout — use it if we can parse it
	if out := stdout.String(); out != "" {
		var sc scorecardResponse
		if err := json.Unmarshal([]byte(out), &sc); err == nil {
			return &sc, nil
		}
	}

	if runErr != nil {
		msg := runErr.Error()
		if s := strings.TrimSpace(stderr.String()); s != "" {
			msg += ": " + s
		}
		return nil, fmt.Errorf("scorecard CLI: %s", msg)
	}
	return nil, fmt.Errorf("scorecard CLI: empty output")
}

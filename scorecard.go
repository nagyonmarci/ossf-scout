package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type scorecardResponse struct {
	Score  float64          `json:"score"`
	Checks []scorecardCheck `json:"checks"`
}

type scorecardCheck struct {
	Name  string `json:"name"`
	Score int    `json:"score"` // -1 = N/A, 0–10
}

func scorecardGet(owner, repo string) (*scorecardResponse, error) {
	apiURL := fmt.Sprintf("https://api.securityscorecards.dev/projects/github.com/%s/%s", owner, repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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

func scorecardCLI(owner, repo, token string) (*scorecardResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scorecard",
		"--repo=github.com/"+owner+"/"+repo,
		"--format=json",
	)
	// Only override GITHUB_TOKEN if we have one — don't clobber an existing env token with an empty string
	if token != "" {
		cmd.Env = append(os.Environ(), "GITHUB_TOKEN="+token, "GITHUB_AUTH_TOKEN="+token)
	} else {
		cmd.Env = os.Environ()
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

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

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, nil
		}
		msg := err.Error()
		if s := strings.TrimSpace(stderr.String()); s != "" {
			msg += ": " + s
		}
		return nil, fmt.Errorf("scorecard CLI: %s", msg)
	}

	var sc scorecardResponse
	if err := json.Unmarshal(out, &sc); err != nil {
		return nil, fmt.Errorf("scorecard CLI output parse: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}
	return &sc, nil
}

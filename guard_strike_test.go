package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestHasLabel(t *testing.T) {
	labels := []ghLabel{{Name: "bug"}, {Name: epicApprovedLabel}}
	if !hasLabel(labels, epicApprovedLabel) {
		t.Error("expected label to be found")
	}
	if hasLabel(labels, "missing") {
		t.Error("expected label to be absent")
	}
}

func TestCurrentStrikeCount(t *testing.T) {
	cases := []struct {
		name   string
		labels []ghLabel
		want   int
	}{
		{"none", nil, 0},
		{"single", []ghLabel{{Name: "scout-strike:2"}}, 2},
		{"ignores unrelated", []ghLabel{{Name: "bug"}, {Name: "scout-strike:1"}}, 1},
		{"takes max on duplicates", []ghLabel{{Name: "scout-strike:1"}, {Name: "scout-strike:3"}}, 3},
	}
	for _, c := range cases {
		if got := currentStrikeCount(c.labels); got != c.want {
			t.Errorf("%s: currentStrikeCount = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestApplyStrikeIdempotent(t *testing.T) {
	labels := map[string]bool{"scout-strike:1": true}
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var payload struct {
				Labels []string `json:"labels"`
			}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			for _, l := range payload.Labels {
				labels[l] = true
			}
			w.WriteHeader(http.StatusOK)
		case "DELETE":
			parts := strings.Split(r.URL.Path, "/")
			delete(labels, parts[len(parts)-1])
			w.WriteHeader(http.StatusOK)
		}
	})

	cfg := guardConfig{Repo: "o/r", PRNumber: 5, Token: "tok"}
	pr := &ghPullRequest{Labels: []ghLabel{{Name: "scout-strike:1"}}}

	newCount, err := applyStrike(cfg, pr)
	if err != nil {
		t.Fatalf("applyStrike: %v", err)
	}
	if newCount != 2 {
		t.Errorf("newCount = %d, want 2", newCount)
	}
	if labels["scout-strike:1"] {
		t.Error("old strike label should have been removed")
	}
	if !labels["scout-strike:2"] {
		t.Error("new strike label should have been added")
	}

	// Re-running with the OLD label state (simulating a retried CI job that
	// hasn't refetched the PR yet) must still converge on strike 2, not stack.
	newCount2, err := applyStrike(cfg, pr)
	if err != nil {
		t.Fatalf("applyStrike (retry): %v", err)
	}
	if newCount2 != 2 {
		t.Errorf("retry newCount = %d, want 2 (idempotent)", newCount2)
	}
}

func TestApplyStrikeDryRunMakesNoRequests(t *testing.T) {
	called := false
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	cfg := guardConfig{Repo: "o/r", PRNumber: 5, Token: "tok", DryRun: true}
	pr := &ghPullRequest{}
	newCount, err := applyStrike(cfg, pr)
	if err != nil {
		t.Fatalf("applyStrike: %v", err)
	}
	if newCount != 1 {
		t.Errorf("newCount = %d, want 1", newCount)
	}
	if called {
		t.Error("dry-run should not make any GitHub API requests")
	}
}

func TestEnforceStrikeLimitClosesAtMaxStrikes(t *testing.T) {
	var gotMethod, gotPath string
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	cfg := guardConfig{Repo: "o/r", PRNumber: 9, Token: "tok"}
	pr := &ghPullRequest{Labels: []ghLabel{{Name: "scout-strike:3"}}}

	terminated, err := enforceStrikeLimit(cfg, pr)
	if err != nil {
		t.Fatalf("enforceStrikeLimit: %v", err)
	}
	if !terminated {
		t.Error("expected PR to be terminated at strike limit")
	}
	if gotMethod != "PATCH" || gotPath != "/repos/o/r/pulls/9" {
		t.Errorf("expected closePR call, got %s %s", gotMethod, gotPath)
	}
}

func TestEnforceStrikeLimitNotYetReached(t *testing.T) {
	cfg := guardConfig{Repo: "o/r", PRNumber: 9, Token: "tok"}
	pr := &ghPullRequest{Labels: []ghLabel{{Name: "scout-strike:1"}}}

	terminated, err := enforceStrikeLimit(cfg, pr)
	if err != nil {
		t.Fatalf("enforceStrikeLimit: %v", err)
	}
	if terminated {
		t.Error("PR below strike limit should not be terminated")
	}
}

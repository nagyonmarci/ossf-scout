package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func stubCloneHead(t *testing.T) {
	t.Helper()
	orig := cloneHeadFn
	cloneHeadFn = func(cfg guardConfig, pr *ghPullRequest) (string, func(), error) {
		dir, err := os.MkdirTemp("", "guard-test-*")
		if err != nil {
			return "", nil, err
		}
		return dir, func() { os.RemoveAll(dir) }, nil
	}
	t.Cleanup(func() { cloneHeadFn = orig })
}

// guardTestState models a fake GitHub PR's mutable state (labels, comments,
// open/closed) behind a single mux, driving runGuard through full scenarios
// without touching the real API.
type guardTestState struct {
	labels  []string
	closed  bool
	comment string
}

func newGuardMux(t *testing.T, state *guardTestState, pr *ghPullRequest, files []ghPRFile) {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/o/r/pulls/9", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" {
			state.closed = true
			w.WriteHeader(http.StatusOK)
			return
		}
		labels := make([]ghLabel, len(state.labels))
		for i, l := range state.labels {
			labels[i] = ghLabel{Name: l}
		}
		p := *pr
		p.Labels = labels
		_ = json.NewEncoder(w).Encode(p)
	})
	mux.HandleFunc("/repos/o/r/pulls/9/files", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "" && r.URL.Query().Get("page") != "1" {
			_ = json.NewEncoder(w).Encode([]ghPRFile{})
			return
		}
		_ = json.NewEncoder(w).Encode(files)
	})
	mux.HandleFunc("/repos/o/r/commits/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"commit":{"verification":{"verified":true,"reason":"valid"}}}`))
	})
	mux.HandleFunc("/repos/o/r/issues/9/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var payload struct {
				Body string `json:"body"`
			}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			state.comment = payload.Body
			_, _ = w.Write([]byte(`{"id":1}`))
			return
		}
		_ = json.NewEncoder(w).Encode([]ghComment{})
	})
	mux.HandleFunc("/repos/o/r/issues/9/labels", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Labels []string `json:"labels"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		state.labels = append(state.labels, payload.Labels...)
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/repos/o/r/issues/9/labels/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	orig := githubAPIBase
	githubAPIBase = srv.URL
	t.Cleanup(func() { githubAPIBase = orig })
}

func TestRunGuardInvalidArgs(t *testing.T) {
	if code := runGuard(guardConfig{Repo: "not-a-valid-repo", PRNumber: 1}); code != guardExitInfraError {
		t.Errorf("invalid repo: code = %d, want %d", code, guardExitInfraError)
	}
	if code := runGuard(guardConfig{Repo: "o/r", PRNumber: 0}); code != guardExitInfraError {
		t.Errorf("invalid PR number: code = %d, want %d", code, guardExitInfraError)
	}
}

func TestRunGuardCleanApprove(t *testing.T) {
	stubCloneHead(t)
	state := &guardTestState{}
	pr := &ghPullRequest{Number: 9, Additions: 5, Deletions: 2}
	pr.Head.SHA = "abc"
	pr.Head.Ref = "feat"
	newGuardMux(t, state, pr, nil) // no changed files -> semgrep/trivy skip cleanly

	triageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fw := w.(http.Flusher)
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"{\"decision\":\"approve\",\"reason\":\"looks good\"}"}}]}` + "\n"))
		fw.Flush()
		_, _ = w.Write([]byte("data: [DONE]\n"))
		fw.Flush()
	}))
	defer triageSrv.Close()

	cfg := guardConfig{Repo: "o/r", PRNumber: 9, Token: "tok", BlastRadiusLimit: 200, OllamaURL: triageSrv.URL, OllamaModel: "llama3.1"}
	code := runGuard(cfg)
	if code != guardExitApproved {
		t.Errorf("code = %d, want %d (approved)", code, guardExitApproved)
	}
	if state.closed {
		t.Error("PR should not be closed on approval")
	}
}

func TestRunGuardHardWallFailureAppliesStrike(t *testing.T) {
	stubCloneHead(t)
	state := &guardTestState{}
	pr := &ghPullRequest{Number: 9, Additions: 500, Deletions: 500} // over blast radius
	pr.Head.SHA = "abc"
	pr.Head.Ref = "feat"
	newGuardMux(t, state, pr, nil)

	cfg := guardConfig{Repo: "o/r", PRNumber: 9, Token: "tok", BlastRadiusLimit: 200}
	code := runGuard(cfg)
	if code != guardExitStrikeApplied {
		t.Errorf("code = %d, want %d (strike applied)", code, guardExitStrikeApplied)
	}
	if len(state.labels) != 1 || state.labels[0] != "scout-strike:1" {
		t.Errorf("labels = %v, want [scout-strike:1]", state.labels)
	}
	if state.closed {
		t.Error("PR should not be closed on strike 1")
	}
	if state.comment == "" {
		t.Error("expected a feedback comment to be posted")
	}
}

func TestRunGuardThirdStrikeCloses(t *testing.T) {
	stubCloneHead(t)
	state := &guardTestState{labels: []string{"scout-strike:2"}}
	pr := &ghPullRequest{Number: 9, Additions: 500, Deletions: 500} // over blast radius, fails again
	pr.Head.SHA = "abc"
	pr.Head.Ref = "feat"
	newGuardMux(t, state, pr, nil)

	cfg := guardConfig{Repo: "o/r", PRNumber: 9, Token: "tok", BlastRadiusLimit: 200}
	code := runGuard(cfg)
	if code != guardExitClosed {
		t.Errorf("code = %d, want %d (closed)", code, guardExitClosed)
	}
	if !state.closed {
		t.Error("expected PR to be force-closed at strike 3")
	}
}

func TestRunGuardAlreadyAtStrikeLimitClosesWithoutRescanning(t *testing.T) {
	state := &guardTestState{labels: []string{"scout-strike:3"}}
	pr := &ghPullRequest{Number: 9}
	pr.Head.SHA = "abc"
	pr.Head.Ref = "feat"
	newGuardMux(t, state, pr, nil)
	// Deliberately do NOT stub cloneHead — if runGuard tried to scan, this test
	// would hang/fail attempting a real network clone, proving enforceStrikeLimit
	// short-circuits before that point.

	cfg := guardConfig{Repo: "o/r", PRNumber: 9, Token: "tok", BlastRadiusLimit: 200}
	code := runGuard(cfg)
	if code != guardExitClosed {
		t.Errorf("code = %d, want %d (closed)", code, guardExitClosed)
	}
	if !state.closed {
		t.Error("expected PR to be closed")
	}
}

package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestFindGuardCommentMatchesMarker(t *testing.T) {
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1,"body":"hi"},{"id":2,"body":"` + guardCommentMarker + `\nstuff"}]`))
	})

	id, found, err := findGuardComment("o/r", 5, "tok")
	if err != nil {
		t.Fatalf("findGuardComment: %v", err)
	}
	if !found || id != 2 {
		t.Errorf("id=%d found=%v, want id=2 found=true", id, found)
	}
}

func TestFindGuardCommentNoneFound(t *testing.T) {
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1,"body":"unrelated"}]`))
	})

	_, found, err := findGuardComment("o/r", 5, "tok")
	if err != nil {
		t.Fatalf("findGuardComment: %v", err)
	}
	if found {
		t.Error("expected no guard comment to be found")
	}
}

func TestUpsertGuardCommentPatchesExisting(t *testing.T) {
	var gotMethod, gotPath string
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte(`[{"id":7,"body":"` + guardCommentMarker + `"}]`))
			return
		}
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	cfg := guardConfig{Repo: "o/r", Token: "tok"}
	if err := upsertGuardComment(cfg, 5, "new body"); err != nil {
		t.Fatalf("upsertGuardComment: %v", err)
	}
	if gotMethod != "PATCH" || gotPath != "/repos/o/r/issues/comments/7" {
		t.Errorf("expected PATCH to existing comment, got %s %s", gotMethod, gotPath)
	}
}

func TestUpsertGuardCommentPostsWhenNoneExists(t *testing.T) {
	var gotMethod string
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"id":1}`))
	})

	cfg := guardConfig{Repo: "o/r", Token: "tok"}
	if err := upsertGuardComment(cfg, 5, "new body"); err != nil {
		t.Fatalf("upsertGuardComment: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST for a new comment, got %s", gotMethod)
	}
}

func TestUpsertGuardCommentDryRunMakesNoRequests(t *testing.T) {
	called := false
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := guardConfig{Repo: "o/r", Token: "tok", DryRun: true}
	if err := upsertGuardComment(cfg, 5, "body"); err != nil {
		t.Fatalf("upsertGuardComment: %v", err)
	}
	if called {
		t.Error("dry-run should not make any GitHub API requests")
	}
}

func TestRenderFeedbackCommentIncludesMarkerAndPolicies(t *testing.T) {
	failures := []hardWallFailure{{Check: "trivy", Policy: "no CVEs", Detail: "CVE-2024-1"}}
	body := renderFeedbackComment(1, failures, nil)
	if !strings.Contains(body, guardCommentMarker) {
		t.Error("expected marker in comment body")
	}
	if !strings.Contains(body, "trivy") || !strings.Contains(body, "CVE-2024-1") {
		t.Error("expected failure details in comment body")
	}
	if !strings.Contains(body, "2 attempt(s) remain") {
		t.Errorf("expected remaining-attempts note, got: %s", body)
	}
}

func TestRenderClosedCommentIncludesMarker(t *testing.T) {
	body := renderClosedComment(nil, &triageDecision{Decision: "reject", Reason: "noise"})
	if !strings.Contains(body, guardCommentMarker) {
		t.Error("expected marker in comment body")
	}
	if !strings.Contains(body, "Closed") {
		t.Error("expected closed-state note in comment body")
	}
}

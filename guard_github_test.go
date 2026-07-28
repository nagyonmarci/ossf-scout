package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withGuardGitHubServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	orig := githubAPIBase
	githubAPIBase = srv.URL
	t.Cleanup(func() { githubAPIBase = orig })
	return srv
}

func TestGhDoSendsMethodAndBody(t *testing.T) {
	var gotMethod, gotBody, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		buf, _ := io.ReadAll(r.Body)
		gotBody = string(buf)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, status, err := ghDo("POST", srv.URL, "tok", []byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("ghDo: %v", err)
	}
	if status != http.StatusCreated {
		t.Errorf("status = %d, want 201", status)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q", body)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
	if gotBody != `{"a":1}` {
		t.Errorf("request body = %q", gotBody)
	}
}

func TestGhDeleteTreats404AsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	if err := ghDelete(srv.URL, "tok"); err != nil {
		t.Errorf("ghDelete should treat 404 as success, got %v", err)
	}
}

func TestGhDeletePropagatesOtherErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if err := ghDelete(srv.URL, "tok"); err == nil {
		t.Error("ghDelete should propagate non-404 errors")
	}
}

func TestGetPullRequestParsesFields(t *testing.T) {
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/pulls/5" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"number":5,"additions":10,"deletions":3,"labels":[{"name":"scout-strike:1"}],"head":{"sha":"abc","ref":"feat"}}`))
	})

	pr, err := getPullRequest("o/r", 5, "tok")
	if err != nil {
		t.Fatalf("getPullRequest: %v", err)
	}
	if pr.Additions != 10 || pr.Deletions != 3 || pr.Head.SHA != "abc" || pr.Head.Ref != "feat" {
		t.Errorf("unexpected pr: %+v", pr)
	}
	if len(pr.Labels) != 1 || pr.Labels[0].Name != "scout-strike:1" {
		t.Errorf("unexpected labels: %+v", pr.Labels)
	}
}

func TestGetPullRequestFilesPaginates(t *testing.T) {
	calls := 0
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("page") == "1" {
			files := make([]ghPRFile, defaultPageSize)
			for i := range files {
				files[i] = ghPRFile{Filename: "f"}
			}
			_ = json.NewEncoder(w).Encode(files)
			return
		}
		_ = json.NewEncoder(w).Encode([]ghPRFile{{Filename: "last.go"}})
	})

	files, err := getPullRequestFiles("o/r", 5, "tok")
	if err != nil {
		t.Fatalf("getPullRequestFiles: %v", err)
	}
	if len(files) != defaultPageSize+1 {
		t.Errorf("got %d files, want %d", len(files), defaultPageSize+1)
	}
	if calls != 2 {
		t.Errorf("expected 2 pages fetched, got %d calls", calls)
	}
}

func TestGetCommitVerification(t *testing.T) {
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/commits/abc123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"commit":{"verification":{"verified":true,"reason":"valid"}}}`))
	})

	verified, reason, err := getCommitVerification("o/r", "abc123", "tok")
	if err != nil {
		t.Fatalf("getCommitVerification: %v", err)
	}
	if !verified || reason != "valid" {
		t.Errorf("verified=%v reason=%q", verified, reason)
	}
}

func TestAddLabelSendsCorrectBody(t *testing.T) {
	var gotPath, gotBody string
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		gotBody = string(buf)
		w.WriteHeader(http.StatusOK)
	})

	if err := addLabel("o/r", 5, "scout-strike:1", "tok"); err != nil {
		t.Fatalf("addLabel: %v", err)
	}
	if gotPath != "/repos/o/r/issues/5/labels" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody != `{"labels":["scout-strike:1"]}` {
		t.Errorf("body = %q", gotBody)
	}
}

func TestRemoveLabelEscapesColon(t *testing.T) {
	var gotPath string
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	if err := removeLabel("o/r", 5, "scout-strike:1", "tok"); err != nil {
		t.Fatalf("removeLabel: %v", err)
	}
	if gotPath != "/repos/o/r/issues/5/labels/scout-strike:1" && gotPath != "/repos/o/r/issues/5/labels/scout-strike%3A1" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestPostAndPatchComment(t *testing.T) {
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/repos/o/r/issues/5/comments":
			_, _ = w.Write([]byte(`{"id":42}`))
		case r.Method == "PATCH" && r.URL.Path == "/repos/o/r/issues/comments/42":
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	id, err := postComment("o/r", 5, "hello", "tok")
	if err != nil {
		t.Fatalf("postComment: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
	if err := patchComment("o/r", id, "updated", "tok"); err != nil {
		t.Fatalf("patchComment: %v", err)
	}
}

func TestClosePRSendsPatchWithClosedState(t *testing.T) {
	var gotMethod, gotBody string
	withGuardGitHubServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		buf, _ := io.ReadAll(r.Body)
		gotBody = string(buf)
		w.WriteHeader(http.StatusOK)
	})

	if err := closePR("o/r", 5, "tok"); err != nil {
		t.Fatalf("closePR: %v", err)
	}
	if gotMethod != "PATCH" {
		t.Errorf("method = %q", gotMethod)
	}
	if gotBody != `{"state":"closed"}` {
		t.Errorf("body = %q", gotBody)
	}
}

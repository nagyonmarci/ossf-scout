package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGhGetSuccessSendsHeaders(t *testing.T) {
	var gotAccept, gotAuth, gotAPIVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		gotAuth = r.Header.Get("Authorization")
		gotAPIVersion = r.Header.Get("X-GitHub-Api-Version")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, err := ghGet(srv.URL, "tok123")
	if err != nil {
		t.Fatalf("ghGet: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q", body)
	}
	if gotAccept != "application/vnd.github+json" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotAPIVersion != "2022-11-28" {
		t.Errorf("X-GitHub-Api-Version = %q", gotAPIVersion)
	}
	if gotAuth != "Bearer tok123" {
		t.Errorf("Authorization = %q, want Bearer tok123", gotAuth)
	}
}

func TestGhGetNoTokenOmitsAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	if _, err := ghGet(srv.URL, ""); err != nil {
		t.Fatalf("ghGet: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization should be empty without a token, got %q", gotAuth)
	}
}

func TestGhGetNon200IncludesStatusAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("API rate limit exceeded"))
	}))
	defer srv.Close()

	_, err := ghGet(srv.URL, "")
	if err == nil {
		t.Fatal("expected an error for HTTP 403")
	}
	if !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("error should include status and body, got: %v", err)
	}
}

func TestGhGetBadURL(t *testing.T) {
	if _, err := ghGet("http://bad host/with space", ""); err == nil {
		t.Error("expected an error for a malformed URL")
	}
}

func TestFetchReposInvalidSingleRepo(t *testing.T) {
	// Clearly invalid shapes must error before any network call is made.
	for _, bad := range []string{"bad", "owner/", "/repo"} {
		_, err := fetchRepos(config{singleRepo: bad})
		if err == nil {
			t.Errorf("fetchRepos(singleRepo=%q): expected an invalid-format error", bad)
			continue
		}
		if !strings.Contains(err.Error(), "invalid repo format") {
			t.Errorf("fetchRepos(singleRepo=%q): error = %v", bad, err)
		}
	}
}

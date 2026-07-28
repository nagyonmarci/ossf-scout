package main

import (
	"errors"
	"testing"
)

func TestCheckBlastRadius(t *testing.T) {
	cfg := guardConfig{BlastRadiusLimit: 200}

	within := &ghPullRequest{Additions: 100, Deletions: 50}
	if f := checkBlastRadius(cfg, within); f != nil {
		t.Errorf("expected no failure within limit, got %+v", f)
	}

	over := &ghPullRequest{Additions: 150, Deletions: 100}
	if f := checkBlastRadius(cfg, over); f == nil {
		t.Error("expected failure over limit")
	}

	overButApproved := &ghPullRequest{Additions: 150, Deletions: 100, Labels: []ghLabel{{Name: epicApprovedLabel}}}
	if f := checkBlastRadius(cfg, overButApproved); f != nil {
		t.Errorf("expected epic-approved label to bypass limit, got %+v", f)
	}
}

func TestCheckProvenance(t *testing.T) {
	cfg := guardConfig{Repo: "o/r"}
	pr := &ghPullRequest{}
	pr.Head.SHA = "abc123"

	verifiedFn := func(repo, sha, token string) (bool, string, error) { return true, "valid", nil }
	if f := checkProvenance(cfg, pr, verifiedFn); f != nil {
		t.Errorf("expected no failure for verified commit, got %+v", f)
	}

	unverifiedFn := func(repo, sha, token string) (bool, string, error) { return false, "unsigned", nil }
	if f := checkProvenance(cfg, pr, unverifiedFn); f == nil {
		t.Error("expected failure for unverified commit")
	}

	errFn := func(repo, sha, token string) (bool, string, error) { return false, "", errors.New("network down") }
	if f := checkProvenance(cfg, pr, errFn); f == nil {
		t.Error("expected failure when verification lookup errors")
	}
}

func TestChangedFilePathsSkipsRemoved(t *testing.T) {
	files := []ghPRFile{
		{Filename: "a.go", Status: "added"},
		{Filename: "b.go", Status: "removed"},
		{Filename: "c.go", Status: "modified"},
	}
	got := changedFilePaths(files)
	if len(got) != 2 || got[0] != "a.go" || got[1] != "c.go" {
		t.Errorf("changedFilePaths = %v", got)
	}
}

func TestShellQuoteEscapesSingleQuotes(t *testing.T) {
	got := shellQuote("a'b")
	want := `'a'\''b'`
	if got != want {
		t.Errorf("shellQuote = %q, want %q", got, want)
	}
}

func TestToolMissing(t *testing.T) {
	cases := []struct {
		out  string
		want bool
	}{
		{"semgrep not installed", true},
		{"sh: trivy: command not found", true},
		{`{"results":[]}`, false},
	}
	for _, c := range cases {
		if got := toolMissing(c.out); got != c.want {
			t.Errorf("toolMissing(%q) = %v, want %v", c.out, got, c.want)
		}
	}
}

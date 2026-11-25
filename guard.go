package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type guardConfig struct {
	Repo             string // owner/repo
	PRNumber         int
	Token            string
	OllamaURL        string
	OllamaModel      string
	BlastRadiusLimit int
	DryRun           bool
}

// runGuard orchestrates the 4-step PR Guard funnel and returns the process exit
// code (see guardExit* constants in constants.go for the contract consumed by
// the CI workflow step).
func runGuard(cfg guardConfig) int {
	if _, _, ok := splitValidRepo(cfg.Repo); !ok {
		fmt.Fprintln(os.Stderr, "guard: invalid -guard-repo, expected owner/repo")
		return guardExitInfraError
	}
	if cfg.PRNumber <= 0 {
		fmt.Fprintln(os.Stderr, "guard: -pr is required and must be a positive PR number")
		return guardExitInfraError
	}

	pr, err := getPullRequest(cfg.Repo, cfg.PRNumber, cfg.Token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "guard: fetch PR:", err)
		return guardExitInfraError
	}

	// A PR already at the strike limit is closed immediately, without re-scanning.
	if terminated, err := enforceStrikeLimit(cfg, pr); err != nil {
		fmt.Fprintln(os.Stderr, "guard:", err)
		return guardExitInfraError
	} else if terminated {
		fmt.Fprintln(os.Stderr, "guard: PR already at strike limit — closed")
		return guardExitClosed
	}

	files, err := getPullRequestFiles(cfg.Repo, cfg.PRNumber, cfg.Token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "guard: fetch PR files:", err)
		return guardExitInfraError
	}

	repoDir, cleanup, err := cloneHeadFn(cfg, pr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "guard: clone PR head:", err)
		return guardExitInfraError
	}
	defer cleanup()

	hardWall := runHardWall(cfg, pr, files, repoDir)
	if !hardWall.Passed {
		return rejectPR(cfg, pr, hardWall.Failures, nil)
	}

	decision, err := runTriage(cfg, pr, files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "guard: triage:", err)
		return guardExitInfraError
	}
	if decision.Decision == "reject" {
		return rejectPR(cfg, pr, nil, &decision)
	}

	return guardExitApproved
}

// rejectPR applies a strike and posts feedback, closing the PR if the strike
// limit is reached this run.
func rejectPR(cfg guardConfig, pr *ghPullRequest, failures []hardWallFailure, triage *triageDecision) int {
	strike, err := applyStrike(cfg, pr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "guard:", err)
		return guardExitInfraError
	}

	if strike >= maxStrikes {
		if !cfg.DryRun {
			if err := closePR(cfg.Repo, cfg.PRNumber, cfg.Token); err != nil {
				fmt.Fprintln(os.Stderr, "guard: close PR:", err)
				return guardExitInfraError
			}
		}
		if err := upsertGuardComment(cfg, cfg.PRNumber, renderClosedComment(failures, triage)); err != nil {
			fmt.Fprintln(os.Stderr, "guard: post closed comment:", err)
		}
		return guardExitClosed
	}

	if err := upsertGuardComment(cfg, cfg.PRNumber, renderFeedbackComment(strike, failures, triage)); err != nil {
		fmt.Fprintln(os.Stderr, "guard: post feedback comment:", err)
	}
	if len(failures) > 0 {
		return guardExitStrikeApplied
	}
	return guardExitTriageRejected
}

// cloneHeadFn is overridden in tests to avoid a real network clone.
var cloneHeadFn = cloneHead

// cloneHead shallow-clones the PR's head ref so Semgrep/Trivy can scan real
// files on disk — the GitHub API's per-file patch is truncated/omitted for
// large or binary diffs, so it isn't a reliable SAST/SCA input.
func cloneHead(cfg guardConfig, pr *ghPullRequest) (dir string, cleanup func(), err error) {
	owner, repoName, ok := splitValidRepo(cfg.Repo)
	if !ok {
		return "", nil, fmt.Errorf("invalid repository name")
	}

	tmpDir, err := os.MkdirTemp("", "ossf-guard-*")
	if err != nil {
		return "", nil, fmt.Errorf("mktemp: %w", err)
	}
	cleanup = func() { os.RemoveAll(tmpDir) } //nolint:errcheck

	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repoName)
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", pr.Head.Ref, cloneURL, tmpDir)
	cmd.Env = gitAuthEnv(cfg.Token)
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", nil, fmt.Errorf("git clone failed: %s", msg)
	}
	return tmpDir, cleanup, nil
}

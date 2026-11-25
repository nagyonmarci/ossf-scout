package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type hardWallFailure struct {
	Check  string // "semgrep" | "trivy" | "provenance" | "blast-radius"
	Policy string
	Detail string
}

type hardWallResult struct {
	Passed   bool
	Failures []hardWallFailure
}

// dependencyManifestPatterns lists filenames that trigger the Trivy SCA check
// when changed by a PR.
var dependencyManifestPatterns = []string{
	"go.mod", "go.sum", "package.json", "package-lock.json", "pnpm-lock.yaml",
	"yarn.lock", "requirements.txt", "Pipfile.lock", "poetry.lock", "Gemfile.lock",
	"composer.lock", "Cargo.lock",
}

// runHardWall runs all four checks unconditionally (no short-circuiting) so a
// single feedback comment can report every violation from one pass. repoDir is
// a shallow clone of the PR head, used by the shell-out SAST/SCA checks.
func runHardWall(cfg guardConfig, pr *ghPullRequest, files []ghPRFile, repoDir string) hardWallResult {
	var failures []hardWallFailure

	if f := checkProvenance(cfg, pr, getCommitVerification); f != nil {
		failures = append(failures, *f)
	}
	if f := checkBlastRadius(cfg, pr); f != nil {
		failures = append(failures, *f)
	}
	if f := checkSemgrep(repoDir, files); f != nil {
		failures = append(failures, *f)
	}
	if f := checkTrivy(repoDir, files); f != nil {
		failures = append(failures, *f)
	}

	return hardWallResult{Passed: len(failures) == 0, Failures: failures}
}

// checkProvenance takes the verification lookup as a parameter so tests can
// inject a fake without a live GitHub server.
func checkProvenance(cfg guardConfig, pr *ghPullRequest, verify func(repo, sha, token string) (bool, string, error)) *hardWallFailure {
	const policy = "Head commit must be GitHub-verified (signed)"
	verified, reason, err := verify(cfg.Repo, pr.Head.SHA, cfg.Token)
	if err != nil {
		return &hardWallFailure{Check: "provenance", Policy: policy,
			Detail: fmt.Sprintf("could not check commit verification: %v", err)}
	}
	if !verified {
		return &hardWallFailure{Check: "provenance", Policy: policy,
			Detail: fmt.Sprintf("head commit %s is not verified (reason: %s)", pr.Head.SHA, reason)}
	}
	return nil
}

func checkBlastRadius(cfg guardConfig, pr *ghPullRequest) *hardWallFailure {
	changed := pr.Additions + pr.Deletions
	if changed <= cfg.BlastRadiusLimit || hasLabel(pr.Labels, epicApprovedLabel) {
		return nil
	}
	return &hardWallFailure{
		Check:  "blast-radius",
		Policy: fmt.Sprintf("Diff size must not exceed %d changed lines unless labeled %q", cfg.BlastRadiusLimit, epicApprovedLabel),
		Detail: fmt.Sprintf("PR changes %d lines (%d additions + %d deletions), exceeding the limit of %d",
			changed, pr.Additions, pr.Deletions, cfg.BlastRadiusLimit),
	}
}

func changedFilePaths(files []ghPRFile) []string {
	paths := make([]string, 0, len(files))
	for _, f := range files {
		if f.Status != "removed" {
			paths = append(paths, f.Filename)
		}
	}
	return paths
}

// shellQuote single-quotes a path for safe interpolation into a /bin/sh -c
// script — changed filenames come from the GitHub API and aren't trusted input.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func quoteAll(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = shellQuote(p)
	}
	return out
}

// toolMissing reports whether a shIn fallback/stderr blob indicates the tool
// binary itself isn't installed, as opposed to the tool running and reporting
// findings. Unlike the existing audit feature (which treats a missing tool as
// an informational skip), the guard treats this as a hard block: the Docker
// image bundles semgrep/trivy/git, so a missing tool means a broken
// environment, not a soft case.
func toolMissing(out string) bool {
	return strings.Contains(out, "not installed") || strings.Contains(out, "command not found")
}

func checkSemgrep(repoDir string, files []ghPRFile) *hardWallFailure {
	const policy = "No ERROR-severity Semgrep findings in changed files"
	paths := changedFilePaths(files)
	if len(paths) == 0 {
		return nil
	}

	out := shIn(repoDir, "semgrep not installed",
		"semgrep --config=auto --json --quiet "+strings.Join(quoteAll(paths), " ")+" 2>&1")
	if toolMissing(out) {
		return &hardWallFailure{Check: "semgrep", Policy: policy,
			Detail: "semgrep is not installed in this environment — guard cannot verify SAST cleanliness"}
	}

	var result struct {
		Results []struct {
			Path  string `json:"path"`
			Start struct {
				Line int `json:"line"`
			} `json:"start"`
			CheckID string `json:"check_id"`
			Extra   struct {
				Message  string `json:"message"`
				Severity string `json:"severity"`
			} `json:"extra"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return &hardWallFailure{Check: "semgrep", Policy: policy,
			Detail: fmt.Sprintf("could not parse semgrep output: %v", err)}
	}

	var findings []string
	for _, r := range result.Results {
		if strings.EqualFold(r.Extra.Severity, "ERROR") {
			findings = append(findings, fmt.Sprintf("%s:%d [%s] %s", r.Path, r.Start.Line, r.CheckID, r.Extra.Message))
		}
	}
	if len(findings) == 0 {
		return nil
	}
	return &hardWallFailure{Check: "semgrep", Policy: policy,
		Detail: truncateField(strings.Join(findings, "\n"), truncGuardCheckDetail)}
}

func checkTrivy(repoDir string, files []ghPRFile) *hardWallFailure {
	const policy = "No CRITICAL/HIGH vulnerabilities in changed dependency manifests"

	var manifestDirs []string
	seen := map[string]bool{}
	for _, f := range files {
		if f.Status == "removed" {
			continue
		}
		base := filepath.Base(f.Filename)
		for _, pat := range dependencyManifestPatterns {
			if base != pat {
				continue
			}
			dir := filepath.Dir(f.Filename)
			if !seen[dir] {
				seen[dir] = true
				manifestDirs = append(manifestDirs, dir)
			}
		}
	}
	if len(manifestDirs) == 0 {
		return nil
	}

	out := shIn(repoDir, "trivy not installed",
		"trivy fs --severity CRITICAL,HIGH --format json --quiet "+strings.Join(quoteAll(manifestDirs), " ")+" 2>&1")
	if toolMissing(out) {
		return &hardWallFailure{Check: "trivy", Policy: policy,
			Detail: "trivy is not installed in this environment — guard cannot verify dependency safety"}
	}

	var report struct {
		Results []struct {
			Target          string `json:"Target"`
			Vulnerabilities []struct {
				VulnerabilityID string `json:"VulnerabilityID"`
				PkgName         string `json:"PkgName"`
				Severity        string `json:"Severity"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		return &hardWallFailure{Check: "trivy", Policy: policy,
			Detail: fmt.Sprintf("could not parse trivy output: %v", err)}
	}

	var findings []string
	for _, r := range report.Results {
		for _, v := range r.Vulnerabilities {
			findings = append(findings, fmt.Sprintf("%s: %s in %s (%s)", r.Target, v.VulnerabilityID, v.PkgName, v.Severity))
		}
	}
	if len(findings) == 0 {
		return nil
	}
	return &hardWallFailure{Check: "trivy", Policy: policy,
		Detail: truncateField(strings.Join(findings, "\n"), truncGuardCheckDetail)}
}

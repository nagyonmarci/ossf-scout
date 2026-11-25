package main

import (
	"fmt"
	"strconv"
	"strings"
)

func hasLabel(labels []ghLabel, name string) bool {
	for _, l := range labels {
		if l.Name == name {
			return true
		}
	}
	return false
}

// currentStrikeCount parses "scout-strike:N" labels and returns the max N found
// (0 if none). Duplicates shouldn't happen, but a race across CI retries could
// leave more than one strike label present — trust the highest, never the
// absence of a lower one.
func currentStrikeCount(labels []ghLabel) int {
	max := 0
	for _, l := range labels {
		if !strings.HasPrefix(l.Name, strikeLabelPrefix) {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimPrefix(l.Name, strikeLabelPrefix)); err == nil && n > max {
			max = n
		}
	}
	return max
}

// enforceStrikeLimit runs before any other step: if the PR already carries the
// terminal strike label, close it immediately without re-scanning. closePR is
// idempotent (closing an already-closed PR is a no-op), so a CI retry after a
// partial failure here is safe.
func enforceStrikeLimit(cfg guardConfig, pr *ghPullRequest) (terminated bool, err error) {
	if currentStrikeCount(pr.Labels) < maxStrikes {
		return false, nil
	}
	if cfg.DryRun {
		return true, nil
	}
	if err := closePR(cfg.Repo, cfg.PRNumber, cfg.Token); err != nil {
		return true, fmt.Errorf("close PR at strike limit: %w", err)
	}
	return true, nil
}

// applyStrike removes any existing scout-strike:N label before adding N+1, so a
// partially-failed prior run (label added, comment post failed) can't leave two
// strike labels stacked.
func applyStrike(cfg guardConfig, pr *ghPullRequest) (newCount int, err error) {
	current := currentStrikeCount(pr.Labels)
	newCount = current + 1

	if cfg.DryRun {
		return newCount, nil
	}
	if current > 0 {
		if err := removeLabel(cfg.Repo, cfg.PRNumber, fmt.Sprintf("%s%d", strikeLabelPrefix, current), cfg.Token); err != nil {
			return newCount, fmt.Errorf("remove old strike label: %w", err)
		}
	}
	if err := addLabel(cfg.Repo, cfg.PRNumber, fmt.Sprintf("%s%d", strikeLabelPrefix, newCount), cfg.Token); err != nil {
		return newCount, fmt.Errorf("add strike label: %w", err)
	}
	return newCount, nil
}

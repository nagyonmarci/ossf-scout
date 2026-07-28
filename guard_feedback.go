package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// findGuardComment scans PR comments for the guard's hidden marker, so repeated
// runs update one comment instead of spamming a new one each time — the same
// convention bots like Dependabot/Renovate use.
func findGuardComment(repo string, pr int, token string) (commentID int64, found bool, err error) {
	comments, err := listComments(repo, pr, token)
	if err != nil {
		return 0, false, err
	}
	for _, c := range comments {
		if strings.Contains(c.Body, guardCommentMarker) {
			return c.ID, true, nil
		}
	}
	return 0, false, nil
}

// upsertGuardComment finds the guard's existing marked comment and patches it,
// or posts a new one if none exists yet.
func upsertGuardComment(cfg guardConfig, pr int, body string) error {
	if cfg.DryRun {
		return nil
	}
	id, found, err := findGuardComment(cfg.Repo, pr, cfg.Token)
	if err != nil {
		return fmt.Errorf("find existing guard comment: %w", err)
	}
	if found {
		return patchComment(cfg.Repo, id, body, cfg.Token)
	}
	_, err = postComment(cfg.Repo, pr, body, cfg.Token)
	return err
}

// feedbackMachineReadable is the fenced JSON block a subsequent agent iteration
// can parse directly, instead of NLP-ing the surrounding prose.
type feedbackMachineReadable struct {
	Strike     int               `json:"strike"`
	MaxStrikes int               `json:"max_strikes"`
	Failures   []hardWallFailure `json:"failures,omitempty"`
	Triage     *triageDecision   `json:"triage,omitempty"`
}

// renderFeedbackComment builds the structured rejection comment for strikes 1-2.
func renderFeedbackComment(strike int, failures []hardWallFailure, triage *triageDecision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n## \U0001F6A7 OSSF-Scout PR Guard — Attempt %d/%d\n\n", guardCommentMarker, strike, maxStrikes)

	summary := "rejected by semantic triage"
	if len(failures) > 0 {
		summary = "rejected by hard-wall checks"
	}
	fmt.Fprintf(&b, "**Result:** ❌ Rejected — %s\n\n", summary)

	if len(failures) > 0 {
		b.WriteString("### Violated policies\n\n")
		for _, f := range failures {
			fmt.Fprintf(&b, "- **%s**: %s\n  ```\n  %s\n  ```\n", f.Check, f.Policy, f.Detail)
		}
		b.WriteString("\n")
	}
	if triage != nil {
		fmt.Fprintf(&b, "### Semantic triage\n\n%s\n\n", triage.Reason)
	}

	mr := feedbackMachineReadable{Strike: strike, MaxStrikes: maxStrikes, Failures: failures, Triage: triage}
	mrJSON, _ := json.Marshal(mr)
	fmt.Fprintf(&b, "### Machine-readable summary\n\n```json\n%s\n```\n\n", mrJSON)

	remaining := maxStrikes - strike
	fmt.Fprintf(&b, "---\n*Next step: address the issues above and push a new commit. %d attempt(s) remain before this PR is automatically closed.*\n", remaining)
	return b.String()
}

// renderClosedComment builds the final comment posted when strike 3 force-closes the PR.
func renderClosedComment(failures []hardWallFailure, triage *triageDecision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n## \U0001F512 OSSF-Scout PR Guard — Closed (strike %d/%d)\n\n", guardCommentMarker, maxStrikes, maxStrikes)
	b.WriteString("This PR has been automatically closed after reaching the maximum number of failed attempts.\n\n")

	if len(failures) > 0 {
		b.WriteString("### Final violated policies\n\n")
		for _, f := range failures {
			fmt.Fprintf(&b, "- **%s**: %s\n  ```\n  %s\n  ```\n", f.Check, f.Policy, f.Detail)
		}
		b.WriteString("\n")
	}
	if triage != nil {
		fmt.Fprintf(&b, "### Semantic triage\n\n%s\n\n", triage.Reason)
	}

	mr := feedbackMachineReadable{Strike: maxStrikes, MaxStrikes: maxStrikes, Failures: failures, Triage: triage}
	mrJSON, _ := json.Marshal(mr)
	fmt.Fprintf(&b, "### Machine-readable summary\n\n```json\n%s\n```\n", mrJSON)
	return b.String()
}

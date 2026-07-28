package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// linkedIssueRe matches GitHub's auto-close keywords, case-insensitively.
var linkedIssueRe = regexp.MustCompile(`(?i)\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s*:?\s*#(\d+)`)

// extractLinkedIssue finds the first "Fixes #N"-style reference in a PR body.
// Used for triage context only — not as a blast-radius bypass.
func extractLinkedIssue(prBody string) (issueNumber int, ok bool) {
	m := linkedIssueRe.FindStringSubmatch(prBody)
	if m == nil {
		return 0, false
	}
	var n int
	if _, err := fmt.Sscanf(m[1], "%d", &n); err != nil {
		return 0, false
	}
	return n, true
}

const guardTriageSystemPrompt = `You are a strict code-review gatekeeper filtering AI-agent-generated pull requests ` +
	`before human review. Given a diff and the problem it claims to solve, decide if the diff genuinely and ` +
	`minimally solves that problem, or is noise (hallucinated, off-topic, unnecessarily large, or unrelated to ` +
	`the referenced issue). Respond with ONLY a single JSON object, no markdown fences, no commentary: ` +
	`{"decision":"approve"|"reject","reason":"<one paragraph>"}`

type triageDecision struct {
	Decision string `json:"decision"` // "approve" | "reject"
	Reason   string `json:"reason"`
}

// buildTriageUserPrompt mirrors buildUserPrompt's structure (audit_prompt.go)
// but is far smaller — a diff-vs-issue judgment, not a full evidence report.
func buildTriageUserPrompt(pr *ghPullRequest, issueTitle, issueBody string, files []ghPRFile) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Pull Request\n\nTitle: %s\n\nBody:\n%s\n\n", pr.Title, truncateField(pr.Body, 2_000))

	if issueTitle != "" {
		fmt.Fprintf(&b, "## Linked issue\n\nTitle: %s\n\nBody:\n%s\n\n", issueTitle, truncateField(issueBody, 4_000))
	} else {
		b.WriteString("## Linked issue\n\nNone found.\n\n")
	}

	b.WriteString("## Changed files\n\n")
	for _, f := range files {
		fmt.Fprintf(&b, "### %s (%s, +%d/-%d)\n\n", f.Filename, f.Status, f.Additions, f.Deletions)
		patch := f.Patch
		if patch == "" {
			b.WriteString("(no patch available — binary or too large)\n\n")
			continue
		}
		fmt.Fprintf(&b, "```diff\n%s\n```\n\n", truncateField(patch, truncGuardPatchPerFile))
	}

	prompt := b.String()
	if len(prompt) > guardMaxPromptChars {
		prompt = prompt[:guardMaxPromptChars] + "\n\n[Diff truncated to fit model context window]"
	}
	return prompt
}

// parseTriageDecision defensively strips ``` fences the model may add despite
// being told not to, then decodes the JSON decision object.
func parseTriageDecision(raw string) (triageDecision, error) {
	s := stripThinkBlocks(raw)
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	var d triageDecision
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return triageDecision{}, fmt.Errorf("could not parse triage decision: %w", err)
	}
	if d.Decision != "approve" && d.Decision != "reject" {
		return triageDecision{}, fmt.Errorf("unexpected triage decision %q", d.Decision)
	}
	return d, nil
}

// runTriage only runs once the hard wall has passed — there's no point judging
// code that already failed SAST/SCA/blast-radius/provenance.
func runTriage(cfg guardConfig, pr *ghPullRequest, files []ghPRFile) (triageDecision, error) {
	issueTitle, issueBody := "", ""
	if issueNum, ok := extractLinkedIssue(pr.Body); ok {
		if issue, err := getIssue(cfg.Repo, issueNum, cfg.Token); err == nil {
			issueTitle, issueBody = issue.Title, issue.Body
		}
	}

	userPrompt := buildTriageUserPrompt(pr, issueTitle, issueBody, files)
	raw, _, _, err := generateOllamaChat(cfg.OllamaURL, cfg.OllamaModel, guardTriageSystemPrompt, userPrompt)
	if err != nil {
		return triageDecision{}, fmt.Errorf("ollama triage request failed: %w", err)
	}
	return parseTriageDecision(raw)
}

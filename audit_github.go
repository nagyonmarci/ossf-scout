package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// fetchIssuesPRsSummary fetches issues and PRs for a repo from the GitHub API
// and returns a compact markdown summary ready for AI prompts, plus the raw JSON.
func fetchIssuesPRsSummary(repo, ghToken string) (summaryMD, dataJSON string, err error) {
	do := func(url string) ([]map[string]interface{}, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)
		if ghToken != "" {
			req.Header.Set("Authorization", "Bearer "+ghToken)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close() //nolint:errcheck
		var items []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			return nil, err
		}
		return items, nil
	}

	base := "https://api.github.com/repos/" + repo

	openIssues, _ := do(base + "/issues?state=open&per_page=100&filter=all")
	closedIssues, _ := do(base + "/issues?state=closed&per_page=50&sort=updated&filter=all")
	openPRs, _ := do(base + "/pulls?state=open&per_page=50")
	mergedPRs, _ := do(base + "/pulls?state=closed&per_page=30&sort=updated")

	data := map[string]interface{}{
		"open_issues":   openIssues,
		"closed_issues": closedIssues,
		"open_prs":      openPRs,
		"merged_prs":    mergedPRs,
		"fetched_at":    time.Now().UTC().Format(time.RFC3339),
	}
	raw, _ := json.Marshal(data)
	dataJSON = string(raw)

	summaryMD = buildIssuesPRsSummary(repo, openIssues, closedIssues, openPRs, mergedPRs)
	return summaryMD, dataJSON, nil
}

// buildIssuesPRsSummary produces a compact markdown block for AI prompt injection.
func buildIssuesPRsSummary(repo string, openIssues, closedIssues, openPRs, mergedPRs []map[string]interface{}) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Issues & PRs — %s\n\n", repo)

	// Count security-labeled issues
	secLabels := []string{"security", "vulnerability", "cve", "exploit", "auth", "injection"}
	var secIssues []string
	for _, iss := range openIssues {
		title, _ := iss["title"].(string)
		if issueTouchesLabels(iss, secLabels) {
			num, _ := iss["number"].(float64)
			secIssues = append(secIssues, fmt.Sprintf("#%.0f %s", num, truncate(title, 80)))
		}
	}

	fmt.Fprintf(&b, "**Open issues:** %d", len(openIssues))
	if len(secIssues) > 0 {
		fmt.Fprintf(&b, " | **Security-labeled:** %d", len(secIssues))
	}
	fmt.Fprintf(&b, " | **Closed (recent):** %d\n", len(closedIssues))

	// Oldest open issue
	if len(openIssues) > 0 {
		oldest := oldestDate(openIssues)
		if oldest != "" {
			fmt.Fprintf(&b, "**Oldest open issue:** %s\n", oldest[:10])
		}
	}

	// Security issues list
	if len(secIssues) > 0 {
		b.WriteString("\n**Open security-labeled issues:**\n")
		for i, s := range secIssues {
			if i >= 10 {
				fmt.Fprintf(&b, "  … and %d more\n", len(secIssues)-10)
				break
			}
			fmt.Fprintf(&b, "  - %s\n", s)
		}
	}

	// Open PRs
	fmt.Fprintf(&b, "\n**Open PRs:** %d", len(openPRs))
	var draftCount int
	var secPRs []string
	for _, pr := range openPRs {
		if d, _ := pr["draft"].(bool); d {
			draftCount++
		}
		title, _ := pr["title"].(string)
		if prTouchesSecurityFiles(pr) || isSecurityTitle(title) {
			num, _ := pr["number"].(float64)
			secPRs = append(secPRs, fmt.Sprintf("#%.0f %s", num, truncate(title, 80)))
		}
	}
	if draftCount > 0 {
		fmt.Fprintf(&b, " (%d draft)", draftCount)
	}
	b.WriteString("\n")
	if len(secPRs) > 0 {
		b.WriteString("**Security-related open PRs:**\n")
		for i, s := range secPRs {
			if i >= 5 {
				break
			}
			fmt.Fprintf(&b, "  - %s\n", s)
		}
	}

	// Recent merged PRs touching security
	var secMerged []string
	for _, pr := range mergedPRs {
		title, _ := pr["title"].(string)
		if isSecurityTitle(title) {
			num, _ := pr["number"].(float64)
			secMerged = append(secMerged, fmt.Sprintf("#%.0f %s", num, truncate(title, 80)))
		}
	}
	if len(secMerged) > 0 {
		b.WriteString("\n**Recently merged security PRs:**\n")
		for i, s := range secMerged {
			if i >= 5 {
				break
			}
			fmt.Fprintf(&b, "  - %s\n", s)
		}
	}

	return b.String()
}

func issueTouchesLabels(issue map[string]interface{}, keywords []string) bool {
	labels, _ := issue["labels"].([]interface{})
	for _, l := range labels {
		lm, _ := l.(map[string]interface{})
		name, _ := lm["name"].(string)
		lower := strings.ToLower(name)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return true
			}
		}
	}
	title, _ := issue["title"].(string)
	return isSecurityTitle(title)
}

func isSecurityTitle(title string) bool {
	lower := strings.ToLower(title)
	keywords := []string{"security", "vuln", "cve-", "exploit", "injection", "xss", "sqli",
		"ssrf", "csrf", "auth", "bypass", "rce", "dos", "leak", "exposure"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func prTouchesSecurityFiles(pr map[string]interface{}) bool {
	// PR webhook payload may not include files, rely on title
	title, _ := pr["title"].(string)
	return isSecurityTitle(title)
}

func oldestDate(items []map[string]interface{}) string {
	oldest := ""
	for _, item := range items {
		created, _ := item["created_at"].(string)
		if created != "" && (oldest == "" || created < oldest) {
			oldest = created
		}
	}
	return oldest
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

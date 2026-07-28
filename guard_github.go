package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ── Low-level request executor ───────────────────────────────────────────────

// ghDo generalizes ghGet's header pattern (github.go:27) to methods beyond GET.
// body may be nil for verbs that don't send one (GET/DELETE).
func ghDo(method, rawURL, token string, body []byte) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close() //nolint:errcheck
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= 300 {
		return respBody, resp.StatusCode, fmt.Errorf("github %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, resp.StatusCode, nil
}

func ghPost(rawURL, token string, body []byte) ([]byte, error) {
	respBody, _, err := ghDo("POST", rawURL, token, body)
	return respBody, err
}

func ghPatch(rawURL, token string, body []byte) ([]byte, error) {
	respBody, _, err := ghDo("PATCH", rawURL, token, body)
	return respBody, err
}

// ghDelete treats a 404 as success — the target (e.g. a label) is already absent.
func ghDelete(rawURL, token string) error {
	_, status, err := ghDo("DELETE", rawURL, token, nil)
	if err != nil && status != http.StatusNotFound {
		return err
	}
	return nil
}

// ── PR Guard API types ────────────────────────────────────────────────────────

type ghLabel struct {
	Name string `json:"name"`
}

type ghPullRequest struct {
	Number       int       `json:"number"`
	State        string    `json:"state"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	ChangedFiles int       `json:"changed_files"`
	Labels       []ghLabel `json:"labels"`
	Head         struct {
		SHA string `json:"sha"`
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

type ghPRFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"` // added / modified / removed / renamed
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"` // unified diff hunk; absent for binary/large files
}

type ghCommitVerification struct {
	Commit struct {
		Verification struct {
			Verified bool   `json:"verified"`
			Reason   string `json:"reason"`
		} `json:"verification"`
	} `json:"commit"`
}

type ghComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

type ghIssue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// ── PR Guard API calls ────────────────────────────────────────────────────────

// githubAPIBase is overridden in tests to point at an httptest.Server.
var githubAPIBase = "https://api.github.com"

func getPullRequest(repo string, pr int, token string) (*ghPullRequest, error) {
	body, err := ghGet(fmt.Sprintf("%s/repos/%s/pulls/%d", githubAPIBase, repo, pr), token)
	if err != nil {
		return nil, err
	}
	var out ghPullRequest
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func getPullRequestFiles(repo string, pr int, token string) ([]ghPRFile, error) {
	var all []ghPRFile
	page := 1
	for {
		u := fmt.Sprintf(githubAPIBase+"/repos/%s/pulls/%d/files?per_page=%d&page=%d",
			repo, pr, defaultPageSize, page)
		body, err := ghGet(u, token)
		if err != nil {
			return nil, err
		}
		var files []ghPRFile
		if err := json.Unmarshal(body, &files); err != nil {
			return nil, err
		}
		if len(files) == 0 {
			break
		}
		all = append(all, files...)
		if len(files) < defaultPageSize {
			break
		}
		page++
	}
	return all, nil
}

func getCommitVerification(repo, sha, token string) (verified bool, reason string, err error) {
	body, err := ghGet(fmt.Sprintf(githubAPIBase+"/repos/%s/commits/%s", repo, sha), token)
	if err != nil {
		return false, "", err
	}
	var out ghCommitVerification
	if err := json.Unmarshal(body, &out); err != nil {
		return false, "", err
	}
	return out.Commit.Verification.Verified, out.Commit.Verification.Reason, nil
}

func getIssue(repo string, issueNumber int, token string) (*ghIssue, error) {
	body, err := ghGet(fmt.Sprintf(githubAPIBase+"/repos/%s/issues/%d", repo, issueNumber), token)
	if err != nil {
		return nil, err
	}
	var out ghIssue
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func addLabel(repo string, pr int, label, token string) error {
	body, _ := json.Marshal(map[string][]string{"labels": {label}})
	_, err := ghPost(fmt.Sprintf(githubAPIBase+"/repos/%s/issues/%d/labels", repo, pr), token, body)
	return err
}

func removeLabel(repo string, pr int, label, token string) error {
	return ghDelete(fmt.Sprintf(githubAPIBase+"/repos/%s/issues/%d/labels/%s", repo, pr, url.PathEscape(label)), token)
}

func postComment(repo string, pr int, commentBody, token string) (int64, error) {
	payload, _ := json.Marshal(map[string]string{"body": commentBody})
	resp, err := ghPost(fmt.Sprintf(githubAPIBase+"/repos/%s/issues/%d/comments", repo, pr), token, payload)
	if err != nil {
		return 0, err
	}
	var c ghComment
	if err := json.Unmarshal(resp, &c); err != nil {
		return 0, err
	}
	return c.ID, nil
}

func patchComment(repo string, commentID int64, commentBody, token string) error {
	payload, _ := json.Marshal(map[string]string{"body": commentBody})
	_, err := ghPatch(fmt.Sprintf(githubAPIBase+"/repos/%s/issues/comments/%d", repo, commentID), token, payload)
	return err
}

func listComments(repo string, pr int, token string) ([]ghComment, error) {
	body, err := ghGet(fmt.Sprintf(githubAPIBase+"/repos/%s/issues/%d/comments?per_page=100", repo, pr), token)
	if err != nil {
		return nil, err
	}
	var out []ghComment
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func closePR(repo string, pr int, token string) error {
	body, _ := json.Marshal(map[string]string{"state": "closed"})
	_, err := ghPatch(fmt.Sprintf(githubAPIBase+"/repos/%s/pulls/%d", repo, pr), token, body)
	return err
}

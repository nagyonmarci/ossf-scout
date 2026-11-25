package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractLinkedIssue(t *testing.T) {
	cases := []struct {
		body    string
		wantNum int
		wantOK  bool
	}{
		{"Fixes #12", 12, true},
		{"closes #7", 7, true},
		{"Resolves: #3", 3, true},
		{"Fixed something unrelated", 0, false},
		{"no reference here", 0, false},
	}
	for _, c := range cases {
		n, ok := extractLinkedIssue(c.body)
		if ok != c.wantOK || (ok && n != c.wantNum) {
			t.Errorf("extractLinkedIssue(%q) = (%d, %v), want (%d, %v)", c.body, n, ok, c.wantNum, c.wantOK)
		}
	}
}

func TestParseTriageDecisionStripsFencesAndThink(t *testing.T) {
	raw := "<think>reasoning here</think>```json\n{\"decision\":\"approve\",\"reason\":\"looks good\"}\n```"
	d, err := parseTriageDecision(raw)
	if err != nil {
		t.Fatalf("parseTriageDecision: %v", err)
	}
	if d.Decision != "approve" || d.Reason != "looks good" {
		t.Errorf("got %+v", d)
	}
}

func TestParseTriageDecisionRejectsUnknownDecision(t *testing.T) {
	if _, err := parseTriageDecision(`{"decision":"maybe","reason":"x"}`); err == nil {
		t.Error("expected error for unrecognized decision value")
	}
}

func TestParseTriageDecisionRejectsInvalidJSON(t *testing.T) {
	if _, err := parseTriageDecision("not json at all"); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// sseChunk formats a single OpenAI-compatible SSE line, matching the shape
// readOllamaStream (audit_ollama.go) expects.
func sseChunk(content string) string {
	return fmt.Sprintf(`data: {"choices":[{"delta":{"content":%q}}]}`+"\n", content)
}

func TestRunTriageParsesOllamaResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fw := w.(http.Flusher)
		_, _ = w.Write([]byte(sseChunk(`{"decision":"reject","reason":"unrelated to the issue"}`)))
		fw.Flush()
		_, _ = w.Write([]byte("data: [DONE]\n"))
		fw.Flush()
	}))
	defer srv.Close()

	cfg := guardConfig{OllamaURL: srv.URL, OllamaModel: "llama3.1"}
	pr := &ghPullRequest{Title: "Add feature", Body: "no issue reference"}
	files := []ghPRFile{{Filename: "main.go", Status: "modified", Patch: "@@ -1 +1 @@\n-a\n+b"}}

	decision, err := runTriage(cfg, pr, files)
	if err != nil {
		t.Fatalf("runTriage: %v", err)
	}
	if decision.Decision != "reject" || decision.Reason != "unrelated to the issue" {
		t.Errorf("got %+v", decision)
	}
}

func TestBuildTriageUserPromptHandlesMissingPatch(t *testing.T) {
	pr := &ghPullRequest{Title: "T", Body: "B"}
	files := []ghPRFile{{Filename: "big.bin", Status: "added"}}
	prompt := buildTriageUserPrompt(pr, "", "", files)
	if !strings.Contains(prompt, "no patch available") {
		t.Errorf("expected missing-patch note in prompt, got: %s", prompt)
	}
}

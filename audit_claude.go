package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ── Claude API types ──────────────────────────────────────────────────────────

type claudeSystemBlock struct {
	Type         string              `json:"type"`
	Text         string              `json:"text"`
	CacheControl *claudeCacheControl `json:"cache_control,omitempty"`
}

type claudeCacheControl struct {
	Type string `json:"type"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    []claudeSystemBlock `json:"system"`
	Messages  []claudeMessage     `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens          int `json:"input_tokens"`
		CacheReadInputTokens int `json:"cache_read_input_tokens"`
		OutputTokens         int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

const defaultModel = "claude-opus-4-8"

// ── Claude report generation ────────────────────────────────────────────────

func callClaude(systemPrompt, userPrompt, apiKey, model string, maxTokens int) (text string, inputTokens, outputTokens int, err error) {
	payload := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System: []claudeSystemBlock{
			{
				Type:         "text",
				Text:         systemPrompt,
				CacheControl: &claudeCacheControl{Type: "ephemeral"},
			},
		},
		Messages: []claudeMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Minute}

	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")

		resp, rerr := client.Do(req)
		if rerr != nil {
			return "", 0, 0, fmt.Errorf("claude API request failed: %w", rerr)
		}

		var cr claudeResponse
		decErr := json.NewDecoder(resp.Body).Decode(&cr)
		resp.Body.Close() //nolint:errcheck
		if decErr != nil {
			return "", 0, 0, fmt.Errorf("claude API decode failed: %w", decErr)
		}

		if cr.Error != nil {
			if cr.Error.Type == "rate_limit_error" && attempt == 0 {
				time.Sleep(65 * time.Second)
				continue
			}
			return "", 0, 0, fmt.Errorf("claude API error %s: %s", cr.Error.Type, cr.Error.Message)
		}
		if len(cr.Content) == 0 {
			return "", 0, 0, fmt.Errorf("claude API returned empty content")
		}
		return cr.Content[0].Text, cr.Usage.InputTokens, cr.Usage.OutputTokens, nil
	}
	return "", 0, 0, fmt.Errorf("claude API rate limit not resolved after retry")
}

func generateReport(ctx *auditContext, apiKey, model string) (report string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultModel
	}
	report, inputTokens, outputTokens, err = callClaude(auditSystemPrompt, buildUserPrompt(ctx), apiKey, model, defaultMaxTokens)
	if err != nil && isContextTooLarge(err) {
		compact := compactForOllama(ctx)
		report, inputTokens, outputTokens, err = callClaude(auditSystemPrompt, buildUserPrompt(compact), apiKey, model, defaultMaxTokens)
	}
	return
}

func generateSplitClaudeReport(ctx *auditContext, apiKey, analysisModel, finalModel string) (report string, inputTokens, outputTokens int, err error) {
	if finalModel == "" {
		finalModel = defaultModel
	}
	if analysisModel == "" {
		analysisModel = "claude-haiku-4-5-20251001"
	}

	sections := buildAuditSummarySections(ctx)
	summaries := make([]auditSectionSummary, 0, len(sections))
	for _, section := range sections {
		prompt := fmt.Sprintf(`Summarize this DevSecOps audit evidence section for a later final report writer.

Section: %s

Rules:
- Keep concrete file paths, tool outputs, commands, package names, workflow names, and API evidence.
- Identify actionable findings, likely false positives, and clear "no issue found" categories.
- Calibrate severity. Do not inflate weak evidence.
- Output concise Markdown with headings: Evidence, Findings, No-Issue Notes, Open Questions.
- Do not write the final report.

Evidence:

%s`, section.Name, section.Content)

		summary, in, out, serr := callClaude(auditSummarySystemPrompt, prompt, apiKey, analysisModel, 2048)
		inputTokens += in
		outputTokens += out
		if serr != nil {
			return "", inputTokens, outputTokens, fmt.Errorf("summarize %s: %w", section.Name, serr)
		}
		summaries = append(summaries, auditSectionSummary{
			Section: section.Name,
			Model:   analysisModel,
			Summary: summary,
		})
	}

	report, in, out, err := callClaude(auditSystemPrompt, buildSummarizedUserPrompt(ctx, summaries), apiKey, finalModel, 8192)
	inputTokens += in
	outputTokens += out
	return report, inputTokens, outputTokens, err
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ── Ollama API types ──────────────────────────────────────────────────────────

// ── Ollama API ────────────────────────────────────────────────────────────────

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type ollamaStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}


// ── Ollama context compaction ────────────────────────────────────────────────

// compactForOllama returns a shallow copy of ctx with verbose tool outputs
// truncated so the total JSON fits within typical Ollama context windows.
func compactForOllama(ctx *auditContext) *auditContext {
	c := *ctx
	c.CICD.WorkflowContents = truncateField(ctx.CICD.WorkflowContents, truncWorkflowContents)
	c.CICD.Zizmor = truncateField(ctx.CICD.Zizmor, truncZizmor)
	c.KeyFiles.AuthMiddleware = truncateField(ctx.KeyFiles.AuthMiddleware, truncAuthMiddleware)
	c.KeyFiles.PermissionSystem = truncateField(ctx.KeyFiles.PermissionSystem, truncPermissionSystem)
	c.KeyFiles.StartupValidation = truncateField(ctx.KeyFiles.StartupValidation, truncStartupValidation)
	c.KeyFiles.ErrorHandler = truncateField(ctx.KeyFiles.ErrorHandler, truncErrorHandler)
	c.KeyFiles.HelmetConfig = truncateField(ctx.KeyFiles.HelmetConfig, truncHelmetConfig)
	c.Dependencies.PnpmAudit = truncateField(ctx.Dependencies.PnpmAudit, truncDepsAudit)
	c.Secrets.Gitleaks = truncateField(ctx.Secrets.Gitleaks, truncGitleaks)
	c.Secrets.TruffleHog = truncateField(ctx.Secrets.TruffleHog, truncTruffleHog)
	c.IaC.Checkov = truncateField(ctx.IaC.Checkov, truncCheckov)
	c.IaC.Trivy = truncateField(ctx.IaC.Trivy, truncTrivy)
	c.IaC.KubeLinter = truncateField(ctx.IaC.KubeLinter, truncKubeLinter)
	return &c
}

func generateOllamaReport(ctx *auditContext, ollamaURL, model string) (report string, inputTokens, outputTokens int, err error) {
	if ollamaURL == "" {
		ollamaURL = defaultOllamaURL
	}
	return generateOllamaReportWith(ctx, ollamaURL, model, false)
}

func generateSplitOllamaReport(ctx *auditContext, ollamaURL, analysisModel, finalModel string) (report string, inputTokens, outputTokens int, err error) {
	if ollamaURL == "" {
		ollamaURL = defaultOllamaURL
	}
	if analysisModel == "" {
		analysisModel = finalModel
	}

	sections := buildAuditSummarySections(ctx)
	summaries := make([]auditSectionSummary, 0, len(sections))
	for _, section := range sections {
		evidence := section.Content
		if len(evidence) > ollamaMaxSummaryPromptChars {
			evidence = evidence[:ollamaMaxSummaryPromptChars] + "\n\n[Section evidence truncated to fit the analysis model context window]"
		}
		prompt := fmt.Sprintf(`Summarize this DevSecOps audit evidence section for a later final report writer.

Section: %s

Rules:
- Keep concrete file paths, tool outputs, commands, package names, workflow names, and API evidence.
- Identify actionable findings, likely false positives, and clear "no issue found" categories.
- Calibrate severity. Do not inflate weak evidence.
- Output concise Markdown with headings: Evidence, Findings, No-Issue Notes, Open Questions.
- Do not write the final report.

Evidence:

%s`, section.Name, evidence)

		summary, in, out, serr := generateOllamaChat(ollamaURL, analysisModel, auditSummarySystemPrompt, prompt)
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

	finalPrompt := buildSummarizedUserPrompt(ctx, summaries)
	report, in, out, err := generateOllamaChat(ollamaURL, finalModel, auditSystemPrompt, finalPrompt)
	inputTokens += in
	outputTokens += out
	if err != nil {
		return "", inputTokens, outputTokens, fmt.Errorf("final report: %w", err)
	}
	return report, inputTokens, outputTokens, nil
}

// ── Ollama report generation ─────────────────────────────────────────────────

func generateOllamaReportWith(ctx *auditContext, ollamaURL, model string, compact bool) (report string, inputTokens, outputTokens int, err error) {
	sendCtx := ctx
	if compact {
		sendCtx = compactForOllama(ctx)
	}
	userPrompt := buildUserPrompt(sendCtx)
	if compact && len(userPrompt) > ollamaMaxPromptChars {
		userPrompt = userPrompt[:ollamaMaxPromptChars] + "\n\n[Context truncated to fit model context window]"
	}
	report, inputTokens, outputTokens, err = generateOllamaChat(ollamaURL, model, auditSystemPrompt, userPrompt)
	if err != nil {
		if !compact && isOllamaCompactRetryError(err) {
			return generateOllamaReportWith(ctx, ollamaURL, model, true)
		}
		return "", 0, 0, err
	}

	return report, inputTokens, outputTokens, nil
}

func generateOllamaChat(ollamaURL, model, systemPrompt, userPrompt string) (report string, inputTokens, outputTokens int, err error) {
	payload := ollamaRequest{
		Model:  model,
		Stream: true,
		Messages: []ollamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Hour}
	resp, err := client.Do(req)
	if err != nil {
		hint := ""
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "refused") {
			hint = " (is Ollama running and reachable? if inside Docker, set OLLAMA_HOST=0.0.0.0 on the host)"
		}
		return "", 0, 0, fmt.Errorf("ollama request failed: %w%s", err, hint)
	}
	defer resp.Body.Close() //nolint:errcheck

	report, inputTokens, outputTokens, err = readOllamaStream(resp)
	if err != nil {
		return "", 0, 0, err
	}

	return report, inputTokens, outputTokens, nil
}

func isOllamaCompactRetryError(err error) bool {
	msg := err.Error()
	contextTooLong := strings.Contains(msg, "exceeds") && strings.Contains(msg, "context")
	backendCrash := strings.Contains(msg, "EOF") || strings.Contains(msg, "api_error")
	return contextTooLong || backendCrash
}

func readOllamaStream(resp *http.Response) (report string, inputTokens, outputTokens int, err error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var or_ ollamaResponse
		if derr := json.Unmarshal(body, &or_); derr == nil && or_.Error != nil {
			return "", 0, 0, fmt.Errorf("ollama error %s: %s", or_.Error.Type, or_.Error.Message)
		}
		return "", 0, 0, fmt.Errorf("ollama HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var b strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		if line == "[DONE]" {
			break
		}

		var chunk ollamaStreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return "", 0, 0, fmt.Errorf("ollama stream decode failed: %w", err)
		}
		if chunk.Error != nil {
			return "", 0, 0, fmt.Errorf("ollama error %s: %s", chunk.Error.Type, chunk.Error.Message)
		}
		if chunk.Usage.PromptTokens > 0 {
			inputTokens = chunk.Usage.PromptTokens
		}
		if chunk.Usage.CompletionTokens > 0 {
			outputTokens = chunk.Usage.CompletionTokens
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				b.WriteString(choice.Delta.Content)
			} else if choice.Message.Content != "" {
				b.WriteString(choice.Message.Content)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", 0, 0, fmt.Errorf("ollama stream failed: %w", err)
	}
	if b.Len() == 0 {
		return "", 0, 0, fmt.Errorf("ollama returned empty content")
	}
	return b.String(), inputTokens, outputTokens, nil
}


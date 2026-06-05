package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ── OpenAI API types ──────────────────────────────────────────────────────────

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model               string          `json:"model"`
	MaxCompletionTokens int             `json:"max_completion_tokens"`
	Messages            []openAIMessage `json:"messages"`
}

type openAIResponse struct {
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
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// ── OpenAI report generation ──────────────────────────────────────────────────

func callOpenAI(systemPrompt, userPrompt, apiKey, model string, maxTokens int) (text string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultOpenAIModel
	}
	payload := openAIRequest{
		Model:     model,
		MaxCompletionTokens: maxTokens,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	bodyBytes, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Minute}

	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest("POST", openAIEndpoint, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, rerr := client.Do(req)
		if rerr != nil {
			return "", 0, 0, fmt.Errorf("openai request failed: %w", rerr)
		}
		var or openAIResponse
		decErr := json.NewDecoder(resp.Body).Decode(&or)
		resp.Body.Close() //nolint:errcheck
		if decErr != nil {
			return "", 0, 0, fmt.Errorf("openai decode failed: %w", decErr)
		}
		if or.Error != nil {
			if (or.Error.Type == "rate_limit_error" || or.Error.Code == "rate_limit_exceeded") && attempt == 0 {
				time.Sleep(65 * time.Second)
				continue
			}
			return "", 0, 0, fmt.Errorf("openai error %s: %s", or.Error.Type, or.Error.Message)
		}
		if len(or.Choices) == 0 {
			return "", 0, 0, fmt.Errorf("openai returned empty choices")
		}
		return or.Choices[0].Message.Content, or.Usage.PromptTokens, or.Usage.CompletionTokens, nil
	}
	return "", 0, 0, fmt.Errorf("openai rate limit not resolved after retry")
}

func generateOpenAIReport(ctx *auditContext, apiKey, model string) (report string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultOpenAIModel
	}
	report, inputTokens, outputTokens, err = callOpenAI(auditSystemPrompt, buildUserPrompt(ctx), apiKey, model, defaultMaxTokens)
	if err != nil && isContextTooLarge(err) {
		compact := compactForOllama(ctx)
		report, inputTokens, outputTokens, err = callOpenAI(auditSystemPrompt, buildUserPrompt(compact), apiKey, model, defaultMaxTokens/2)
	}
	return
}

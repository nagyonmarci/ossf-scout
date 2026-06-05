package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ── Gemini API types ──────────────────────────────────────────────────────────

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  struct {
		MaxOutputTokens int `json:"maxOutputTokens"`
	} `json:"generation_config"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// ── Gemini report generation ──────────────────────────────────────────────────

func callGemini(systemPrompt, userPrompt, apiKey, model string, maxTokens int) (text string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultGeminiModel
	}
	payload := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: userPrompt}}},
		},
	}
	payload.GenerationConfig.MaxOutputTokens = maxTokens

	bodyBytes, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Minute}
	endpoint := fmt.Sprintf(geminiEndpointFmt, model, apiKey)

	for attempt := 0; attempt < 2; attempt++ {
		req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, rerr := client.Do(req)
		if rerr != nil {
			return "", 0, 0, fmt.Errorf("gemini request failed: %w", rerr)
		}
		var gr geminiResponse
		decErr := json.NewDecoder(resp.Body).Decode(&gr)
		resp.Body.Close() //nolint:errcheck
		if decErr != nil {
			return "", 0, 0, fmt.Errorf("gemini decode failed: %w", decErr)
		}
		if gr.Error != nil {
			if gr.Error.Code == 429 && attempt == 0 {
				time.Sleep(65 * time.Second)
				continue
			}
			return "", 0, 0, fmt.Errorf("gemini error %d %s: %s", gr.Error.Code, gr.Error.Status, gr.Error.Message)
		}
		if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
			return "", 0, 0, fmt.Errorf("gemini returned empty candidates")
		}
		return gr.Candidates[0].Content.Parts[0].Text,
			gr.UsageMetadata.PromptTokenCount,
			gr.UsageMetadata.CandidatesTokenCount, nil
	}
	return "", 0, 0, fmt.Errorf("gemini rate limit not resolved after retry")
}

func generateGeminiReport(ctx *auditContext, apiKey, model string) (report string, inputTokens, outputTokens int, err error) {
	if model == "" {
		model = defaultGeminiModel
	}
	return callGemini(auditSystemPrompt, buildUserPrompt(ctx), apiKey, model, 8192)
}

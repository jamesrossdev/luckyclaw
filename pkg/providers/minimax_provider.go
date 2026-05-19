package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesrossdev/luckyclaw/pkg/logger"
)

const minimaxBaseURL = "https://api.minimax.io/v1"
const minimaxTextEndpoint = "/text/chatcompletion_v2"

type MiniMaxProvider struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

func NewMiniMaxProvider(apiKey, apiBase, proxy string) *MiniMaxProvider {
	baseURL := apiBase
	if baseURL == "" {
		baseURL = minimaxBaseURL
	}
	client := &http.Client{Timeout: 120 * time.Second}
	if proxy != "" {
		if proxyURL, err := url.Parse(proxy); err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		}
	}
	return &MiniMaxProvider{
		apiKey:     apiKey,
		apiBase:    strings.TrimRight(baseURL, "/"),
		httpClient: client,
	}
}

func (p *MiniMaxProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("MiniMax API key not configured")
	}

	normalizedModel := p.normalizeModel(model)

	var formattedMessages []interface{}
	mediaCount := 0
	mediaFilteredCount := 0

	for _, msg := range messages {
		formattedMsg := map[string]interface{}{
			"role": msg.Role,
		}

		if len(msg.MediaPaths) > 0 {
			var contentArray []interface{}
			if msg.Content != "" {
				contentArray = append(contentArray, map[string]interface{}{
					"type": "text",
					"text": msg.Content,
				})
			}

			for _, imgPath := range msg.MediaPaths {
				if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imgPath,
						},
					})
					mediaCount++
					continue
				}

				const maxMediaBytes = 2 * 1024 * 1024
				if fi, statErr := os.Stat(imgPath); statErr != nil || fi.Size() > maxMediaBytes {
					mediaFilteredCount++
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("[media too large to embed: %s]", filepath.Base(imgPath)),
					})
					continue
				}

				mimeType := ""
				imgPathLower := strings.ToLower(imgPath)
				if strings.HasSuffix(imgPathLower, ".jpg") || strings.HasSuffix(imgPathLower, ".jpeg") {
					mimeType = "image/jpeg"
				} else if strings.HasSuffix(imgPathLower, ".png") {
					mimeType = "image/png"
				} else if strings.HasSuffix(imgPathLower, ".gif") {
					mimeType = "image/gif"
				} else if strings.HasSuffix(imgPathLower, ".webp") {
					mimeType = "image/webp"
				}

				if mimeType == "" {
					mediaFilteredCount++
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("[attachment omitted: unsupported format %s]", filepath.Base(imgPath)),
					})
					continue
				}

				if imgData, err := os.ReadFile(imgPath); err == nil {
					base64Str := base64.StdEncoding.EncodeToString(imgData)
					dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)
					contentArray = append(contentArray, map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]string{
							"url": dataURL,
						},
					})
					mediaCount++
				} else {
					mediaFilteredCount++
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("[attachment read error: %s]", filepath.Base(imgPath)),
					})
				}
			}
			formattedMsg["content"] = contentArray
		} else {
			formattedMsg["content"] = msg.Content
		}

		if len(msg.ToolCalls) > 0 {
			formattedMsg["tool_calls"] = msg.ToolCalls
		}
		if msg.ToolCallID != "" {
			formattedMsg["tool_call_id"] = msg.ToolCallID
		}

		formattedMessages = append(formattedMessages, formattedMsg)
	}

	requestBody := map[string]interface{}{
		"model":           normalizedModel,
		"messages":        formattedMessages,
		"stream":          false,
		"reasoning_split": true,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	if maxTokens, ok := options["max_tokens"].(int); ok {
		requestBody["max_completion_tokens"] = maxTokens
	}

	if temperature, ok := options["temperature"].(float64); ok {
		requestBody["temperature"] = temperature
	}

	logger.DebugCF("minimax", "MiniMax request", map[string]any{
		"provider":             "minimax",
		"model":                normalizedModel,
		"original_model":       model,
		"has_media":            mediaCount > 0,
		"media_count":          mediaCount,
		"media_filtered_count": mediaFilteredCount,
	})

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MiniMax request: %w", err)
	}

	url := p.apiBase + minimaxTextEndpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create MiniMax request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send MiniMax request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MiniMax response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MiniMax API request failed:\n  Status: %d\n  Body:   %s", resp.StatusCode, string(body))
	}

	response, err := p.parseResponse(body)
	if err != nil {
		return nil, err
	}

	showReasoning, _ := options["show_reasoning"].(bool)
	response.Content = BuildUserVisibleContent(response.Content, response.Reasoning, showReasoning)

	return response, nil
}

func (p *MiniMaxProvider) parseResponse(body []byte) (*LLMResponse, error) {
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
				ReasoningDetails []struct {
					Text string `json:"text"`
				} `json:"reasoning_details"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage    *UsageInfo `json:"usage"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MiniMax response: %w", err)
	}

	if apiResponse.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("MiniMax API error (%d): %s", apiResponse.BaseResp.StatusCode, apiResponse.BaseResp.StatusMsg)
	}

	if len(apiResponse.Choices) == 0 {
		return &LLMResponse{
			Content:      "",
			FinishReason: "stop",
		}, nil
	}

	choice := apiResponse.Choices[0]
	content := choice.Message.Content
	reasoning := strings.TrimSpace(choice.Message.ReasoningContent)
	if len(choice.Message.ReasoningDetails) > 0 {
		reasoningParts := make([]string, 0, len(choice.Message.ReasoningDetails))
		for _, part := range choice.Message.ReasoningDetails {
			partText := strings.TrimSpace(part.Text)
			if partText != "" {
				reasoningParts = append(reasoningParts, partText)
			}
		}
		if len(reasoningParts) > 0 {
			reasoning = strings.Join(reasoningParts, "\n\n")
		}
	}

	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		arguments := make(map[string]interface{})
		name := ""

		if tc.Type == "function" && tc.Function != nil {
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		} else if tc.Function != nil {
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		}

		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      name,
			Arguments: arguments,
		})
	}

	logger.DebugCF("minimax", "MiniMax response", map[string]any{
		"provider":       "minimax",
		"content_length": len(content),
		"tool_calls":     len(toolCalls),
		"finish_reason":  choice.FinishReason,
	})

	return &LLMResponse{
		Content:      content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
		Reasoning:    reasoning,
	}, nil
}

func (p *MiniMaxProvider) normalizeModel(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return trimmed
	}

	lower := strings.ToLower(trimmed)

	if strings.HasPrefix(lower, "minimax-coding-plan/") {
		if idx := strings.Index(trimmed, "/"); idx >= 0 && idx+1 < len(trimmed) {
			trimmed = trimmed[idx+1:]
			lower = strings.ToLower(trimmed)
		}
	}

	switch lower {
	case "minimax-m2.7":
		return "MiniMax-M2.7"
	case "minimax-m2":
		return "MiniMax-M2"
	}

	if strings.HasPrefix(lower, "minimax-") {
		suffix := strings.TrimPrefix(lower, "minimax-")
		if strings.HasPrefix(suffix, "m") && len(suffix) > 1 {
			suffix = "M" + suffix[1:]
		}
		return "MiniMax-" + suffix
	}

	return trimmed
}

func (p *MiniMaxProvider) GetDefaultModel() string {
	return "minimax-m2.7"
}

package providers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jamesrossdev/luckyclaw/pkg/logger"
)

func parseOpenAICompatResponse(body []byte, providerName string) (*LLMResponse, error) {
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
		return nil, fmt.Errorf("failed to unmarshal %s response: %w", providerName, err)
	}

	if apiResponse.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("%s API error (%d): %s", providerName, apiResponse.BaseResp.StatusCode, apiResponse.BaseResp.StatusMsg)
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

	if providerName != "" {
		logger.DebugCF(providerName, providerName+" response", map[string]any{
			"content_length": len(content),
			"tool_calls":     len(toolCalls),
			"finish_reason":  choice.FinishReason,
		})
	}

	return &LLMResponse{
		Content:      content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
		Reasoning:    reasoning,
	}, nil
}

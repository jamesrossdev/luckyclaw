package providers

import (
	"strings"
	"testing"
)

func TestHTTPProviderParseResponse_BaseRespError(t *testing.T) {
	p := &HTTPProvider{}
	_, err := p.parseResponse([]byte(`{"base_resp":{"status_code":1004,"status_msg":"invalid api key"}}`))
	if err == nil {
		t.Fatal("expected error for non-zero base_resp.status_code")
	}
	if !strings.Contains(err.Error(), "1004") {
		t.Fatalf("expected status code in error, got %q", err.Error())
	}
}

func TestHTTPProviderParseResponse_UsesReasoningDetails(t *testing.T) {
	p := &HTTPProvider{}
	resp, err := p.parseResponse([]byte(`{
		"choices": [{
			"message": {
				"content": "final answer",
				"reasoning_content": "fallback",
				"reasoning_details": [{"text": "step one"}, {"text": "step two"}]
			},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 1, "completion_tokens": 2, "total_tokens": 3}
	}`))
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if resp.Reasoning != "step one\n\nstep two" {
		t.Fatalf("unexpected reasoning: %q", resp.Reasoning)
	}
	if resp.Content != "final answer" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
}

func TestMiniMaxProviderParseResponse_BaseRespError(t *testing.T) {
	p := &MiniMaxProvider{}
	_, err := p.parseResponse([]byte(`{"base_resp":{"status_code":1004,"status_msg":"invalid api key"}}`))
	if err == nil {
		t.Fatal("expected error for non-zero base_resp.status_code")
	}
	if !strings.Contains(err.Error(), "1004") {
		t.Fatalf("expected status code in error, got %q", err.Error())
	}
}

func TestMiniMaxProviderParseResponse_UsesReasoningDetails(t *testing.T) {
	p := &MiniMaxProvider{}
	resp, err := p.parseResponse([]byte(`{
		"choices": [{
			"message": {
				"content": "final answer",
				"reasoning_content": "fallback",
				"reasoning_details": [{"text": "step one"}, {"text": "step two"}]
			},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 1, "completion_tokens": 2, "total_tokens": 3}
	}`))
	if err != nil {
		t.Fatalf("parseResponse() error: %v", err)
	}
	if resp.Reasoning != "step one\n\nstep two" {
		t.Fatalf("unexpected reasoning: %q", resp.Reasoning)
	}
	if resp.Content != "final answer" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
}

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type SendFileCallback func(channel, chatID, filePath, caption string) error

type SendFileTool struct {
	sendCallback   SendFileCallback
	defaultChannel string
	defaultChatID  string
}

func NewSendFileTool() *SendFileTool {
	return &SendFileTool{}
}

func (t *SendFileTool) Name() string {
	return "send_file"
}

func (t *SendFileTool) Description() string {
	return "Send a file as a Telegram attachment. Use this to share research reports, logs, or any workspace file directly with the user."
}

func (t *SendFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file to send",
			},
			"caption": map[string]interface{}{
				"type":        "string",
				"description": "Optional caption/summary to display with the file",
			},
		},
		"required": []string{"path"},
	}
}

func (t *SendFileTool) SetContext(channel, chatID string) {
	t.defaultChannel = channel
	t.defaultChatID = chatID
}

func (t *SendFileTool) SetSendCallback(callback SendFileCallback) {
	t.sendCallback = callback
}

func (t *SendFileTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &ToolResult{ForLLM: "path is required", IsError: true}
	}

	caption, _ := args["caption"].(string)

	// Verify file exists
	info, err := os.Stat(path)
	if err != nil {
		return &ToolResult{ForLLM: fmt.Sprintf("file not found: %s", path), IsError: true}
	}
	if info.IsDir() {
		return &ToolResult{ForLLM: "path is a directory, not a file", IsError: true}
	}

	channel := t.defaultChannel
	chatID := t.defaultChatID

	if channel == "" || chatID == "" {
		return &ToolResult{ForLLM: "No target channel/chat specified", IsError: true}
	}

	if t.sendCallback == nil {
		return &ToolResult{ForLLM: "File sending not configured", IsError: true}
	}

	if err := t.sendCallback(channel, chatID, path, caption); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("sending file: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	return &ToolResult{
		ForLLM: fmt.Sprintf("File %s sent to %s:%s", filepath.Base(path), channel, chatID),
		Silent: true,
	}
}

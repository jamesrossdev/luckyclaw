package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesrossdev/luckyclaw/pkg/providers"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"telegram:123456", "telegram_123456"},
		{"discord:987654321", "discord_987654321"},
		{"slack:C01234", "slack_C01234"},
		{"no-colons-here", "no-colons-here"},
		{"multiple:colons:here", "multiple_colons_here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSave_WithColonInKey(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	// Create a session with a key containing colon (typical channel session key).
	key := "telegram:123456"
	sm.GetOrCreate(key)
	sm.AddMessage(key, "user", "hello")

	// Save should succeed even though the key contains ':'
	if err := sm.Save(key); err != nil {
		t.Fatalf("Save(%q) failed: %v", key, err)
	}

	// The file on disk should use sanitized name.
	expectedFile := filepath.Join(tmpDir, "telegram_123456.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("expected session file %s to exist", expectedFile)
	}

	// Load into a fresh manager and verify the session round-trips.
	sm2 := NewSessionManager(tmpDir)
	history := sm2.GetHistory(key)
	if len(history) != 1 {
		t.Fatalf("expected 1 message after reload, got %d", len(history))
	}
	if history[0].Content != "hello" {
		t.Errorf("expected message content %q, got %q", "hello", history[0].Content)
	}
}

func TestSave_RejectsPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	badKeys := []string{"", ".", "..", "foo/bar", "foo\\bar"}
	for _, key := range badKeys {
		sm.GetOrCreate(key)
		if err := sm.Save(key); err == nil {
			t.Errorf("Save(%q) should have failed but didn't", key)
		}
	}
}

func TestAddFullMessage_OversizedContent_Truncated(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	key := "test:session"
	sm.GetOrCreate(key)

	// Create message with content larger than maxMessageContentSize (1MB)
	largeContent := strings.Repeat("A", 1*1024*1024+100)
	msg := providers.Message{
		Role:    "user",
		Content: largeContent,
	}

	sm.AddFullMessage(key, msg)

	history := sm.GetHistory(key)
	if len(history) != 1 {
		t.Fatalf("expected 1 message, got %d", len(history))
	}

	// Content should be truncated
	if len(history[0].Content) > maxMessageContentSize {
		t.Errorf("expected content to be truncated to <= %d bytes, got %d", maxMessageContentSize, len(history[0].Content))
	}

	// Should contain truncation marker
	if !strings.Contains(history[0].Content, "[Content truncated") {
		t.Errorf("expected content to contain '[Content truncated]', got: %s", history[0].Content)
	}
}

func TestAddFullMessage_NormalContent_Preserved(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	key := "test:session"
	sm.GetOrCreate(key)

	normalContent := "Hello, this is a normal message"
	msg := providers.Message{
		Role:    "user",
		Content: normalContent,
	}

	sm.AddFullMessage(key, msg)

	history := sm.GetHistory(key)
	if len(history) != 1 {
		t.Fatalf("expected 1 message, got %d", len(history))
	}

	// Content should be preserved exactly
	if history[0].Content != normalContent {
		t.Errorf("expected content %q, got %q", normalContent, history[0].Content)
	}
}

func TestAddFullMessage_ExactlyAtLimit(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	key := "test:session"
	sm.GetOrCreate(key)

	// Create message with content exactly at maxMessageContentSize
	exactContent := strings.Repeat("B", maxMessageContentSize)
	msg := providers.Message{
		Role:    "user",
		Content: exactContent,
	}

	sm.AddFullMessage(key, msg)

	history := sm.GetHistory(key)
	if len(history) != 1 {
		t.Fatalf("expected 1 message, got %d", len(history))
	}

	// Content at exactly the limit should be preserved (not truncated)
	if len(history[0].Content) != maxMessageContentSize {
		t.Errorf("expected content length %d, got %d", maxMessageContentSize, len(history[0].Content))
	}
}

func TestAddFullMessage_BinaryContent_NotTruncated(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSessionManager(tmpDir)

	key := "test:session"
	sm.GetOrCreate(key)

	// Note: Session manager doesn't detect binary content, it just tracks content length
	// The read_file tool is responsible for rejecting binary files before they reach the session
	// Here we just verify that non-truncated binary content is stored as-is
	binaryContent := string(make([]byte, 500)) // 500 bytes of zeros
	msg := providers.Message{
		Role:    "tool",
		Content: binaryContent,
	}

	sm.AddFullMessage(key, msg)

	history := sm.GetHistory(key)
	if len(history) != 1 {
		t.Fatalf("expected 1 message, got %d", len(history))
	}

	// Binary content under limit should be stored
	if len(history[0].Content) != 500 {
		t.Errorf("expected content length 500, got %d", len(history[0].Content))
	}
}

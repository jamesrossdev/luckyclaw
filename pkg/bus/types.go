package bus

import "context"

type InboundMessage struct {
	Channel    string            `json:"channel"`
	SenderID   string            `json:"sender_id"`
	ChatID     string            `json:"chat_id"`
	Content    string            `json:"content"`
	Media      []string          `json:"media,omitempty"`
	SessionKey string            `json:"session_key"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type OutboundMessage struct {
	Channel   string `json:"channel"`
	ChatID    string `json:"chat_id"`
	Content   string `json:"content"`
	FilePath  string `json:"file_path,omitempty"`   // For file attachments
	ReplyToID string `json:"reply_to_id,omitempty"` // Discord: reply to this message ID
}

type MessageHandler func(InboundMessage) error

// OutboundHandler sends a message synchronously, returning any error.
// Used by SendDirect for error-aware sends (e.g. the message tool).
type OutboundHandler func(ctx context.Context, msg OutboundMessage) error

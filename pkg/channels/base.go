package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/jamesrossdev/luckyclaw/pkg/bus"
	"github.com/jamesrossdev/luckyclaw/pkg/logger"
)

type Channel interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(ctx context.Context, msg bus.OutboundMessage) error
	IsRunning() bool
	IsAllowed(senderID string) bool
}

type BaseChannel struct {
	config    interface{}
	bus       *bus.MessageBus
	running   bool
	name      string
	allowList []string
}

func NewBaseChannel(name string, config interface{}, bus *bus.MessageBus, allowList []string) *BaseChannel {
	return &BaseChannel{
		config:    config,
		bus:       bus,
		name:      name,
		allowList: allowList,
		running:   false,
	}
}

func (c *BaseChannel) Name() string {
	return c.name
}

func (c *BaseChannel) IsRunning() bool {
	return c.running
}

func canonicalID(s string) string {
	s = strings.TrimSpace(s)
	if at := strings.LastIndex(s, "@"); at >= 0 {
		s = s[:at]
	}
	return s
}

func senderCandidates(senderID string) map[string]struct{} {
	candidates := make(map[string]struct{})
	s := strings.TrimSpace(senderID)
	if s == "" {
		return candidates
	}

	if idx := strings.Index(s, "|"); idx > 0 {
		if left := canonicalID(s[:idx]); left != "" {
			candidates[left] = struct{}{}
		}
		rightRaw := strings.TrimSpace(s[idx+1:])
		if isWhatsAppIdentity(rightRaw) {
			if right := canonicalID(rightRaw); right != "" {
				candidates[right] = struct{}{}
			}
		}
		return candidates
	}

	if id := canonicalID(s); id != "" {
		candidates[id] = struct{}{}
	}
	return candidates
}

func isWhatsAppIdentity(s string) bool {
	return strings.HasSuffix(s, "@s.whatsapp.net") ||
		strings.HasSuffix(s, "@lid") ||
		strings.HasSuffix(s, "@c.us")
}

func canonicalAllowedID(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(s, "@") {
		return "", false
	}
	if strings.Contains(s, "|") {
		return "", false
	}
	s = canonicalID(s)
	if s == "" {
		return "", false
	}
	return s, true
}

func (c *BaseChannel) IsAllowed(senderID string) bool {
	if len(c.allowList) == 0 {
		return true
	}

	candidates := senderCandidates(senderID)
	for _, allowed := range c.allowList {
		if allowedCanonical, ok := canonicalAllowedID(allowed); ok {
			if _, exists := candidates[allowedCanonical]; exists {
				return true
			}
		}
	}
	return false
}

func (c *BaseChannel) HandleMessage(senderID, chatID, content string, media []string, metadata map[string]string, stanzaID string, replyToParticipant string) {
	logger.DebugCF("channels", "HandleMessage reply-to", map[string]any{
		"stanzaID":           stanzaID,
		"replyToParticipant": replyToParticipant,
	})

	if !c.IsAllowed(senderID) {
		logger.WarnCF("channels", "Message dropped by allowlist", map[string]any{"sender": senderID, "chat": chatID, "allowed": c.allowList})
		return
	}

	// Build session key: channel:chatID
	sessionKey := fmt.Sprintf("%s:%s", c.name, chatID)

	msg := bus.InboundMessage{
		Channel:            c.name,
		SenderID:           senderID,
		ChatID:             chatID,
		Content:            content,
		Media:              media,
		SessionKey:         sessionKey,
		Metadata:           metadata,
		StanzaID:           stanzaID,
		ReplyToParticipant: replyToParticipant,
	}

	c.bus.PublishInbound(msg)
}

func (c *BaseChannel) SetRunning(running bool) {
	c.running = running
}

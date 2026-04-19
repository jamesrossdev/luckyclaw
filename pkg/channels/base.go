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

// canonicalSenderID normalizes a sender ID to its bare phone/LID.
// It strips any domain suffix (e.g. @s.whatsapp.net) and any compound "|username" segment.
func canonicalSenderID(s string) string {
	s = strings.TrimSpace(s)
	if at := strings.LastIndex(s, "@"); at >= 0 {
		s = s[:at]
	}
	if idx := strings.Index(s, "|"); idx > 0 {
		s = s[:idx]
	}
	return s
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
	if at := strings.LastIndex(s, "@"); at >= 0 {
		s = s[:at]
	}
	if s == "" {
		return "", false
	}
	return s, true
}

func (c *BaseChannel) IsAllowed(senderID string) bool {
	if len(c.allowList) == 0 {
		return true
	}

	idOnly := canonicalSenderID(senderID)
	for _, allowed := range c.allowList {
		if allowedCanonical, ok := canonicalAllowedID(allowed); ok && idOnly == allowedCanonical {
			return true
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

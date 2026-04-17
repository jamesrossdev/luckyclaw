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

func (c *BaseChannel) IsAllowed(senderID string) bool {
	if len(c.allowList) == 0 {
		return true
	}

	// normalizeSenderID converts a possibly-compound WhatsApp sender ID to its
	// bare form for consistent allowlist matching. WhatsApp can present the sender
	// as:
	//   - "phone|lid@s.whatsapp.net"  (compound from GetAltJID)
	//   - "phone"                       (bare phone)
	//   - "lid"                         (bare LID)
	// The allowlist stores bare values like "254108092659" so we strip any
	// "@domain" suffix and the "|username" compound separator.
	normalized := senderID

	// If there's a domain suffix, strip it first (e.g. @s.whatsapp.net or @lid).
	if at := strings.LastIndex(normalized, "@"); at >= 0 {
		tmp := normalized[:at]
		// Also strip any compound "|username" leftover after domain removal.
		if idx := strings.Index(tmp, "|"); idx > 0 {
			tmp = tmp[:idx]
		}
		if tmp != "" {
			normalized = tmp
		}
	} else if idx := strings.Index(normalized, "|"); idx > 0 {
		// No domain but has compound prefix.
		normalized = normalized[:idx]
	}

	for _, allowed := range c.allowList {
		// Strip leading "@" from allowed value for username matching.
		trimmed := strings.TrimPrefix(allowed, "@")
		allowedID := trimmed
		allowedUser := ""
		if idx := strings.Index(trimmed, "|"); idx > 0 {
			allowedID = trimmed[:idx]
			allowedUser = trimmed[idx+1:]
		}

		// Normalize allowed similarly (in case stored value carries a suffix).
		allowedNormalized := allowed
		if at := strings.LastIndex(allowedNormalized, "@"); at >= 0 {
			allowedNormalized = allowedNormalized[:at]
		}
		if idx := strings.Index(allowedNormalized, "|"); idx > 0 {
			allowedNormalized = allowedNormalized[:idx]
		}

		// Compare normalized forms.
		if normalized == allowedNormalized ||
			normalized == allowedID ||
			normalized == trimmed ||
			idPartMatch(normalized, allowedID) ||
			normalized == allowedUser {
			return true
		}
	}
	return false
}

// idPartMatch checks whether the bare sender ID matches an allowed ID,
// also allowing legacy compound entries like "123456|username".
func idPartMatch(sender, allowedID string) bool {
	if sender == allowedID {
		return true
	}
	// If allowedID already contains "|username", compare only the ID part.
	if idx := strings.Index(allowedID, "|"); idx > 0 {
		return sender == allowedID[:idx]
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

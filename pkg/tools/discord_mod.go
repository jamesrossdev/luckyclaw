package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Discord moderation tool callback types.
// These are set by the agent loop from the DiscordChannel methods.
type DeleteMessageCallback func(channelID, messageID string) error
type TimeoutUserCallback func(guildID, userID string, until time.Time) error
type SendMessageCallback func(channel, chatID, content string) error

// --- discord_delete_message tool ---

type DiscordDeleteMessageTool struct {
	deleteCallback DeleteMessageCallback
	sendCallback   SendMessageCallback
	defaultChannel string
	defaultChatID  string
}

func NewDiscordDeleteMessageTool() *DiscordDeleteMessageTool {
	return &DiscordDeleteMessageTool{}
}

func (t *DiscordDeleteMessageTool) Name() string {
	return "discord_delete_message"
}

func (t *DiscordDeleteMessageTool) Description() string {
	return "Delete a message in a Discord channel. Use this to remove offensive or rule-breaking messages. Requires channel_id and message_id (available in quoted message metadata)."
}

func (t *DiscordDeleteMessageTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"description": "The Discord channel ID where the message is",
			},
			"message_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the message to delete",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Reason for deletion (logged to mod-log)",
			},
			"log_channel_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Channel ID to send a moderation log to (e.g., mod-log channel)",
			},
		},
		"required": []string{"channel_id", "message_id", "reason"},
	}
}

func (t *DiscordDeleteMessageTool) SetContext(channel, chatID string) {
	t.defaultChannel = channel
	t.defaultChatID = chatID
}

func (t *DiscordDeleteMessageTool) SetDeleteCallback(cb DeleteMessageCallback) {
	t.deleteCallback = cb
}

func (t *DiscordDeleteMessageTool) SetSendCallback(cb SendMessageCallback) {
	t.sendCallback = cb
}

func (t *DiscordDeleteMessageTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	if t.defaultChannel != "discord" {
		return ErrorResult("discord_delete_message can only be used in Discord channels")
	}

	channelID, _ := args["channel_id"].(string)
	messageID, _ := args["message_id"].(string)
	reason, _ := args["reason"].(string)
	logChannelID, _ := args["log_channel_id"].(string)

	if channelID == "" || messageID == "" {
		return ErrorResult("channel_id and message_id are required")
	}
	if reason == "" {
		reason = "No reason provided"
	}

	if t.deleteCallback == nil {
		return ErrorResult("Discord delete not configured")
	}

	if err := t.deleteCallback(channelID, messageID); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Failed to delete message: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	// Auto-log to mod-log
	if t.sendCallback != nil && logChannelID != "" {
		logMsg := fmt.Sprintf("🗑️ **Message Deleted**\n**Channel:** <#%s>\n**Reason:** %s\n**Actioned by:** LuckyClaw", channelID, reason)
		_ = t.sendCallback("discord", logChannelID, logMsg)
	}

	return &ToolResult{
		ForLLM: fmt.Sprintf("Message %s deleted from channel %s. Reason: %s", messageID, channelID, reason),
	}
}

// --- discord_timeout_user tool ---

type DiscordTimeoutUserTool struct {
	timeoutCallback TimeoutUserCallback
	sendCallback    SendMessageCallback
	defaultChannel  string
	defaultChatID   string
	defaultGuildID  string
}

func NewDiscordTimeoutUserTool() *DiscordTimeoutUserTool {
	return &DiscordTimeoutUserTool{}
}

func (t *DiscordTimeoutUserTool) Name() string {
	return "discord_timeout_user"
}

func (t *DiscordTimeoutUserTool) Description() string {
	return "Timeout (mute) a user in a Discord server for a specified duration. Use this for rule violations like hate speech or spam. Maximum duration is 28 days."
}

func (t *DiscordTimeoutUserTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "The Discord user ID to timeout. If the user mentions them like <@123456>, the ID is 123456. EXTRACT THIS AUTOMATICALLY from the raw message and DO NOT ask the user to provide it.",
			},
			"duration_minutes": map[string]interface{}{
				"type":        "number",
				"description": "Timeout duration in minutes (max 40320 = 28 days)",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Reason for the timeout (logged to mod-log)",
			},
			"log_channel_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: Channel ID to send a moderation log to (e.g., mod-log channel)",
			},
		},
		"required": []string{"user_id", "duration_minutes", "reason"},
	}
}

func (t *DiscordTimeoutUserTool) SetContext(channel, chatID string) {
	t.defaultChannel = channel
	t.defaultChatID = chatID
}

// SetGuildID stores the guild ID from message metadata.
// Called from updateToolContexts — never relying on the LLM to supply it.
func (t *DiscordTimeoutUserTool) SetGuildID(guildID string) {
	t.defaultGuildID = guildID
}

func (t *DiscordTimeoutUserTool) SetTimeoutCallback(cb TimeoutUserCallback) {
	t.timeoutCallback = cb
}

func (t *DiscordTimeoutUserTool) SetSendCallback(cb SendMessageCallback) {
	t.sendCallback = cb
}

func (t *DiscordTimeoutUserTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	if t.defaultChannel != "discord" {
		return ErrorResult("discord_timeout_user can only be used in Discord channels")
	}

	guildID := t.defaultGuildID // Always use context guild_id
	rawUserID, _ := args["user_id"].(string)
	reason, _ := args["reason"].(string)
	logChannelID, _ := args["log_channel_id"].(string)

	// Clean up user ID if the LLM passed a raw mention like <@123456>
	userID := strings.TrimPrefix(strings.TrimSuffix(rawUserID, ">"), "<@")
	userID = strings.TrimPrefix(userID, "!") // Sometimes it's <@!123456>

	// Handle duration_minutes as float64 (JSON numbers are float64)
	durationMinutes := 0.0
	if v, ok := args["duration_minutes"].(float64); ok {
		durationMinutes = v
	}

	if guildID == "" {
		return ErrorResult("guild_id not available in context")
	}
	if userID == "" {
		return ErrorResult("user_id is required")
	}
	if durationMinutes <= 0 {
		return ErrorResult("duration_minutes must be a positive number")
	}
	if reason == "" {
		reason = "No reason provided"
	}

	// Cap at 28 days (Discord maximum)
	const maxMinutes = 28 * 24 * 60
	if durationMinutes > maxMinutes {
		durationMinutes = maxMinutes
	}

	if t.timeoutCallback == nil {
		return ErrorResult("Discord timeout not configured")
	}

	until := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	if err := t.timeoutCallback(guildID, userID, until); err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Failed to timeout user: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	// Auto-log to mod-log
	if t.sendCallback != nil && logChannelID != "" {
		logMsg := fmt.Sprintf("⚠️ **User Timeout**\n**User:** <@%s>\n**Duration:** %.0f minutes\n**Reason:** %s\n**Actioned by:** LuckyClaw", userID, durationMinutes, reason)
		_ = t.sendCallback("discord", logChannelID, logMsg)
	}

	return &ToolResult{
		ForLLM: fmt.Sprintf("User %s timed out for %.0f minutes in guild %s. Reason: %s",
			userID, durationMinutes, guildID, reason),
	}
}

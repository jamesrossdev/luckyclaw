package channels

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jamesrossdev/luckyclaw/pkg/bus"
	"github.com/jamesrossdev/luckyclaw/pkg/config"
	"github.com/jamesrossdev/luckyclaw/pkg/logger"
	"github.com/jamesrossdev/luckyclaw/pkg/utils"
	"github.com/jamesrossdev/luckyclaw/pkg/voice"
)

const maxTimeoutDays = 28 // Discord maximum timeout duration

const (
	transcriptionTimeout = 30 * time.Second
	sendTimeout          = 10 * time.Second
)

type DiscordChannel struct {
	*BaseChannel
	session     *discordgo.Session
	config      config.DiscordConfig
	transcriber *voice.GroqTranscriber
	ctx         context.Context
}

func NewDiscordChannel(cfg config.DiscordConfig, bus *bus.MessageBus) (*DiscordChannel, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	base := NewBaseChannel("discord", cfg, bus, cfg.AllowFrom)

	return &DiscordChannel{
		BaseChannel: base,
		session:     session,
		config:      cfg,
		transcriber: nil,
		ctx:         context.Background(),
	}, nil
}

func (c *DiscordChannel) SetTranscriber(transcriber *voice.GroqTranscriber) {
	c.transcriber = transcriber
}

func (c *DiscordChannel) getContext() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func (c *DiscordChannel) Start(ctx context.Context) error {
	logger.InfoC("discord", "Starting Discord bot")

	c.ctx = ctx
	c.session.AddHandler(c.handleMessage)

	if err := c.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	c.SetRunning(true)

	botUser, err := c.session.User("@me")
	if err != nil {
		return fmt.Errorf("failed to get bot user: %w", err)
	}
	logger.InfoCF("discord", "Discord bot connected", map[string]any{
		"username": botUser.Username,
		"user_id":  botUser.ID,
	})

	return nil
}

func (c *DiscordChannel) Stop(ctx context.Context) error {
	logger.InfoC("discord", "Stopping Discord bot")
	c.SetRunning(false)

	if err := c.session.Close(); err != nil {
		return fmt.Errorf("failed to close discord session: %w", err)
	}

	return nil
}

func (c *DiscordChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("discord bot not running")
	}

	channelID := msg.ChatID
	if channelID == "" {
		return fmt.Errorf("channel ID is empty")
	}

	runes := []rune(msg.Content)
	if len(runes) == 0 {
		return nil
	}

	chunks := splitMessage(msg.Content, 1500) // Discord has a limit of 2000 characters per message, leave 500 for natural split e.g. code blocks

	for i, chunk := range chunks {
		// Only quote-reply on the first chunk
		replyToID := ""
		if i == 0 {
			replyToID = msg.ReplyToID
		}
		if err := c.sendChunk(ctx, channelID, chunk, replyToID); err != nil {
			return err
		}
	}

	return nil
}

// splitMessage splits long messages into chunks, preserving code block integrity
// Uses natural boundaries (newlines, spaces) and extends messages slightly to avoid breaking code blocks
func splitMessage(content string, limit int) []string {
	var messages []string

	for len(content) > 0 {
		if len(content) <= limit {
			messages = append(messages, content)
			break
		}

		msgEnd := limit

		// Find natural split point within the limit
		msgEnd = findLastNewline(content[:limit], 200)
		if msgEnd <= 0 {
			msgEnd = findLastSpace(content[:limit], 100)
		}
		if msgEnd <= 0 {
			msgEnd = limit
		}

		// Check if this would end with an incomplete code block
		candidate := content[:msgEnd]
		unclosedIdx := findLastUnclosedCodeBlock(candidate)

		if unclosedIdx >= 0 {
			// Message would end with incomplete code block
			// Try to extend to include the closing ``` (with some buffer)
			extendedLimit := limit + 500 // Allow 500 char buffer for code blocks
			if len(content) > extendedLimit {
				closingIdx := findNextClosingCodeBlock(content, msgEnd)
				if closingIdx > 0 && closingIdx <= extendedLimit {
					// Extend to include the closing ```
					msgEnd = closingIdx
				} else {
					// Can't find closing, split before the code block
					msgEnd = findLastNewline(content[:unclosedIdx], 200)
					if msgEnd <= 0 {
						msgEnd = findLastSpace(content[:unclosedIdx], 100)
					}
					if msgEnd <= 0 {
						msgEnd = unclosedIdx
					}
				}
			} else {
				// Remaining content fits within extended limit
				msgEnd = len(content)
			}
		}

		if msgEnd <= 0 {
			msgEnd = limit
		}

		messages = append(messages, content[:msgEnd])
		content = strings.TrimSpace(content[msgEnd:])
	}

	return messages
}

// findLastUnclosedCodeBlock finds the last opening ``` that doesn't have a closing ```
// Returns the position of the opening ``` or -1 if all code blocks are complete
func findLastUnclosedCodeBlock(text string) int {
	count := 0
	lastOpenIdx := -1

	for i := 0; i < len(text); i++ {
		if i+2 < len(text) && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			if count == 0 {
				lastOpenIdx = i
			}
			count++
			i += 2
		}
	}

	// If odd number of ``` markers, last one is unclosed
	if count%2 == 1 {
		return lastOpenIdx
	}
	return -1
}

// findNextClosingCodeBlock finds the next closing ``` starting from a position
// Returns the position after the closing ``` or -1 if not found
func findNextClosingCodeBlock(text string, startIdx int) int {
	for i := startIdx; i < len(text); i++ {
		if i+2 < len(text) && text[i] == '`' && text[i+1] == '`' && text[i+2] == '`' {
			return i + 3
		}
	}
	return -1
}

// findLastNewline finds the last newline character within the last N characters
// Returns the position of the newline or -1 if not found
func findLastNewline(s string, searchWindow int) int {
	searchStart := len(s) - searchWindow
	if searchStart < 0 {
		searchStart = 0
	}
	for i := len(s) - 1; i >= searchStart; i-- {
		if s[i] == '\n' {
			return i
		}
	}
	return -1
}

// findLastSpace finds the last space character within the last N characters
// Returns the position of the space or -1 if not found
func findLastSpace(s string, searchWindow int) int {
	searchStart := len(s) - searchWindow
	if searchStart < 0 {
		searchStart = 0
	}
	for i := len(s) - 1; i >= searchStart; i-- {
		if s[i] == ' ' || s[i] == '\t' {
			return i
		}
	}
	return -1
}

func (c *DiscordChannel) sendChunk(ctx context.Context, channelID, content, replyToID string) error {
	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		if replyToID != "" {
			// Send with quote-reply reference
			_, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Content: content,
				Reference: &discordgo.MessageReference{
					MessageID: replyToID,
					ChannelID: channelID,
				},
				AllowedMentions: &discordgo.MessageAllowedMentions{
					RepliedUser: true,
				},
			})
			done <- err
		} else {
			_, err := c.session.ChannelMessageSend(channelID, content)
			done <- err
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send discord message: %w", err)
		}
		return nil
	case <-sendCtx.Done():
		return fmt.Errorf("send message timeout: %w", sendCtx.Err())
	}
}

// appendContent safely appends a suffix to existing text, joining with a newline.
func appendContent(content, suffix string) string {
	if content == "" {
		return suffix
	}
	return content + "\n" + suffix
}

func (c *DiscordChannel) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil || m.Author == nil {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	// DM filter: drop direct messages when disable_dms is set
	if c.config.DisableDMs && m.GuildID == "" {
		logger.DebugCF("discord", "DM ignored (disable_dms=true)", map[string]any{
			"user_id": m.Author.ID,
		})
		return
	}

	// In server channels (GuildID != ""), only respond when @mentioned or when
	// the message is a reply to one of the bot's own messages.
	if m.GuildID != "" {
		botID := s.State.User.ID
		mentioned := false
		for _, u := range m.Mentions {
			if u.ID == botID {
				mentioned = true
				break
			}
		}
		// Also respond when user replies to a bot message
		if !mentioned && m.ReferencedMessage != nil && m.ReferencedMessage.Author != nil {
			if m.ReferencedMessage.Author.ID == botID {
				mentioned = true
			}
		}
		// Also respond when the bot's own managed role is @mentioned
		if !mentioned && len(m.MentionRoles) > 0 {
			if botMember, err := s.State.Member(m.GuildID, botID); err == nil {
				for _, botRoleID := range botMember.Roles {
					for _, mentionedRoleID := range m.MentionRoles {
						if botRoleID == mentionedRoleID {
							mentioned = true
							break
						}
					}
					if mentioned {
						break
					}
				}
			}
		}
		if !mentioned {
			return
		}
	}

	// Persistent typing indicator: refresh every 8s for up to 90s so the
	// "LuckyClaw is typing..." badge stays visible during long LLM calls.
	typingCtx, stopTyping := context.WithTimeout(context.Background(), 90*time.Second)
	go func() {
		for {
			select {
			case <-typingCtx.Done():
				return
			default:
				_ = c.session.ChannelTyping(m.ChannelID)
				time.Sleep(8 * time.Second)
			}
		}
	}()
	defer stopTyping()

	if !c.IsAllowed(m.Author.ID) {
		logger.DebugCF("discord", "Message rejected by allowlist", map[string]any{
			"user_id": m.Author.ID,
		})
		return
	}

	senderID := m.Author.ID
	senderName := m.Author.Username
	if m.Author.Discriminator != "" && m.Author.Discriminator != "0" {
		senderName += "#" + m.Author.Discriminator
	}

	content := m.Content

	// Read quoted/referenced message — ONLY for snitch flow (quoting another user's
	// message). Skip when the quoted message is from the bot itself, since the bot
	// already has its own responses in session history.
	if m.ReferencedMessage != nil && m.ReferencedMessage.Author != nil {
		if m.ReferencedMessage.Author.ID != s.State.User.ID {
			quotedAuthor := m.ReferencedMessage.Author.Username
			quotedContent := m.ReferencedMessage.Content
			if quotedContent != "" {
				content = fmt.Sprintf("[Quoted message from %s (user_id: %s, message_id: %s): \"%s\"]\n%s",
					quotedAuthor, m.ReferencedMessage.Author.ID, m.ReferencedMessage.ID, quotedContent, content)
			}
		}
	}

	mediaPaths := make([]string, 0, len(m.Attachments))
	localFiles := make([]string, 0, len(m.Attachments))

	// Temp file cleanup is handled centrally by loop.go after LLM processing.
	// Do NOT defer os.Remove here — it races with the provider reading the file.

	for _, attachment := range m.Attachments {
		isAudio := utils.IsAudioFile(attachment.Filename, attachment.ContentType)

		if isAudio {
			localPath := c.downloadAttachment(attachment.URL, attachment.Filename)
			if localPath != "" {
				localFiles = append(localFiles, localPath)

				transcribedText := ""
				if c.transcriber != nil && c.transcriber.IsAvailable() {
					ctx, cancel := context.WithTimeout(c.getContext(), transcriptionTimeout)
					result, err := c.transcriber.Transcribe(ctx, localPath)
					cancel() // Release context resources immediately to avoid leak in loop

					if err != nil {
						logger.ErrorCF("discord", "Voice transcription failed", map[string]any{
							"error": err.Error(),
						})
						transcribedText = fmt.Sprintf("[audio: %s (transcription failed)]", attachment.Filename)
					} else {
						transcribedText = fmt.Sprintf("[audio transcription: %s]", result.Text)
						logger.DebugCF("discord", "Audio transcribed successfully", map[string]any{
							"text": result.Text,
						})
					}
				} else {
					transcribedText = fmt.Sprintf("[audio: %s]", attachment.Filename)
				}

				content = appendContent(content, transcribedText)
			} else {
				logger.WarnCF("discord", "Failed to download audio attachment", map[string]any{
					"url":      attachment.URL,
					"filename": attachment.Filename,
				})
				mediaPaths = append(mediaPaths, attachment.URL)
				content = appendContent(content, fmt.Sprintf("[attachment: %s]", attachment.URL))
			}
		} else {
			mediaPaths = append(mediaPaths, attachment.URL)
			content = appendContent(content, fmt.Sprintf("[attachment: %s]", attachment.URL))
		}
	}

	if content == "" && len(mediaPaths) == 0 {
		return
	}

	if content == "" {
		content = "[media only]"
	}

	logger.DebugCF("discord", "Received message", map[string]any{
		"sender_name": senderName,
		"sender_id":   senderID,
		"preview":     utils.Truncate(content, 50),
	})

	metadata := map[string]string{
		"message_id":   m.ID,
		"user_id":      senderID,
		"username":     m.Author.Username,
		"display_name": senderName,
		"guild_id":     m.GuildID,
		"channel_id":   m.ChannelID,
		"is_dm":        fmt.Sprintf("%t", m.GuildID == ""),
	}

	// Resolve sender's Discord roles to names so the LLM can see them.
	if m.Member != nil && len(m.Member.Roles) > 0 {
		var roleNames []string
		for _, roleID := range m.Member.Roles {
			if role, err := s.State.Role(m.GuildID, roleID); err == nil {
				roleNames = append(roleNames, role.Name)
			}
		}
		if len(roleNames) > 0 {
			metadata["sender_roles"] = strings.Join(roleNames, ", ")
		}
	}

	// Trigger "is typing" indicator while the agent processes this message
	s.ChannelTyping(m.ChannelID)

	c.HandleMessage(senderID, m.ChannelID, content, mediaPaths, metadata, "", "")
}

func (c *DiscordChannel) downloadAttachment(url, filename string) string {
	return utils.DownloadFile(url, filename, utils.DownloadOptions{
		LoggerPrefix: "discord",
	})
}

// DeleteMessage deletes a message in the given channel.
// Used by the discord_delete_message tool.
func (c *DiscordChannel) DeleteMessage(channelID, messageID string) error {
	if !c.IsRunning() {
		return fmt.Errorf("discord bot not running")
	}
	return c.session.ChannelMessageDelete(channelID, messageID)
}

// TimeoutUser applies a timeout to a guild member until the specified time.
// Used by the discord_timeout_user tool.
func (c *DiscordChannel) TimeoutUser(guildID, userID string, until time.Time) error {
	if !c.IsRunning() {
		return fmt.Errorf("discord bot not running")
	}
	// Discord maximum timeout is 28 days
	maxUntil := time.Now().Add(time.Duration(maxTimeoutDays) * 24 * time.Hour)
	if until.After(maxUntil) {
		until = maxUntil
	}
	return c.session.GuildMemberTimeout(guildID, userID, &until)
}

// Session returns the underlying discordgo session.
// Used to set up moderation tool callbacks.
func (c *DiscordChannel) Session() *discordgo.Session {
	return c.session
}

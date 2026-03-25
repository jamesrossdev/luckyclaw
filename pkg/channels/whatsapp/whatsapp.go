package whatsapp

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"

	"golang.org/x/image/draw"

	"github.com/jamesrossdev/luckyclaw/pkg/bus"
	"github.com/jamesrossdev/luckyclaw/pkg/channels"
	"github.com/jamesrossdev/luckyclaw/pkg/config"
	"github.com/jamesrossdev/luckyclaw/pkg/logger"
)

const (
	sqliteDriver   = "sqlite"
	whatsappDBName = "store.db"

	reconnectInitial    = 5 * time.Second
	reconnectMax        = 5 * time.Minute
	reconnectMultiplier = 2.0
)

// compressImage strictly downsizes uploaded multi-modal image buffers
func compressImage(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if w <= 512 && h <= 512 {
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60})
		return buf.Bytes(), err
	}

	var newW, newH int
	if w > h {
		newW = 512
		newH = h * 512 / w
	} else {
		newH = 512
		newW = w * 512 / h
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 60})
	return buf.Bytes(), err
}

// WhatsAppChannel implements the WhatsApp channel using whatsmeow (in-process, no external bridge).
type WhatsAppChannel struct {
	*channels.BaseChannel
	config       config.WhatsAppConfig
	storePath    string
	client       *whatsmeow.Client
	container    *sqlstore.Container
	db           *sql.DB
	mu           sync.Mutex
	runCtx       context.Context
	runCancel    context.CancelFunc
	reconnectMu  sync.Mutex
	reconnecting bool
	stopping     atomic.Bool    // set once Stop begins; prevents new wg.Add calls
	wg           sync.WaitGroup // tracks background goroutines (QR handler, reconnect)
	rateLimiter  map[string][]time.Time
	rateLimitMu  sync.Mutex
}

// NewWhatsAppChannel creates a WhatsApp channel that uses whatsmeow for connection.
func NewWhatsAppChannel(
	cfg config.WhatsAppConfig,
	bus *bus.MessageBus,
	storePath string,
) (*WhatsAppChannel, error) {
	base := channels.NewBaseChannel("whatsapp", cfg, bus, cfg.AllowFrom)
	if storePath == "" {
		storePath = "whatsapp"
	}
	c := &WhatsAppChannel{
		BaseChannel: base,
		config:      cfg,
		storePath:   storePath,
		rateLimiter: make(map[string][]time.Time),
	}
	return c, nil
}

func (c *WhatsAppChannel) Start(ctx context.Context) error {
	logger.InfoCF("whatsapp", "Starting WhatsApp native channel (whatsmeow)", map[string]any{"store": c.storePath})

	c.reconnectMu.Lock()
	c.stopping.Store(false)
	c.reconnecting = false
	c.reconnectMu.Unlock()

	if err := os.MkdirAll(c.storePath, 0o700); err != nil {
		return fmt.Errorf("create session store dir: %w", err)
	}

	dbPath := filepath.Join(c.storePath, whatsappDBName)
	connStr := "file:" + dbPath + "?_foreign_keys=on"

	db, err := sql.Open(sqliteDriver, connStr)
	if err != nil {
		return fmt.Errorf("open whatsapp store: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	waLogger := waLog.Stdout("WhatsApp", "WARN", true)
	container := sqlstore.NewWithDB(db, sqliteDriver, waLogger)
	c.mu.Lock()
	c.container = container
	c.db = db
	c.mu.Unlock()
	if err = container.Upgrade(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("open whatsapp store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		_ = container.Close()
		return fmt.Errorf("get device store: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, waLogger)
	c.runCtx, c.runCancel = context.WithCancel(ctx)

	client.AddEventHandler(c.eventHandler)

	c.mu.Lock()
	c.container = container
	c.client = client
	c.mu.Unlock()

	startOK := false
	defer func() {
		if startOK {
			return
		}
		c.runCancel()
		client.Disconnect()
		c.mu.Lock()
		c.client = nil
		c.container = nil
		c.mu.Unlock()
		_ = container.Close()
	}()

	if client.Store.ID == nil {
		qrChan, err := client.GetQRChannel(c.runCtx)
		if err != nil {
			return fmt.Errorf("get QR channel: %w", err)
		}
		if err := client.Connect(); err != nil {
			return fmt.Errorf("connect: %w", err)
		}

		c.reconnectMu.Lock()
		if c.stopping.Load() {
			c.reconnectMu.Unlock()
			return fmt.Errorf("channel stopped during QR setup")
		}
		c.wg.Add(1)
		c.reconnectMu.Unlock()
		go func() {
			defer c.wg.Done()
			for {
				select {
				case <-c.runCtx.Done():
					return
				case evt, ok := <-qrChan:
					if !ok {
						return
					}
					if evt.Event == "code" {
						fmt.Println("\n  🦞 WhatsApp Pairing Required")
						fmt.Println("  ──────────────────────────────")
						fmt.Println("  Scan this QR code with WhatsApp (Linked Devices):")
						qrterminal.GenerateWithConfig(evt.Code, qrterminal.Config{
							Level:      qrterminal.L,
							Writer:     os.Stdout,
							HalfBlocks: true,
						})
					} else {
						logger.InfoCF("whatsapp", "WhatsApp login event", map[string]any{"event": evt.Event})
					}
				}
			}
		}()
	} else {
		if err := client.Connect(); err != nil {
			return fmt.Errorf("connect: %w", err)
		}
	}

	startOK = true
	c.BaseChannel.SetRunning(true)
	logger.InfoC("whatsapp", "WhatsApp channel connected")
	return nil
}

func (c *WhatsAppChannel) Stop(ctx context.Context) error {
	logger.InfoC("whatsapp", "Stopping WhatsApp channel")

	c.reconnectMu.Lock()
	c.stopping.Store(true)
	c.reconnectMu.Unlock()

	if c.runCancel != nil {
		c.runCancel()
	}

	c.mu.Lock()
	client := c.client
	container := c.container
	db := c.db
	c.mu.Unlock()

	if client != nil {
		client.Disconnect()
	}

	if db != nil {
		db.Close()
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		logger.WarnC("whatsapp", fmt.Sprintf("Stop context canceled before all goroutines finished: %v", ctx.Err()))
	}

	c.mu.Lock()
	c.client = nil
	c.container = nil
	c.mu.Unlock()

	if container != nil {
		_ = container.Close()
	}
	c.BaseChannel.SetRunning(false)
	return nil
}

func (c *WhatsAppChannel) eventHandler(evt any) {
	switch evt.(type) {
	case *events.Message:
		c.handleIncoming(evt.(*events.Message))
	case *events.Disconnected:
		logger.InfoCF("whatsapp", "WhatsApp disconnected, will attempt reconnection", nil)
		c.reconnectMu.Lock()
		if c.reconnecting {
			c.reconnectMu.Unlock()
			return
		}
		if c.stopping.Load() {
			c.reconnectMu.Unlock()
			return
		}
		c.reconnecting = true
		c.wg.Add(1)
		c.reconnectMu.Unlock()
		go func() {
			defer c.wg.Done()
			c.reconnectWithBackoff()
		}()
	}
}

func (c *WhatsAppChannel) reconnectWithBackoff() {
	defer func() {
		c.reconnectMu.Lock()
		c.reconnecting = false
		c.reconnectMu.Unlock()
	}()

	backoff := reconnectInitial
	for {
		select {
		case <-c.runCtx.Done():
			return
		default:
		}

		c.mu.Lock()
		client := c.client
		c.mu.Unlock()
		if client == nil {
			return
		}

		logger.InfoCF("whatsapp", "WhatsApp reconnecting", map[string]any{"backoff": backoff.String()})
		err := client.Connect()
		if err == nil {
			logger.InfoC("whatsapp", "WhatsApp reconnected")
			return
		}

		logger.WarnCF("whatsapp", "WhatsApp reconnect failed", map[string]any{"error": err.Error()})

		select {
		case <-c.runCtx.Done():
			return
		case <-time.After(backoff):
			if backoff < reconnectMax {
				next := time.Duration(float64(backoff) * reconnectMultiplier)
				if next > reconnectMax {
					next = reconnectMax
				}
				backoff = next
			}
		}
	}
}

func (c *WhatsAppChannel) handleIncoming(evt *events.Message) {
	if evt.Message == nil {
		return
	}
	senderID := evt.Info.Sender.User
	chatID := evt.Info.Chat.String()

	logger.InfoCF(
		"whatsapp",
		"RAW WhatsApp message received",
		map[string]any{"sender_user": senderID, "chat_string": chatID, "is_from_me": evt.Info.IsFromMe},
	)

	// If it's from me, it means I sent it from another device (like my phone)
	// while the bot is also logged in. We should handle this so "Note to self" works.
	if evt.Info.IsFromMe && evt.Info.Chat.User != evt.Info.Sender.User {
		// It was sent TO someone else by me, ignore it to prevent replying to our own outgoing messages
		return
	}

	// v0.2.2-rc6 Rate Limiter (Disabled by default)
	const maxMessagesPerMinute = 5
	const rateLimitEnabled = false
	if rateLimitEnabled {
		c.rateLimitMu.Lock()
		now := time.Now()
		var recent []time.Time
		for _, t := range c.rateLimiter[senderID] {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}
		recent = append(recent, now)
		c.rateLimiter[senderID] = recent
		c.rateLimitMu.Unlock()

		if len(recent) > maxMessagesPerMinute {
			logger.WarnCF("whatsapp", "Rate limit exceeded, dropping message", map[string]any{"sender": senderID})
			return
		}
	}

	content := evt.Message.GetConversation()
	if content == "" && evt.Message.ExtendedTextMessage != nil {
		content = evt.Message.ExtendedTextMessage.GetText()
	}

	var mediaPaths []string
	var localFiles []string

	var data []byte
	var err error
	var fileLength uint64
	var mimetype string
	var filename string

	c.mu.Lock()
	pclient := c.client
	c.mu.Unlock()

	if pclient != nil {
		if evt.Message.DocumentMessage != nil {
			fileLength = evt.Message.DocumentMessage.GetFileLength()
			mimetype = evt.Message.DocumentMessage.GetMimetype()
			filename = evt.Message.DocumentMessage.GetFileName()
			if content == "" {
				content = evt.Message.DocumentMessage.GetCaption()
			}
			if fileLength <= 5_000_000 {
				data, err = pclient.Download(context.Background(), evt.Message.DocumentMessage)
			}
		} else if evt.Message.ImageMessage != nil {
			fileLength = evt.Message.ImageMessage.GetFileLength()
			mimetype = evt.Message.ImageMessage.GetMimetype()
			if content == "" {
				content = evt.Message.ImageMessage.GetCaption()
			}
			if fileLength <= 5_000_000 {
				data, err = pclient.Download(context.Background(), evt.Message.ImageMessage)
			}
		} else if evt.Message.AudioMessage != nil {
			fileLength = evt.Message.AudioMessage.GetFileLength()
			mimetype = evt.Message.AudioMessage.GetMimetype()
			if fileLength <= 5_000_000 {
				data, err = pclient.Download(context.Background(), evt.Message.AudioMessage)
			}
		} else if evt.Message.VideoMessage != nil {
			fileLength = evt.Message.VideoMessage.GetFileLength()
			mimetype = evt.Message.VideoMessage.GetMimetype()
			if content == "" {
				content = evt.Message.VideoMessage.GetCaption()
			}
			if fileLength <= 5_000_000 {
				data, err = pclient.Download(context.Background(), evt.Message.VideoMessage)
			}
		}

		if fileLength > 5_000_000 {
			logger.WarnCF("whatsapp", "Dropping media exceeding 5MB limit", map[string]any{"sender": senderID, "size": fileLength})
			if content == "" {
				return
			}
		} else if len(data) > 0 {
			lowerFilename := strings.ToLower(filename)
			isText := strings.HasPrefix(mimetype, "text/") ||
				strings.HasPrefix(mimetype, "application/json") ||
				strings.HasSuffix(lowerFilename, ".txt") ||
				strings.HasSuffix(lowerFilename, ".md") ||
				strings.HasSuffix(lowerFilename, ".csv") ||
				strings.HasSuffix(lowerFilename, ".json") ||
				strings.HasSuffix(lowerFilename, ".log") ||
				strings.HasSuffix(lowerFilename, ".yaml") ||
				strings.HasSuffix(lowerFilename, ".yml")

			if isText {
				logger.InfoCF("whatsapp", "Ingested plain-text file natively", map[string]any{"filename": filename, "size": len(data)})
				if filename == "" {
					filename = "attached_file.txt"
				}
				content = fmt.Sprintf("[Attached File: %s]\n\n%s\n\n%s", filename, string(data), content)
			} else {
				if strings.HasPrefix(mimetype, "image/") || evt.Message.ImageMessage != nil {
					if compressedData, errCompress := compressImage(data); errCompress == nil {
						data = compressedData
					} else {
						logger.WarnCF("whatsapp", "Image compression failed", map[string]any{"error": errCompress.Error()})
					}
				}

				if tmpFile, err := os.CreateTemp("", "wa-media-*"); err == nil {
					tmpFile.Write(data)
					tmpFile.Close()
					localFiles = append(localFiles, tmpFile.Name())
					mediaPaths = append(mediaPaths, tmpFile.Name())
					content = fmt.Sprintf("[media loaded]\n%s", content)
				} else {
					logger.WarnCF("whatsapp", "Failed to save media", map[string]any{"error": err.Error()})
				}
			}
		} else if err != nil {
			logger.WarnCF("whatsapp", "Failed to download media", map[string]any{"error": err.Error()})
		}
	}

	// --- v0.2.2-rc4: Group Spam Filtering & Contextual Quotes ---
	var isGroup = strings.HasSuffix(chatID, "@g.us")
	var isMentioned = false
	var isReplyToBot = false
	var botJID string
	var botLID string

	c.mu.Lock()
	if c.client != nil && c.client.Store != nil {
		if c.client.Store.ID != nil {
			botJID = c.client.Store.ID.User
		}
		// Fetch the cryptographic Local ID generated by Meta Multi-Device
		lid := c.client.Store.GetLID()
		if lid.User != "" {
			botLID = lid.User
		}
	}
	c.mu.Unlock()

	var quotedText string
	var quotedSender string

	var ctxInfo *waE2E.ContextInfo
	if evt.Message.ExtendedTextMessage != nil {
		ctxInfo = evt.Message.ExtendedTextMessage.ContextInfo
	} else if evt.Message.ImageMessage != nil {
		ctxInfo = evt.Message.ImageMessage.ContextInfo
	} else if evt.Message.VideoMessage != nil {
		ctxInfo = evt.Message.VideoMessage.ContextInfo
	} else if evt.Message.DocumentMessage != nil {
		ctxInfo = evt.Message.DocumentMessage.ContextInfo
	} else if evt.Message.AudioMessage != nil {
		ctxInfo = evt.Message.AudioMessage.ContextInfo
	}

	if ctxInfo != nil {

		for _, mentioned := range ctxInfo.GetMentionedJID() {
			if (botJID != "" && strings.HasPrefix(mentioned, botJID+"@")) ||
				(botLID != "" && strings.HasPrefix(mentioned, botLID+"@")) {
				isMentioned = true
				break
			}
		}

		if ctxInfo.GetParticipant() != "" {
			participant := ctxInfo.GetParticipant()
			if (botJID != "" && strings.HasPrefix(participant, botJID+"@")) ||
				(botLID != "" && strings.HasPrefix(participant, botLID+"@")) {
				isReplyToBot = true
			}

			if ctxInfo.QuotedMessage != nil {
				qMsg := ctxInfo.QuotedMessage
				if qMsg.GetConversation() != "" {
					quotedText = qMsg.GetConversation()
				} else if qMsg.ExtendedTextMessage != nil {
					quotedText = qMsg.ExtendedTextMessage.GetText()
				}

				if quotedText != "" {
					qSender := participant
					if parts := strings.Split(qSender, "@"); len(parts) > 0 {
						quotedSender = parts[0]
					} else {
						quotedSender = qSender
					}
				}
			}
		}

		// Optional debug log
		if isGroup {
			logger.InfoCF("whatsapp", "[DEBUG] Mention Tracking", map[string]any{
				"botJID":   botJID,
				"botLID":   botLID,
				"mentions": ctxInfo.GetMentionedJID(),
				"replyTo":  ctxInfo.GetParticipant(),
			})
		}
	}

	if isGroup {
		// Silently drop if not interacting with the bot
		if !isMentioned && !isReplyToBot {
			return
		}
	}

	if quotedText != "" {
		content = fmt.Sprintf("[Quoted from %s: %q]\n\n%s", quotedSender, quotedText, content)
	}
	// ------------------------------------------------------------

	if content == "" && len(mediaPaths) == 0 {
		logger.InfoCF("whatsapp", "Dropping empty message", map[string]any{"sender": senderID})
		return
	}

	metadata := make(map[string]string)
	metadata["message_id"] = evt.Info.ID
	if evt.Info.PushName != "" {
		metadata["user_name"] = evt.Info.PushName
	}

	if pclient != nil {
		targetJID := evt.Info.Chat
		// Run psychological triggers in a separate goroutine so we don't block the BaseChannel defer loop
		go func(tJID types.JID, msgID types.MessageID, tStamp time.Time, sender types.JID) {
			pclient.MarkRead(context.Background(), []types.MessageID{msgID}, tStamp, tJID, sender)
			time.Sleep(1 * time.Second)
			pclient.SendChatPresence(context.Background(), tJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)

			// Hold the typing indicator for roughly 3 seconds to simulate human reading,
			// then drop it natively to prevent looping.
			time.Sleep(3 * time.Second)
			pclient.SendChatPresence(context.Background(), tJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
		}(targetJID, evt.Info.ID, evt.Info.Timestamp, evt.Info.Sender)
	}

	// BaseChannel.HandleMessage(senderID, chatID, content, media []string, metadata map[string]string)
	c.HandleMessage(senderID, chatID, content, mediaPaths, metadata)
}

func (c *WhatsAppChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("whatsapp channel not running")
	}

	c.mu.Lock()
	client := c.client
	c.mu.Unlock()

	if client == nil || !client.IsConnected() {
		return fmt.Errorf("whatsapp connection not established")
	}

	if client.Store.ID == nil {
		return fmt.Errorf("whatsapp not yet paired")
	}

	to, err := types.ParseJID(msg.ChatID)
	if err != nil {
		// If it doesn't contain @, assume it's a phone number and add the user server
		if !strings.Contains(msg.ChatID, "@") {
			to = types.NewJID(msg.ChatID, types.DefaultUserServer)
		} else {
			return fmt.Errorf("invalid chat id %q: %w", msg.ChatID, err)
		}
	}

	waMsg := &waE2E.Message{
		Conversation: proto.String(msg.Content),
	}

	if _, err = client.SendMessage(ctx, to, waMsg); err != nil {
		return fmt.Errorf("whatsapp send failed: %w", err)
	}
	return nil
}

func (c *WhatsAppChannel) SetRunning(running bool) {
	c.BaseChannel.SetRunning(running)
}

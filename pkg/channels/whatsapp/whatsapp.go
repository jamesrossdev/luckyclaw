package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
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

// WhatsAppChannel implements the WhatsApp channel using whatsmeow (in-process, no external bridge).
type WhatsAppChannel struct {
	*channels.BaseChannel
	config       config.WhatsAppConfig
	storePath    string
	client       *whatsmeow.Client
	container    *sqlstore.Container
	mu           sync.Mutex
	runCtx       context.Context
	runCancel    context.CancelFunc
	reconnectMu  sync.Mutex
	reconnecting bool
	stopping     atomic.Bool    // set once Stop begins; prevents new wg.Add calls
	wg           sync.WaitGroup // tracks background goroutines (QR handler, reconnect)
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
	c.mu.Unlock()

	if client != nil {
		client.Disconnect()
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
	content := evt.Message.GetConversation()
	if content == "" && evt.Message.ExtendedTextMessage != nil {
		content = evt.Message.ExtendedTextMessage.GetText()
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

	if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.ContextInfo != nil {
		ctxInfo := evt.Message.ExtendedTextMessage.ContextInfo
		
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
				"botJID": botJID,
				"botLID": botLID,
				"mentions": ctxInfo.GetMentionedJID(),
				"replyTo": ctxInfo.GetParticipant(),
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
	
	if content == "" {
		logger.InfoCF("whatsapp", "Dropping empty message", map[string]any{"sender": senderID})
		return
	}

	metadata := make(map[string]string)
	metadata["message_id"] = evt.Info.ID
	if evt.Info.PushName != "" {
		metadata["user_name"] = evt.Info.PushName
	}
	
	// BaseChannel.HandleMessage(senderID, chatID, content, media []string, metadata map[string]string)
	c.HandleMessage(senderID, chatID, content, nil, metadata)
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

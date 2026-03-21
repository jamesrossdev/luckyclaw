package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"
)

// PerformSetup initializes whatsmeow temporarily, handles QR Code pairing (if not already paired),
// and waits for the user to send the expectedCode in order to securely link their LID.
func PerformSetup(sessionPath string, expectedCode string) (string, error) {
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}

	dbPath := filepath.Join(sessionPath, whatsappDBName)
	storeURI := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", dbPath)

	db, err := sql.Open(sqliteDriver, storeURI)
	if err != nil {
		return "", fmt.Errorf("open sqlite db: %w", err)
	}
	defer db.Close()

	waLogger := waLog.Noop
	container := sqlstore.NewWithDB(db, sqliteDriver, waLogger)
	if err = container.Upgrade(context.Background()); err != nil {
		return "", fmt.Errorf("upgrade store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return "", fmt.Errorf("get device store: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, waLogger)

	var foundLID string
	var mu sync.Mutex
	done := make(chan struct{})

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			content := v.Message.GetConversation()
			if content == "" && v.Message.ExtendedTextMessage != nil {
				content = v.Message.ExtendedTextMessage.GetText()
			}
			if strings.TrimSpace(content) == expectedCode {
				mu.Lock()
				foundLID = v.Info.Sender.User
				mu.Unlock()
				select {
				case <-done:
				default:
					close(done)
				}
			}
		}
	})

	fmt.Println("  Initializing pairing session... please wait")

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return "", err
		}
		
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("\n  Scan this QR code with WhatsApp (Linked Devices)")
			} else {
				fmt.Println("  QR Status:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return "", err
		}
	}

	defer client.Disconnect()

	if expectedCode == "" {
		fmt.Printf("\n  ✓ WhatsApp Linked! (Open access mode)\n")
		return "", nil
	}

	fmt.Printf("\n  ✓ WhatsApp Linked!\n")
	fmt.Printf("  To authorize your number, please send this exact code from your phone to the bot: %s\n", expectedCode)
	fmt.Println("  Waiting for verification message... (Press Ctrl+C to cancel setup)")

	// Wait for the done signal or a timeout
	select {
	case <-done:
		return foundLID, nil
	case <-time.After(3 * time.Minute):
		return "", fmt.Errorf("timeout waiting for verification code")
	}
}

package whatsapp

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func openTestDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	uri := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open(sqliteDriver, uri)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := ensureInboundDedupTable(db); err != nil {
		t.Fatalf("ensure inbound dedup table: %v", err)
	}
	return db
}

func TestIsDuplicateStanzaMemoryAndKeyScope(t *testing.T) {
	db := openTestDB(t, filepath.Join(t.TempDir(), "store.db"))
	defer db.Close()

	c := &WhatsAppChannel{db: db}

	dup, source, _ := c.isDuplicateStanza("chat-1", "123@s.whatsapp.net", "stanza-1", false)
	if dup {
		t.Fatalf("first stanza should not be duplicate")
	}
	if source != "" {
		t.Fatalf("unexpected source on first insert: %s", source)
	}

	dup, source, _ = c.isDuplicateStanza("chat-1", "123@s.whatsapp.net", "stanza-1", false)
	if !dup {
		t.Fatalf("second identical stanza should be duplicate")
	}
	if source != "memory" {
		t.Fatalf("expected memory duplicate source, got %s", source)
	}

	dup, _, _ = c.isDuplicateStanza("chat-2", "123@s.whatsapp.net", "stanza-1", false)
	if dup {
		t.Fatalf("same stanza in different chat should not be duplicate")
	}
}

func TestIsDuplicateStanzaPersistsAcrossRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "store.db")

	db1 := openTestDB(t, dbPath)
	c1 := &WhatsAppChannel{db: db1}
	dup, _, _ := c1.isDuplicateStanza("chat-1", "123@s.whatsapp.net", "stanza-2", false)
	if dup {
		t.Fatalf("first stanza insert should not be duplicate")
	}
	_ = db1.Close()

	db2 := openTestDB(t, dbPath)
	defer db2.Close()
	c2 := &WhatsAppChannel{db: db2}

	dup, source, _ := c2.isDuplicateStanza("chat-1", "123@s.whatsapp.net", "stanza-2", false)
	if !dup {
		t.Fatalf("expected sqlite-backed duplicate after restart")
	}
	if source != "sqlite" {
		t.Fatalf("expected sqlite source, got %s", source)
	}
}

func TestIsDuplicateStanzaIgnoresExpiredDBEntry(t *testing.T) {
	db := openTestDB(t, filepath.Join(t.TempDir(), "store.db"))
	defer db.Close()

	key := makeInboundDedupKey("chat-3", "123@s.whatsapp.net", "stanza-3", false)
	expired := time.Now().Add(-inboundDedupDBTTL - time.Minute).Unix()
	if _, err := db.Exec(`
INSERT INTO inbound_dedup (dedup_key, stanza_id, chat_id, sender_jid, is_from_me, seen_at)
VALUES (?, ?, ?, ?, ?, ?)
`, key, "stanza-3", "chat-3", "123@s.whatsapp.net", 0, expired); err != nil {
		t.Fatalf("seed expired row: %v", err)
	}

	c := &WhatsAppChannel{db: db}
	dup, source, _ := c.isDuplicateStanza("chat-3", "123@s.whatsapp.net", "stanza-3", false)
	if dup {
		t.Fatalf("expired row should not be treated as duplicate")
	}
	if source != "" {
		t.Fatalf("expected empty source for non-duplicate, got %s", source)
	}
}

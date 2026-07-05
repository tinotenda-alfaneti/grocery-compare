package db

import (
	"path/filepath"
	"testing"
)

func TestOpen_CreatesSchemaAndSeedsStores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	conn, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer conn.Close()

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM stores`).Scan(&count); err != nil {
		t.Fatalf("querying stores: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 seeded stores, got %d", count)
	}

	var settingsCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM settings`).Scan(&settingsCount); err != nil {
		t.Fatalf("querying settings: %v", err)
	}
	if settingsCount != 1 {
		t.Fatalf("expected exactly 1 settings row, got %d", settingsCount)
	}
}

func TestOpen_IsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	conn1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open failed: %v", err)
	}
	if _, err := conn1.Exec(`INSERT INTO canonical_items (name) VALUES ('porridge oats')`); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	conn1.Close()

	// Reopening (simulating a pod restart) must not wipe existing data or
	// error out on the already-applied schema.
	conn2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	defer conn2.Close()

	var name string
	if err := conn2.QueryRow(`SELECT name FROM canonical_items WHERE name = 'porridge oats'`).Scan(&name); err != nil {
		t.Fatalf("expected previously inserted row to survive reopen: %v", err)
	}
}

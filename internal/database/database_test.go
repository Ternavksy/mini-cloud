package database

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenMigratesDatabaseAndConfiguresPool(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "metadata.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if db.Stats().MaxOpenConnections != 5 {
		t.Fatalf("MaxOpenConnections = %d, want 5", db.Stats().MaxOpenConnections)
	}

	var tableName string
	err = db.QueryRowContext(context.Background(), `SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'collections'`).Scan(&tableName)
	if err != nil {
		t.Fatalf("collections migration missing: %v", err)
	}
	if tableName != "collections" {
		t.Fatalf("table name = %q, want collections", tableName)
	}
}

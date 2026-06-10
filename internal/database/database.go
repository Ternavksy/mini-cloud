package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open metadata database: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping metadata database: %w", err)
	}

	if err := Migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS collections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS collection_files (
			collection_id TEXT NOT NULL,
			file_id TEXT NOT NULL,
			position INTEGER NOT NULL,
			PRIMARY KEY (collection_id, file_id),
			FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("apply database migration: %w", err)
		}
	}

	return nil
}

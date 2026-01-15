package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := ensureSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) UpsertLatest(ctx context.Context, record Record) (bool, error) {
	const stmt = `
		INSERT INTO scan_records (ip, port, service, timestamp, response)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(ip, port, service) DO UPDATE SET
			timestamp = excluded.timestamp,
			response = excluded.response
		WHERE excluded.timestamp > scan_records.timestamp
			OR (excluded.timestamp = scan_records.timestamp AND excluded.response != scan_records.response);
	`

	result, err := s.db.ExecContext(ctx, stmt, record.Ip, record.Port, record.Service, record.Timestamp, record.Response)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func ensureSchema(db *sql.DB) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS scan_records (
			ip TEXT NOT NULL,
			port INTEGER NOT NULL,
			service TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			response TEXT NOT NULL,
			PRIMARY KEY (ip, port, service)
		);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

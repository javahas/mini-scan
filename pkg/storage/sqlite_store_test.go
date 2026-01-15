package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSQLiteStoreUpsertLatest(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan_records.db")
	store, err := NewSQLiteStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	record := Record{Ip: "10.0.0.2", Port: 22, Service: "ssh", Timestamp: 100, Response: "a"}
	updated, err := store.UpsertLatest(context.Background(), record)
	if err != nil {
		t.Fatalf("UpsertLatest: %v", err)
	}
	if !updated {
		t.Fatalf("expected update for initial record")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.2", Port: 22, Service: "ssh", Timestamp: 50, Response: "b"})
	if err != nil {
		t.Fatalf("UpsertLatest older: %v", err)
	}
	if updated {
		t.Fatalf("expected no update for older record")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.2", Port: 22, Service: "ssh", Timestamp: 100, Response: "a"})
	if err != nil {
		t.Fatalf("UpsertLatest same: %v", err)
	}
	if updated {
		t.Fatalf("expected no update for identical record")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.2", Port: 22, Service: "ssh", Timestamp: 100, Response: "b"})
	if err != nil {
		t.Fatalf("UpsertLatest same timestamp: %v", err)
	}
	if !updated {
		t.Fatalf("expected update for same timestamp with different response")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.2", Port: 22, Service: "ssh", Timestamp: 200, Response: "c"})
	if err != nil {
		t.Fatalf("UpsertLatest newer: %v", err)
	}
	if !updated {
		t.Fatalf("expected update for newer record")
	}

	got := fetchSQLiteRecord(t, store, "10.0.0.2", 22, "ssh")
	if got.Timestamp != 200 || got.Response != "c" {
		t.Fatalf("unexpected record: %+v", got)
	}
}

func fetchSQLiteRecord(t *testing.T, store *SQLiteStore, ip string, port uint32, service string) Record {
	t.Helper()

	var ts int64
	var resp string
	row := store.db.QueryRow(`SELECT timestamp, response FROM scan_records WHERE ip = ? AND port = ? AND service = ?`, ip, port, service)
	if err := row.Scan(&ts, &resp); err != nil {
		t.Fatalf("scan record: %v", err)
	}

	return Record{
		Ip:        ip,
		Port:      port,
		Service:   service,
		Timestamp: ts,
		Response:  resp,
	}
}

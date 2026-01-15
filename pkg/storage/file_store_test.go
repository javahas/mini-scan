package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestFileStoreUpsertLatest(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "scan_records.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	key := fmt.Sprintf("%s|%d|%s", "10.0.0.1", 443, "https")
	record := Record{Ip: "10.0.0.1", Port: 443, Service: "https", Timestamp: 100, Response: "a"}

	updated, err := store.UpsertLatest(context.Background(), record)
	if err != nil {
		t.Fatalf("UpsertLatest: %v", err)
	}
	if !updated {
		t.Fatalf("expected update for initial record")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.1", Port: 443, Service: "https", Timestamp: 50, Response: "b"})
	if err != nil {
		t.Fatalf("UpsertLatest older: %v", err)
	}
	if updated {
		t.Fatalf("expected no update for older record")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.1", Port: 443, Service: "https", Timestamp: 100, Response: "a"})
	if err != nil {
		t.Fatalf("UpsertLatest same: %v", err)
	}
	if updated {
		t.Fatalf("expected no update for identical record")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.1", Port: 443, Service: "https", Timestamp: 100, Response: "b"})
	if err != nil {
		t.Fatalf("UpsertLatest same timestamp: %v", err)
	}
	if !updated {
		t.Fatalf("expected update for same timestamp with different response")
	}

	updated, err = store.UpsertLatest(context.Background(), Record{Ip: "10.0.0.1", Port: 443, Service: "https", Timestamp: 200, Response: "c"})
	if err != nil {
		t.Fatalf("UpsertLatest newer: %v", err)
	}
	if !updated {
		t.Fatalf("expected update for newer record")
	}

	reloaded, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore reload: %v", err)
	}

	got, ok := reloaded.data[key]
	if !ok {
		t.Fatalf("expected record for key %q", key)
	}
	if got.Timestamp != 200 || got.Response != "c" {
		t.Fatalf("unexpected record after reload: %+v", got)
	}
}

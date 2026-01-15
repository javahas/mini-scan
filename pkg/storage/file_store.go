package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileStore struct {
	path string
	mu   sync.Mutex
	data map[string]Record
}

func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{
		path: path,
		data: make(map[string]Record),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *FileStore) UpsertLatest(ctx context.Context, record Record) (bool, error) {
	_ = ctx

	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s|%d|%s", record.Ip, record.Port, record.Service)
	existing, ok := s.data[key]
	if ok {
		if record.Timestamp < existing.Timestamp {
			return false, nil
		}
		if record.Timestamp == existing.Timestamp && record.Response == existing.Response {
			return false, nil
		}
	}

	s.data[key] = record
	if err := s.persist(); err != nil {
		return false, err
	}

	return true, nil
}

func (s *FileStore) load() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&s.data); err != nil {
		return err
	}

	return nil
}

func (s *FileStore) persist() error {
	dir := filepath.Dir(s.path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(s.data); err != nil {
		return err
	}

	return file.Sync()
}

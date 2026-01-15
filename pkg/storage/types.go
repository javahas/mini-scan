package storage

import "context"

type Record struct {
	Ip        string `json:"ip"`
	Port      uint32 `json:"port"`
	Service   string `json:"service"`
	Timestamp int64  `json:"timestamp"`
	Response  string `json:"response"`
}

type Store interface {
	UpsertLatest(ctx context.Context, record Record) (bool, error)
}

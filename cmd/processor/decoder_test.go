package main

import (
	"encoding/json"
	"testing"

	"github.com/censys/scan-takehome/pkg/scanning"
)

func TestDecodeResponse(t *testing.T) {
	t.Parallel()

	v1Payload, err := json.Marshal(scanning.V1Data{ResponseBytesUtf8: []byte("hello")})
	if err != nil {
		t.Fatalf("marshal v1: %v", err)
	}
	got, err := decodeResponse(scanning.V1, v1Payload)
	if err != nil {
		t.Fatalf("decodeResponse v1: %v", err)
	}
	if got != "hello" {
		t.Fatalf("unexpected v1 response: %q", got)
	}

	v2Payload, err := json.Marshal(scanning.V2Data{ResponseStr: "world"})
	if err != nil {
		t.Fatalf("marshal v2: %v", err)
	}
	got, err = decodeResponse(scanning.V2, v2Payload)
	if err != nil {
		t.Fatalf("decodeResponse v2: %v", err)
	}
	if got != "world" {
		t.Fatalf("unexpected v2 response: %q", got)
	}
}

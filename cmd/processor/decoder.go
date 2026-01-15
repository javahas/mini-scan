package main

import (
	"encoding/json"
	"fmt"

	"github.com/censys/scan-takehome/pkg/scanning"
	"github.com/censys/scan-takehome/pkg/storage"
)

func decodeRecord(data []byte) (storage.Record, error) {
	var env scanning.Scan
	if err := json.Unmarshal(data, &env); err != nil {
		return storage.Record{}, err
	}

	payload, err := json.Marshal(env.Data)
	if err != nil {
		return storage.Record{}, err
	}

	response, err := decodeResponse(env.DataVersion, payload)
	if err != nil {
		return storage.Record{}, err
	}

	return storage.Record{
		Ip:        env.Ip,
		Port:      env.Port,
		Service:   env.Service,
		Timestamp: env.Timestamp,
		Response:  response,
	}, nil
}

func decodeResponse(version int, payload json.RawMessage) (string, error) {
	switch version {
	case scanning.V1:
		var data scanning.V1Data
		if err := json.Unmarshal(payload, &data); err != nil {
			return "", err
		}
		return string(data.ResponseBytesUtf8), nil
	case scanning.V2:
		var data scanning.V2Data
		if err := json.Unmarshal(payload, &data); err != nil {
			return "", err
		}
		return data.ResponseStr, nil
	default:
		return "", fmt.Errorf("unknown data version: %d", version)
	}
}

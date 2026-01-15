package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/censys/scan-takehome/pkg/storage"
)

const (
	StorageTypeFile    = "file"
	StorageTypeSqlLite = "sqlite"
)

func main() {
	projectId := flag.String("project", "test-project", "GCP Project ID")
	subscriptionID := flag.String("subscription", "scan-sub", "Pub/Sub subscription ID")
	storeType := flag.String("store-type", "file", "Storage backend: file or sqlite")
	storePath := flag.String("store-path", "data/file_recs.json", "Path for file-backed store")
	flag.Parse()
	store, err := initStore(*storeType, *storePath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, *projectId)
	if err != nil {
		log.Fatalf("failed to create pubsub client: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(*subscriptionID)
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.MaxOutstandingBytes = 10 << 20

	log.Printf("Data processor started: project=%s subscription=%s store=%s", *projectId, *subscriptionID, *storeType)

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		record, err := decodeRecord(msg.Data)
		if err != nil {
			log.Printf("failed to decode message: %v", err)
			msg.Nack()
			return
		}

		updated, err := store.UpsertLatest(ctx, record)
		if err != nil {
			log.Printf("failed to store scan record: %v", err)
			msg.Nack()
			return
		}

		if updated {
			log.Printf("updated %s:%d/%s at %d", record.Ip, record.Port, record.Service, record.Timestamp)
		}
		msg.Ack()
	})
	if err != nil {
		log.Fatalf("receive loop failed: %v", err)
	}

}

func initStore(storeType, storePath string) (storage.Store, error) {
	switch storeType {
	case StorageTypeFile:
		return storage.NewFileStore(storePath)
	case StorageTypeSqlLite:
		return storage.NewSQLiteStore(storePath)
	default:
		return nil, fmt.Errorf("unknown store type: %s", storeType)
	}
}

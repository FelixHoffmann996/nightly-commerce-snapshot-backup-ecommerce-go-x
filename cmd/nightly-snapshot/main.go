package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	snapshot "github.com/infrai-examples/nightly-commerce-snapshot"
)

func main() {
	once := flag.Bool("once", false, "run one snapshot and exit")
	inputPath := flag.String("input", "commerce.json", "commerce export JSON")
	bucket := flag.String("bucket", "commerce-nightly", "snapshot bucket")
	hour := flag.Int("hour", 2, "UTC hour for the nightly run")
	flag.Parse()

	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := snapshot.NewStorageClient(key)
	ctx := context.Background()
	if err := client.CreateBucket(ctx, *bucket); err != nil {
		log.Fatal(err)
	}
	run := func(day time.Time) {
		if err := archive(ctx, client, *bucket, *inputPath, day); err != nil {
			log.Printf("snapshot failed: %v", err)
		}
	}
	if *once {
		run(time.Now())
		return
	}
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), *hour, 0, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		time.Sleep(time.Until(next))
		run(next)
	}
}

func archive(ctx context.Context, client *snapshot.StorageClient, bucket, inputPath string, day time.Time) error {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read commerce export: %w", err)
	}
	var input snapshot.CommerceData
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("decode commerce export: %w", err)
	}
	record, err := snapshot.Build(day, input)
	if err != nil {
		return err
	}
	body, err := snapshot.Encode(record)
	if err != nil {
		return err
	}
	date := day.UTC().Format("2006-01-02")
	objectKey := "snapshots/" + date + ".json"
	signedURL, err := client.PresignPut(ctx, bucket, objectKey, "commerce-snapshot-"+date, len(body))
	if err != nil {
		return err
	}
	if err := client.Upload(ctx, signedURL, body); err != nil {
		return err
	}
	log.Printf("stored %s/%s with %d reconciled orders", bucket, objectKey, len(record.Orders))
	return nil
}

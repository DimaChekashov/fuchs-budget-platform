package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/db"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/outbox"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	_ = godotenv.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	pool, err := db.NewPostgresPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")
	topic := env("KAFKA_TOPIC", "wallet.operations.v1")
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.RequiredAcks(kgo.AllISRAcks()))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	relay := outbox.Relay{Pool: pool, Publish: func(ctx context.Context, walletID uuid.UUID, data []byte) error {
		return client.ProduceSync(ctx, &kgo.Record{Topic: topic, Key: []byte(walletID.String()), Value: data}).FirstErr()
	}}
	log.Printf("publishing wallet outbox to %s", topic)
	relay.Run(ctx)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

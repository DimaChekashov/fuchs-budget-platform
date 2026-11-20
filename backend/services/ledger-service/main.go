package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/ledger-service/internal/handler"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/ledger-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/ledger-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	reports := repository.Reports{Pool: pool}
	worker, cancelWorker := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.RunConsumer(worker, strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
			env("KAFKA_TOPIC", "wallet.operations.v1"), reports)
	}()
	defer func() { cancelWorker(); <-done }()

	api := handler.API{Reports: reports}
	server := http.Server{Addr: ":" + env("SERVER_PORT", "8082"), Handler: api.Routes(env("JWT_SECRET", "secret")),
		ReadHeaderTimeout: 5 * time.Second}
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdown)
		case <-stopped:
		}
	}()
	defer close(stopped)
	log.Printf("ledger-service listening on %s", server.Addr)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

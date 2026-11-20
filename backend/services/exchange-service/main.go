package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/exchange-service/internal/handler"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/exchange-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/exchange-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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

	options, err := redis.ParseURL(env("REDIS_URL", "redis://localhost:6379/0"))
	if err != nil {
		return err
	}
	options.DialTimeout, options.ReadTimeout, options.WriteTimeout = 500*time.Millisecond, 500*time.Millisecond, 500*time.Millisecond
	options.MaxRetries = -1
	cache := redis.NewClient(options)
	defer cache.Close()
	interval, err := time.ParseDuration(env("RATES_REFRESH_INTERVAL", "6h"))
	if err != nil || interval <= 0 {
		return errors.New("invalid RATES_REFRESH_INTERVAL")
	}
	catalog := repository.Catalog{Pool: pool}
	rates := service.Rates{Store: catalog, Cache: cache, Client: &http.Client{Timeout: 5 * time.Second},
		ProviderURL: env("RATES_PROVIDER_URL", "https://api.frankfurter.dev/v2")}
	worker, cancelWorker := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); rates.Run(worker, interval) }()
	defer func() { cancelWorker(); <-done }()

	api := handler.API{Catalog: catalog, Rates: rates}
	server := http.Server{Addr: ":" + env("SERVER_PORT", "8083"), Handler: api.Routes(env("JWT_SECRET", "secret")),
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
	log.Printf("exchange-service listening on %s", server.Addr)
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

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/config"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/db"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/router"
)

func main() {
	config := config.Load()

	ctx := context.Background()

	pool, err := db.NewPostgresPool(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	log.Println("connected to postgres!")

	router := router.New(pool, config)

	log.Printf("wallet-service running on :%s\n", config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, router))
}

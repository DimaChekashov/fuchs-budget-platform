package main

import (
	"context"
	"log"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/config"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/db"
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
	
}
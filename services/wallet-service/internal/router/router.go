package router

import (
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/config"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/handler"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/middleware"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	walletRepository := repository.NewWalletRepository(pool)
	transactionRepository := repository.NewTransactionRepository(pool)
	walletService := service.NewWalletService(walletRepository, transactionRepository)
	walletHandler := handler.NewWalletHandler(walletService)

	auth := middleware.Auth(cfg.JWTSecret)

	mux.Handle("GET /api/v1/wallets", auth(http.HandlerFunc(walletHandler.GetWallets)))
	mux.Handle("POST /api/v1/wallets", auth(http.HandlerFunc(walletHandler.CreateWallet)))
	mux.Handle("POST /api/v1/wallets/deposit", auth(http.HandlerFunc(walletHandler.Deposit)))
	mux.Handle("POST /api/v1/wallets/withdraw", auth(http.HandlerFunc(walletHandler.Withdraw)))

	return mux
}

package router

import (
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/pkg/jwtmiddleware"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/config"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/handler"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	walletRepository := repository.NewWalletRepository(pool)
	transactionRepository := repository.NewTransactionRepository(pool)
	walletService := service.NewWalletService(pool, walletRepository, transactionRepository, service.ExchangeCurrencyCheck(cfg.ExchangeURL))
	walletHandler := handler.NewWalletHandler(walletService)

	auth := jwtmiddleware.Auth(cfg.JWTSecret)

	mux.Handle("GET /api/v1/wallets", auth(http.HandlerFunc(walletHandler.GetWallets)))
	mux.Handle("POST /api/v1/wallets", auth(http.HandlerFunc(walletHandler.CreateWallet)))
	mux.Handle("POST /api/v1/wallets/deposit", auth(http.HandlerFunc(walletHandler.Deposit)))
	mux.Handle("POST /api/v1/wallets/withdraw", auth(http.HandlerFunc(walletHandler.Withdraw)))
	mux.Handle("GET /api/v1/wallets/{id}", auth(http.HandlerFunc(walletHandler.GetWallet)))
	mux.Handle("PATCH /api/v1/wallets/{id}", auth(http.HandlerFunc(walletHandler.RenameWallet)))
	mux.Handle("GET /api/v1/wallets/{id}/transactions", auth(http.HandlerFunc(walletHandler.GetHistory)))

	return mux
}

package router

import (
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/config"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/handler"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/middleware"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)

	health := handler.NewHealthHandler(pool)
	auth := handler.NewAuthHandler(userService, authService)
	user := handler.NewUserHandler(userService)

	authMiddleware := middleware.Auth(authService)

	mux.HandleFunc("GET /health", health.Health)
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc("POST /api/v1/auth/register", auth.Register)

	mux.Handle("GET /api/v1/users/me", authMiddleware(http.HandlerFunc(user.Me)))
	mux.Handle("GET /api/v1/users", authMiddleware(http.HandlerFunc(user.GetUsers)))

	return mux
}

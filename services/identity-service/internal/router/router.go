package router

import (
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/handler"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)

	health := handler.NewHealthHandler(pool)
	user := handler.NewUserHandler(userService)
	auth := handler.NewAuthHandler(userService)

	mux.HandleFunc("GET /health", health.Health)

	mux.HandleFunc("GET /api/v1/users", user.GetUsers)
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc("POST /api/v1/auth/register", auth.Register)

	return mux
}

package router

import (
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/handler"
)

func New() http.Handler {
	mux := http.NewServeMux()

	health := handler.NewHealthHandler()

	mux.HandleFunc("GET /health", health.Health)

	return mux
}

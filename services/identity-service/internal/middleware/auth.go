package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/service"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const EmailKey contextKey = "email"

func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := authService.ParseToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
			ctx = context.WithValue(ctx, EmailKey, claims["email"])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package handler

import (
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/pkg/jwtmiddleware"
	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "get users — coming soon",
	})
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(jwtmiddleware.UserIDKey).(string)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"erorr": "unauthorized"})
	}

	email, _ := r.Context().Value(jwtmiddleware.EmailKey).(string)

	writeJSON(w, http.StatusOK, map[string]any{
		"id":    userID,
		"email": email,
	})
}

package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/middleware"
	"github.com/DimaChekashov/fuchs-budget-platform/services/wallet-service/internal/service"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	wallets, err := h.walletService.GetUserWallets(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, wallets)
}

type createWalletRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req createWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Currency == "" {
		req.Currency = "RUB"
	}

	wallet, err := h.walletService.CreateWallet(r.Context(), userID, req.Name, req.Currency)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, wallet)
}

type depositRequest struct {
	WalletID    string  `json:"wallet_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid wallet_id"})
		return
	}

	txn, err := h.walletService.Deposit(r.Context(), service.DepositInput{
		WalletID:    walletID,
		UserID:      userID,
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWalletNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "wallet not found"})
		case errors.Is(err, service.ErrWrongOwner):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid wallet_id"})
		return
	}

	txn, err := h.walletService.Withdraw(r.Context(), service.WithdrawInput{
		WalletID:    walletID,
		UserID:      userID,
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientFunds):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "insufficient funds"})
		case errors.Is(err, service.ErrWalletNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "wallet not found"})
		case errors.Is(err, service.ErrWrongOwner):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

func getUserID(r *http.Request) uuid.UUID {
	id, _ := r.Context().Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(id)
	return userID
}

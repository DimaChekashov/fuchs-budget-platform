package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/pkg/jwtmiddleware"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/service"
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

	wallet, err := h.walletService.CreateWallet(r.Context(), userID, req.Name, req.Currency, r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnknownCurrency):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrCurrencyUnavailable):
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "currency catalog unavailable"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusCreated, wallet)
}

func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	h.record(w, r, h.walletService.Deposit)
}

func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	h.record(w, r, h.walletService.Withdraw)
}

type operationRequest struct {
	WalletID    string          `json:"wallet_id"`
	Amount      json.RawMessage `json:"amount"`
	Description string          `json:"description"`
}

func (h *WalletHandler) record(w http.ResponseWriter, r *http.Request, run func(context.Context, service.OperationInput) (*domain.Transaction, error)) {
	var req operationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid wallet_id"})
		return
	}
	var key *string
	if values := r.Header.Values("Idempotency-Key"); len(values) > 0 {
		if len(values) != 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": service.ErrInvalidKey.Error()})
			return
		}
		key = &values[0]
	}
	txn, err := run(r.Context(), service.OperationInput{
		WalletID: walletID, UserID: getUserID(r), Amount: req.Amount,
		Description: req.Description, IdempotencyKey: key,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount), errors.Is(err, service.ErrInvalidKey),
			errors.Is(err, service.ErrInsufficientFunds), errors.Is(err, service.ErrBalanceLimit), errors.Is(err, service.ErrCurrencyPrecision):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrWalletNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "wallet not found"})
		case errors.Is(err, service.ErrWrongOwner):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		case errors.Is(err, service.ErrIdempotencyConflict):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

func getUserID(r *http.Request) uuid.UUID {
	id := jwtmiddleware.GetUserID(r)
	userID, _ := uuid.Parse(id)
	return userID
}

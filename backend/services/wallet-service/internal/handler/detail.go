package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/wallet-service/internal/service"
	"github.com/google/uuid"
)

func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	id, ok := walletPathID(w, r)
	if !ok {
		return
	}
	wallet, err := h.walletService.GetWallet(r.Context(), id, getUserID(r))
	writeWalletResult(w, wallet, err)
}

func (h *WalletHandler) RenameWallet(w http.ResponseWriter, r *http.Request) {
	id, ok := walletPathID(w, r)
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	wallet, err := h.walletService.RenameWallet(r.Context(), id, getUserID(r), req.Name)
	writeWalletResult(w, wallet, err)
}

func (h *WalletHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := walletPathID(w, r)
	if !ok {
		return
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid offset"})
			return
		}
		offset = value
	}
	history, err := h.walletService.GetHistory(r.Context(), id, getUserID(r), offset)
	writeWalletResult(w, history, err)
}

func walletPathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid wallet_id"})
		return uuid.Nil, false
	}
	return id, true
}

func writeWalletResult(w http.ResponseWriter, result any, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, result)
	case errors.Is(err, service.ErrWalletNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "wallet not found"})
	case errors.Is(err, service.ErrInvalidName):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

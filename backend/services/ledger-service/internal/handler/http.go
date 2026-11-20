package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/pkg/jwtmiddleware"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/ledger-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/ledger-service/internal/repository"
	"github.com/google/uuid"
)

type API struct{ Reports repository.Reports }

func (a API) Routes(secret string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/reports/summary", jwtmiddleware.Auth(secret)(http.HandlerFunc(a.summary)))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Reports.Pool.Ping(r.Context()); err != nil {
			respond(w, http.StatusServiceUnavailable, map[string]string{"status": "postgres unavailable"})
			return
		}
		respond(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func (a API) summary(w http.ResponseWriter, r *http.Request) {
	viewer, err := uuid.Parse(jwtmiddleware.GetUserID(r))
	if err != nil || viewer == uuid.Nil {
		failure(w, http.StatusUnauthorized, "invalid user")
		return
	}
	filters := r.URL.Query()
	to := time.Now().UTC().Truncate(24 * time.Hour)
	if raw := filters.Get("to"); raw != "" {
		if to, err = time.Parse(time.DateOnly, raw); err != nil {
			failure(w, http.StatusBadRequest, "invalid end date")
			return
		}
	}
	from := to.AddDate(0, 0, -29)
	if raw := filters.Get("from"); raw != "" {
		if from, err = time.Parse(time.DateOnly, raw); err != nil {
			failure(w, http.StatusBadRequest, "invalid start date")
			return
		}
	}
	if from.After(to) || to.After(from.AddDate(0, 0, 365)) {
		failure(w, http.StatusBadRequest, "the date range must contain at most 366 days")
		return
	}
	var wallet *uuid.UUID
	if raw := filters.Get("wallet_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil || id == uuid.Nil {
			failure(w, http.StatusBadRequest, "invalid wallet_id")
			return
		}
		wallet = &id
	}
	currency := filters.Get("currency")
	if utf8.RuneCountInString(currency) > 10 {
		failure(w, http.StatusBadRequest, "invalid currency")
		return
	}
	report, err := a.Reports.Summary(r.Context(), domain.Query{UserID: viewer, WalletID: wallet,
		Currency: currency, From: from, To: to})
	if err != nil {
		failure(w, http.StatusInternalServerError, "could not read report")
		return
	}
	respond(w, http.StatusOK, report)
}

func failure(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

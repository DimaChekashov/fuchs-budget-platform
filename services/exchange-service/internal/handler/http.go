package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/DimaChekashov/fuchs-budget-platform/pkg/jwtmiddleware"
	"github.com/DimaChekashov/fuchs-budget-platform/services/exchange-service/internal/domain"
	"github.com/DimaChekashov/fuchs-budget-platform/services/exchange-service/internal/repository"
	"github.com/DimaChekashov/fuchs-budget-platform/services/exchange-service/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type API struct {
	Catalog repository.Catalog
	Rates   service.Rates
}

var codePattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{2,9}$`)

func (a API) Routes(secret string) http.Handler {
	mux := http.NewServeMux()
	auth := jwtmiddleware.Auth(secret)
	mux.Handle("GET /api/v1/currencies", auth(http.HandlerFunc(a.currencies)))
	mux.Handle("GET /api/v1/currencies/{code}", auth(http.HandlerFunc(a.currency)))
	mux.Handle("POST /api/v1/currencies", auth(http.HandlerFunc(a.addCurrency)))
	mux.Handle("GET /api/v1/rates", auth(http.HandlerFunc(a.latest)))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Catalog.Pool.Ping(r.Context()); err != nil {
			problem(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		send(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func userID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(jwtmiddleware.GetUserID(r))
	if err != nil || id == uuid.Nil {
		problem(w, http.StatusUnauthorized, "invalid user")
		return uuid.Nil, false
	}
	return id, true
}

func (a API) currencies(w http.ResponseWriter, r *http.Request) {
	user, ok := userID(w, r)
	if !ok {
		return
	}
	rows, err := a.Catalog.List(r.Context(), user)
	if err != nil {
		problem(w, http.StatusInternalServerError, "could not read currencies")
		return
	}
	send(w, http.StatusOK, rows)
}

func (a API) currency(w http.ResponseWriter, r *http.Request) {
	user, ok := userID(w, r)
	if !ok {
		return
	}
	code := strings.ToUpper(r.PathValue("code"))
	if !codePattern.MatchString(code) {
		problem(w, http.StatusBadRequest, "invalid currency code")
		return
	}
	currency, err := a.Catalog.Get(r.Context(), user, code)
	if errors.Is(err, pgx.ErrNoRows) {
		problem(w, http.StatusNotFound, "currency not found")
		return
	}
	if err != nil {
		problem(w, http.StatusInternalServerError, "could not read currency")
		return
	}
	send(w, http.StatusOK, currency)
}

func (a API) addCurrency(w http.ResponseWriter, r *http.Request) {
	user, ok := userID(w, r)
	if !ok {
		return
	}
	var input struct {
		Code      string `json:"code"`
		Precision *int   `json:"precision"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil {
		problem(w, http.StatusBadRequest, "invalid body")
		return
	}
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if !codePattern.MatchString(code) || input.Precision == nil || *input.Precision < 0 || *input.Precision > 2 {
		problem(w, http.StatusBadRequest, "invalid currency")
		return
	}
	currency := domain.Currency{Code: code, Precision: *input.Precision}
	added, err := a.Catalog.Add(r.Context(), user, currency)
	if err != nil {
		problem(w, http.StatusInternalServerError, "could not add currency")
		return
	}
	if !added {
		problem(w, http.StatusConflict, "currency already exists")
		return
	}
	send(w, http.StatusCreated, currency)
}

func (a API) latest(w http.ResponseWriter, r *http.Request) {
	if _, ok := userID(w, r); !ok {
		return
	}
	base := strings.ToUpper(r.URL.Query().Get("base"))
	if base == "" {
		base = "RUB"
	}
	if !codePattern.MatchString(base) {
		problem(w, http.StatusBadRequest, "invalid base currency")
		return
	}
	rates, err := a.Rates.Latest(r.Context(), base)
	if err != nil {
		problem(w, http.StatusInternalServerError, "could not read exchange rates")
		return
	}
	if rates == nil {
		rates = []domain.Rate{}
	}
	send(w, http.StatusOK, rates)
}

func problem(w http.ResponseWriter, status int, message string) {
	send(w, status, map[string]string{"error": message})
}

func send(w http.ResponseWriter, status int, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(result)
}

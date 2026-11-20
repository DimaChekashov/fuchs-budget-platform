package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/exchange-service/internal/domain"
)

type memoryRates struct{ rates []domain.Rate }

func (s *memoryRates) Bases(context.Context) ([]string, error) { return []string{"RUB", "USD"}, nil }
func (s *memoryRates) SaveRates(_ context.Context, rates []domain.Rate) error {
	s.rates = append(s.rates, rates...)
	return nil
}
func (s *memoryRates) LatestRates(_ context.Context, _ string) ([]domain.Rate, error) {
	return s.rates, nil
}

func TestRatesKeepLastObservationWhenProviderFails(t *testing.T) {
	var fail atomic.Bool
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() || r.URL.Query().Get("base") == "USD" {
			w.WriteHeader(503)
			return
		}
		if r.URL.Path != "/v2/rates" || r.URL.Query().Get("quotes") != "USD" {
			t.Errorf("unexpected URL %s", r.URL)
			return
		}
		fmt.Fprint(w, `[{"base":"RUB","quote":"USD","date":"2026-08-20","rate":0.0123456789}]`)
	}))
	defer provider.Close()

	store := &memoryRates{}
	rates := Rates{Store: store, Client: provider.Client(), ProviderURL: provider.URL + "/v2"}
	if err := rates.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	fail.Store(true)
	if err := rates.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	saved, err := rates.Latest(t.Context(), "RUB")
	if err != nil || len(saved) != 1 || saved[0].Value != "0.0123456789" ||
		saved[0].Source != "Frankfurter" || saved[0].ObservedOn != "2026-08-20" {
		t.Fatalf("unexpected stored rates: %+v, %v", saved, err)
	}
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/DimaChekashov/fuchs-budget-platform/services/exchange-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RateStore interface {
	Bases(context.Context) ([]string, error)
	SaveRates(context.Context, []domain.Rate) error
	LatestRates(context.Context, string) ([]domain.Rate, error)
}

type Rates struct {
	Store       RateStore
	Cache       *redis.Client
	Client      *http.Client
	ProviderURL string
}

const cachePrefix = "fuchs:exchange:rates:v1:"

func (r Rates) Latest(ctx context.Context, base string) ([]domain.Rate, error) {
	if r.Cache != nil {
		if data, err := r.Cache.Get(ctx, cachePrefix+base).Bytes(); err == nil {
			var cached []domain.Rate
			if json.Unmarshal(data, &cached) == nil {
				return cached, nil
			}
		}
	}
	rates, err := r.Store.LatestRates(ctx, base)
	if err == nil && r.Cache != nil {
		if data, marshalErr := json.Marshal(rates); marshalErr == nil {
			_ = r.Cache.Set(ctx, cachePrefix+base, data, time.Minute).Err()
		}
	}
	return rates, err
}

func (r Rates) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for ctx.Err() == nil {
		if err := r.Refresh(ctx); err != nil && ctx.Err() == nil {
			log.Printf("refresh exchange rates: %v", err)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (r Rates) Refresh(ctx context.Context) error {
	bases, err := r.Store.Bases(ctx)
	if err != nil {
		return err
	}
	for _, base := range bases {
		if err := r.refreshBase(ctx, base, bases); err != nil {
			// A missing currency or one failed pair must not prevent the other bases from updating.
			log.Printf("refresh exchange rates for %s: %v", base, err)
		}
	}
	return ctx.Err()
}

func (r Rates) refreshBase(ctx context.Context, base string, bases []string) error {
	endpoint, err := url.Parse(strings.TrimRight(r.ProviderURL, "/") + "/rates")
	if err != nil {
		return err
	}
	query := endpoint.Query()
	query.Set("base", base)
	quotes := make([]string, 0, len(bases)-1)
	for _, code := range bases {
		if code != base {
			quotes = append(quotes, code)
		}
	}
	query.Set("quotes", strings.Join(quotes, ","))
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}
	res, err := r.Client.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("provider status %d", res.StatusCode)
	}
	var data []struct {
		Date  string      `json:"date"`
		Base  string      `json:"base"`
		Quote string      `json:"quote"`
		Rate  json.Number `json:"rate"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&data); err != nil {
		return err
	}
	rates := make([]domain.Rate, 0, len(data))
	for _, row := range data {
		if row.Base != base || !slices.Contains(quotes, row.Quote) {
			continue
		}
		day, err := time.Parse(time.DateOnly, row.Date)
		rate, numberErr := row.Rate.Float64()
		if err == nil && numberErr == nil && rate > 0 {
			rates = append(rates, domain.Rate{Base: base, Quote: row.Quote, ObservedOn: day.Format(time.DateOnly),
				Value: string(row.Rate), Source: "Frankfurter"})
		}
	}
	if err := r.Store.SaveRates(ctx, rates); err != nil {
		return err
	}
	if r.Cache != nil {
		_ = r.Cache.Del(ctx, cachePrefix+base).Err()
	}
	return nil
}

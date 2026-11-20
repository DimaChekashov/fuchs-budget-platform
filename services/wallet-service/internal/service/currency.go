package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var (
	ErrUnknownCurrency     = errors.New("currency is not in your catalog")
	ErrCurrencyUnavailable = errors.New("currency catalog unavailable")
	currencyCode           = regexp.MustCompile(`^[A-Z][A-Z0-9]{2,9}$`)
)

type CurrencyCheck func(ctx context.Context, code, authorization string) (int, error)

func ExchangeCurrencyCheck(base string) CurrencyCheck {
	if base == "" {
		return nil
	}
	client := &http.Client{
		Timeout:       3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	return func(ctx context.Context, code, authorization string) (int, error) {
		if !currencyCode.MatchString(code) {
			return 0, ErrUnknownCurrency
		}
		endpoint, err := url.JoinPath(base, "api/v1/currencies", code)
		if err != nil {
			return 0, ErrCurrencyUnavailable
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return 0, ErrCurrencyUnavailable
		}
		req.Header.Set("Authorization", authorization)
		response, err := client.Do(req)
		if err != nil {
			return 0, ErrCurrencyUnavailable
		}
		defer response.Body.Close()
		if response.StatusCode == http.StatusNotFound {
			return 0, ErrUnknownCurrency
		}
		if response.StatusCode != http.StatusOK {
			return 0, ErrCurrencyUnavailable
		}
		var result struct {
			Code      string `json:"code"`
			Precision *int   `json:"precision"`
		}
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil || result.Code != code ||
			result.Precision == nil || *result.Precision < 0 || *result.Precision > 2 {
			return 0, ErrCurrencyUnavailable
		}
		return *result.Precision, nil
	}
}

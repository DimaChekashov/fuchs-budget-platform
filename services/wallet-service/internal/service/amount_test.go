package service

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseAmount(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"1e2", "100.00"}, {"1.230", "1.23"}, {"0.01", "0.01"},
		{"9999999999999999.99", "9999999999999999.99"},
		{"10000000000000000", ""}, {"0.001", ""}, {"-1", ""}, {"0", ""},
		{"", ""}, {"null", ""}, {`"1.00"`, ""}, {"1e-999999999", ""},
	} {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseAmount(json.RawMessage(tc.in))
			if tc.want == "" {
				if !errors.Is(err, ErrInvalidAmount) {
					t.Fatalf("got %q, %v; want invalid amount", got, err)
				}
				return
			}
			if err != nil || string(got) != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

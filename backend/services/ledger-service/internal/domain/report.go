package domain

import (
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	Version    int       `json:"version"`
	ID         uuid.UUID `json:"operation_id"`
	UserID     uuid.UUID `json:"user_id"`
	WalletID   uuid.UUID `json:"wallet_id"`
	Type       string    `json:"type"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	RecordedAt time.Time `json:"recorded_at"`
}

type Query struct {
	UserID   uuid.UUID
	WalletID *uuid.UUID
	Currency string
	From     time.Time
	To       time.Time
}

type Total struct {
	Currency string `json:"currency"`
	Income   string `json:"income"`
	Expenses string `json:"expenses"`
}

type Report struct {
	From      string     `json:"from"`
	To        string     `json:"to"`
	UpdatedAt *time.Time `json:"updated_at"`
	Totals    []Total    `json:"totals"`
}

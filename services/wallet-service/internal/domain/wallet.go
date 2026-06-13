package domain

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Currency  string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

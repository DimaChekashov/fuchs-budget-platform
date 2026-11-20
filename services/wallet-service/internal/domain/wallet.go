package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Currency  string
	Balance   json.Number
	CreatedAt time.Time
	UpdatedAt time.Time
}

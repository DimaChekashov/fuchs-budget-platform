package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionDeposit    TransactionType = "deposit"
	TransactionWithdrawal TransactionType = "withdrawal"
	TransactionTransfer   TransactionType = "transfer"
)

type Transaction struct {
	ID          uuid.UUID
	WalletID    uuid.UUID
	Type        TransactionType
	Amount      json.Number
	Description string
	CreatedAt   time.Time
}

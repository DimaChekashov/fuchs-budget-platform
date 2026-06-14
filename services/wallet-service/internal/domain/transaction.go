package domain

import (
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
	Amount      float64
	Description string
	CreatedAt   time.Time
}

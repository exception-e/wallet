package domain

import (
	"time"

	"github.com/google/uuid"
)

// Wallet holds money in a single currency. Balance is measured in minor units.
type Wallet struct {
	ID        uuid.UUID
	Balance   int64
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

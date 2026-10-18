package types

import (
	"walletApp/internal/domain"

	"github.com/google/uuid"
)

type ApplyOperationInput struct {
	WalletID       uuid.UUID
	IdempotencyKey string
	Type           domain.OperationType
	Amount         int64
}

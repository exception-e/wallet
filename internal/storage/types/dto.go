package types

import (
	"context"
	"walletApp/internal/domain"

	"github.com/google/uuid"
)

type UpdateInput struct {
	Id      uuid.UUID
	Balance *int64
}

type TxFunc = func(ctx context.Context) error

type CreateInput struct {
	WalletID       uuid.UUID
	IdempotencyKey string
	Type           domain.OperationType
	Amount         int64
	Status         domain.OperationStatus
	BalanceAfter   *int64
	RejectionCode  *domain.RejectionCode
}

type GetByIdempotencyKeyInput struct {
	WalletID       uuid.UUID
	IdempotencyKey string
}

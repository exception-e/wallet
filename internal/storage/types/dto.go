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

type CreateWalletInput struct {
	Currency string
	Balance  int64
}

type CreateOperationInput struct {
	WalletID       uuid.UUID
	IdempotencyKey string
	Type           domain.OperationType
	Amount         int64
	Status         domain.OperationStatus
	RejectionCode  *domain.RejectionCode
}

type GetOperationByIdempotencyKeyInput struct {
	WalletID       uuid.UUID
	IdempotencyKey string
}

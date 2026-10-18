package types

import (
	"context"
	"walletApp/internal/domain"

	"github.com/google/uuid"
)

type Storage interface {
	RunInTx(ctx context.Context, fn TxFunc) error

	CreateWallet(ctx context.Context, input CreateWalletInput) (*domain.Wallet, error)
	GetWallet(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	UpdateWallet(ctx context.Context, input UpdateInput) (*domain.Wallet, error)
	LockWallet(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)

	CreateOperation(ctx context.Context, input CreateOperationInput) (*domain.Operation, error)
	GetOperationByIdempotencyKey(ctx context.Context, input GetOperationByIdempotencyKeyInput) (*domain.Operation, error)
}

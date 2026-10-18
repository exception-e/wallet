package types

import (
	"context"
	"walletApp/internal/domain"

	"github.com/google/uuid"
)

type Storage interface {
	Get(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	Update(ctx context.Context, input UpdateInput) (*domain.Wallet, error)
	Lock(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	RunInTx(ctx context.Context, fn TxFunc) error
	Create(ctx context.Context, input CreateInput) (*domain.Operation, error)
	GetByIdempotencyKey(ctx context.Context, input GetByIdempotencyKeyInput) (*domain.Operation, error)
}

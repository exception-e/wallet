package types

import (
	"context"
	"walletApp/internal/domain"

	"github.com/google/uuid"
)

type Service interface {
	Get(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	ApplyOperation(ctx context.Context, input ApplyOperationInput) (*domain.Operation, error)
}

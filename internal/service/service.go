package service

import (
	"context"
	"errors"
	"fmt"
	"walletApp/internal/domain"
	types "walletApp/internal/service/types"
	storage "walletApp/internal/storage/types"

	"github.com/google/uuid"
)

var _ types.Service = (*WalletService)(nil)

type WalletService struct {
	walletStorage storage.Storage
}

func New(storage storage.Storage) types.Service {
	return &WalletService{
		walletStorage: storage,
	}
}

func (s *WalletService) Get(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("walletService.Get: %w", types.ErrInvalidWalletID)
	}

	wallet, err := s.walletStorage.Get(ctx, id)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("walletService.Get (walletID=%s): %w", id, types.ErrWalletNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletService.Get (walletID=%s): %w", id, err)
	}

	return wallet, nil
}

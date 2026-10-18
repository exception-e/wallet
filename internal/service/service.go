package service

import (
	"context"
	"errors"
	"fmt"
	"walletApp/internal/domain"
	types "walletApp/internal/service/types"
	storagetypes "walletApp/internal/storage/types"

	"github.com/google/uuid"
)

var _ types.Service = (*WalletService)(nil)

type WalletService struct {
	walletStorage storagetypes.Storage
}

func New(storage storagetypes.Storage) types.Service {
	return &WalletService{
		walletStorage: storage,
	}
}

func (s *WalletService) Create(ctx context.Context, input types.CreateInput) (*domain.Wallet, error) {
	if len(input.Currency) != 3 {
		return nil, fmt.Errorf("walletService.Create (currency=%q): %w", input.Currency, types.ErrInvalidCurrency)
	}
	for _, letter := range input.Currency {
		if letter < 'A' || letter > 'Z' {
			return nil, fmt.Errorf("walletService.Create (currency=%q): %w", input.Currency, types.ErrInvalidCurrency)
		}
	}

	wallet, err := s.walletStorage.CreateWallet(ctx, storagetypes.CreateWalletInput{
		Currency: input.Currency,
		Balance:  0,
	})
	if err != nil {
		return nil, fmt.Errorf("walletService.Create (create wallet, currency=%s): %w", input.Currency, err)
	}

	return wallet, nil
}

func (s *WalletService) Get(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("walletService.Get: %w", types.ErrInvalidWalletID)
	}

	wallet, err := s.walletStorage.GetWallet(ctx, id)
	if errors.Is(err, storagetypes.ErrNotFound) {
		return nil, fmt.Errorf("walletService.Get (walletID=%s): %w", id, types.ErrWalletNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletService.Get (walletID=%s): %w", id, err)
	}

	return wallet, nil
}

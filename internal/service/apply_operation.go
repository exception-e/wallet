package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
	"walletApp/internal/domain"
	"walletApp/internal/service/types"

	"github.com/google/uuid"

	storagetypes "walletApp/internal/storage/types"
)

const maxIdempotencyKeyBytes = 255

func (s *WalletService) ApplyOperation(ctx context.Context, input types.ApplyOperationInput) (*domain.Operation, error) {
	if err := validateOperation(input); err != nil {
		return nil, fmt.Errorf("walletService.ApplyOperation (validate, walletID=%s): %w", input.WalletID, err)
	}

	var operation *domain.Operation
	err := s.walletStorage.RunInTx(ctx, func(txCtx context.Context) error {
		var txErr error
		operation, txErr = s.applyOperationInTx(txCtx, input)
		return txErr
	})
	if err != nil {
		// A result read or written inside the transaction is not returned if commit failed.
		return nil, fmt.Errorf("walletService.ApplyOperation (run transaction, walletID=%s): %w", input.WalletID, err)
	}

	return operation, nil
}

func (s *WalletService) applyOperationInTx(ctx context.Context, input types.ApplyOperationInput) (*domain.Operation, error) {
	wallet, err := s.walletStorage.Lock(ctx, input.WalletID)
	if errors.Is(err, storagetypes.ErrNotFound) {
		return nil, fmt.Errorf("walletService.applyOperationInTx (lock walletID=%s): %w", input.WalletID, types.ErrWalletNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletService.applyOperationInTx (lock walletID=%s): %w", input.WalletID, err)
	}

	createInput := storagetypes.CreateInput{
		WalletID:       input.WalletID,
		IdempotencyKey: input.IdempotencyKey,
		Type:           input.Type,
		Amount:         input.Amount,
		Status:         domain.OperationStatusApplied,
	}

	balance := wallet.Balance
	switch input.Type {
	case domain.OperationTypeDeposit:
		if input.Amount > math.MaxInt64-balance {
			code := domain.RejectionCodeBalanceOverflow
			createInput.RejectionCode = &code
		} else {
			balance += input.Amount
		}
	case domain.OperationTypeWithdraw:
		if input.Amount > balance {
			code := domain.RejectionCodeInsufficientFunds
			createInput.RejectionCode = &code
		} else {
			balance -= input.Amount
		}
	default:
		return nil, fmt.Errorf("walletService.applyOperationInTx (walletID=%s): %w", input.WalletID, types.ErrInvalidOperationType)
	}

	if createInput.RejectionCode != nil {
		createInput.Status = domain.OperationStatusRejected
	} else {
		createInput.BalanceAfter = &balance
	}

	operation, err := s.walletStorage.Create(ctx, createInput)
	if errors.Is(err, storagetypes.ErrAlreadyExists) {
		operation, err = s.walletStorage.GetByIdempotencyKey(ctx, storagetypes.GetByIdempotencyKeyInput{
			WalletID:       input.WalletID,
			IdempotencyKey: input.IdempotencyKey,
		})
		if err != nil {
			return nil, fmt.Errorf("walletService.applyOperationInTx (get operation, walletID=%s): %w", input.WalletID, err)
		}
		if operation.Type != input.Type || operation.Balance != input.Amount {
			return nil, fmt.Errorf("walletService.applyOperationInTx (check idempotency, walletID=%s): %w", input.WalletID, types.ErrIdempotencyConflict)
		}
		return operation, nil
	}
	if errors.Is(err, storagetypes.ErrNotFound) {
		return nil, fmt.Errorf("walletService.applyOperationInTx (create operation, walletID=%s): %w", input.WalletID, types.ErrWalletNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("walletService.applyOperationInTx (create operation, walletID=%s): %w", input.WalletID, err)
	}

	if operation.Status == domain.OperationStatusApplied {
		_, err = s.walletStorage.Update(ctx, storagetypes.UpdateInput{
			Id:      input.WalletID,
			Balance: &balance,
		})
		if errors.Is(err, storagetypes.ErrNotFound) {
			return nil, fmt.Errorf("walletService.applyOperationInTx (update walletID=%s): %w", input.WalletID, types.ErrWalletNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("walletService.applyOperationInTx (update walletID=%s): %w", input.WalletID, err)
		}
	}

	return operation, nil
}

func validateOperation(input types.ApplyOperationInput) error {
	if input.WalletID == uuid.Nil {
		return fmt.Errorf("wallet.validateOperation: %w", types.ErrInvalidWalletID)
	}
	if len(input.IdempotencyKey) > maxIdempotencyKeyBytes ||
		strings.TrimSpace(input.IdempotencyKey) == "" ||
		!utf8.ValidString(input.IdempotencyKey) ||
		strings.ContainsRune(input.IdempotencyKey, '\x00') {
		return fmt.Errorf("wallet.validateOperation: %w", types.ErrInvalidIdempotencyKey)
	}
	if input.Amount <= 0 {
		return fmt.Errorf("wallet.validateOperation: %w", types.ErrInvalidAmount)
	}
	if input.Type != domain.OperationTypeDeposit && input.Type != domain.OperationTypeWithdraw {
		return fmt.Errorf("wallet.validateOperation: %w", types.ErrInvalidOperationType)
	}

	return nil
}

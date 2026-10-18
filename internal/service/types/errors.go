package types

import "errors"

var (
	ErrWalletNotFound        = errors.New("wallet not found")
	ErrInvalidWalletID       = errors.New("invalid wallet ID")
	ErrInvalidAmount         = errors.New("amount must be positive")
	ErrInvalidOperationType  = errors.New("invalid operation type")
	ErrInvalidIdempotencyKey = errors.New("invalid idempotency key")
	ErrIdempotencyConflict   = errors.New("idempotency key was used with different parameters")
)

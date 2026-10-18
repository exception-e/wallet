package types

import "errors"

// Repository methods wrap these errors with the method and wallet ID.
var (
	ErrNotFound            = errors.New("wallet not found")
	ErrTransactionRequired = errors.New("transaction required")
	ErrAlreadyExists       = errors.New("operation already exists")
)

package types

import "errors"

// Handler methods wrap these errors with call context; internal causes are not exposed in HTTP responses.
var (
	ErrInvalidRequest        = errors.New("invalid request")
	ErrInvalidWalletID       = errors.New("invalid wallet UUID")
	ErrMissingIdempotencyKey = errors.New("missing Idempotency-Key header")
	ErrUnsupportedMediaType  = errors.New("content type must be application/json")
)

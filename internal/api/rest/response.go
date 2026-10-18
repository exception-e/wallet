package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"walletApp/internal/api/rest/types"

	servicetypes "walletApp/internal/service/types"
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusBadRequest
	var response types.ErrorResponse
	var bodyTooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &bodyTooLarge):
		status = http.StatusRequestEntityTooLarge
		response = types.ErrorResponse{Code: "REQUEST_TOO_LARGE", Message: "request body exceeds 1 MiB"}
	case errors.Is(err, types.ErrUnsupportedMediaType):
		status = http.StatusUnsupportedMediaType
		response = types.ErrorResponse{Code: "UNSUPPORTED_MEDIA_TYPE", Message: types.ErrUnsupportedMediaType.Error()}
	case errors.Is(err, types.ErrInvalidRequest):
		response = types.ErrorResponse{Code: "INVALID_REQUEST", Message: types.ErrInvalidRequest.Error()}
	case errors.Is(err, types.ErrInvalidWalletID), errors.Is(err, servicetypes.ErrInvalidWalletID):
		response = types.ErrorResponse{Code: "INVALID_WALLET_ID", Message: types.ErrInvalidWalletID.Error()}
	case errors.Is(err, types.ErrMissingIdempotencyKey):
		response = types.ErrorResponse{Code: "MISSING_IDEMPOTENCY_KEY", Message: types.ErrMissingIdempotencyKey.Error()}
	//case errors.Is(err, servicetypes.ErrInvalidCurrency):
	//	response = types.ErrorResponse{Code: "INVALID_CURRENCY", Message: servicetypes.ErrInvalidCurrency.Error()}
	case errors.Is(err, servicetypes.ErrInvalidAmount):
		response = types.ErrorResponse{Code: "INVALID_AMOUNT", Message: servicetypes.ErrInvalidAmount.Error()}
	case errors.Is(err, servicetypes.ErrInvalidOperationType):
		response = types.ErrorResponse{Code: "INVALID_OPERATION_TYPE", Message: servicetypes.ErrInvalidOperationType.Error()}
	case errors.Is(err, servicetypes.ErrInvalidIdempotencyKey):
		response = types.ErrorResponse{Code: "INVALID_IDEMPOTENCY_KEY", Message: servicetypes.ErrInvalidIdempotencyKey.Error()}
	case errors.Is(err, servicetypes.ErrWalletNotFound):
		status = http.StatusNotFound
		response = types.ErrorResponse{Code: "WALLET_NOT_FOUND", Message: servicetypes.ErrWalletNotFound.Error()}
	case errors.Is(err, servicetypes.ErrIdempotencyConflict):
		status = http.StatusConflict
		response = types.ErrorResponse{Code: "IDEMPOTENCY_CONFLICT", Message: servicetypes.ErrIdempotencyConflict.Error()}
	default:
		status = http.StatusInternalServerError
		response = types.ErrorResponse{Code: "INTERNAL_ERROR", Message: "internal server error"}
		slog.ErrorContext(r.Context(), "HTTP request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	}
	writeJSON(w, r, status, response)
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, response any) {
	body, err := json.Marshal(response)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to encode HTTP response", "error", fmt.Errorf("rest.writeJSON (encode): %w", err))
		status = http.StatusInternalServerError
		body = []byte(`{"code":"INTERNAL_ERROR","message":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(body, '\n')); err != nil {
		slog.ErrorContext(r.Context(), "Failed to write HTTP response", "error", fmt.Errorf("rest.writeJSON (write): %w", err))
	}
}

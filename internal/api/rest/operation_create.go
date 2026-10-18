package rest

import (
	"fmt"
	"net/http"
	"walletApp/internal/api/rest/types"
	"walletApp/internal/domain"

	servicetypes "walletApp/internal/service/types"
)

func (h *handler) CreateOperation(w http.ResponseWriter, r *http.Request) {
	keys := r.Header.Values("Idempotency-Key")
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "") {
		h.writeError(w, r, fmt.Errorf("restHandler.CreateOperation: %w", types.ErrMissingIdempotencyKey))
		return
	}
	if len(keys) != 1 {
		h.writeError(w, r, fmt.Errorf("restHandler.CreateOperation (multiple idempotency keys): %w", types.ErrInvalidRequest))
		return
	}

	input, err := decodeJSON[types.CreateOperationRequest](w, r)
	if err != nil {
		h.writeError(w, r, fmt.Errorf("restHandler.CreateOperation (decode request): %w", err))
		return
	}

	operation, err := h.service.ApplyOperation(r.Context(), servicetypes.ApplyOperationInput{
		WalletID:       input.WalletID,
		IdempotencyKey: keys[0],
		Type:           domain.OperationType(input.OperationType),
		Amount:         input.Amount,
	})
	if err != nil {
		h.writeError(w, r, fmt.Errorf("restHandler.CreateOperation (apply operation): %w", err))
		return
	}

	status := http.StatusOK
	if operation.Status == domain.OperationStatusRejected {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, r, status, operationResponse(operation))
}

func operationResponse(operation *domain.Operation) types.OperationResponse {
	response := types.OperationResponse{
		OperationID: operation.ID.String(),
		WalletID:    operation.WalletID.String(),
		Status:      string(operation.Status),
	}
	if operation.RejectionCode != nil {
		code := string(*operation.RejectionCode)
		response.RejectionCode = &code
	}
	return response
}

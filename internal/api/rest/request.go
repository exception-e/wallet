package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"walletApp/internal/api/rest/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxRequestBodyBytes = 1 << 20

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("rest.decodeJSON (content type): %w: %w", types.ErrUnsupportedMediaType, err)
	}
	if mediaType != "application/json" {
		return nil, fmt.Errorf("rest.decodeJSON: %w", types.ErrUnsupportedMediaType)
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input *T
	if err := decoder.Decode(&input); err != nil {
		return nil, fmt.Errorf("rest.decodeJSON (decode body): %w: %w", types.ErrInvalidRequest, err)
	}
	if input == nil {
		return nil, fmt.Errorf("rest.decodeJSON (null body): %w", types.ErrInvalidRequest)
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, fmt.Errorf("rest.decodeJSON (trailing data): %w: %w", types.ErrInvalidRequest, err)
		}
		return nil, fmt.Errorf("rest.decodeJSON (multiple JSON values): %w", types.ErrInvalidRequest)
	}

	return input, nil
}

func walletIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, fmt.Errorf("rest.walletIDFromRequest: %w: %w", types.ErrInvalidWalletID, err)
	}
	if id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("rest.walletIDFromRequest: %w", types.ErrInvalidWalletID)
	}
	return id, nil
}

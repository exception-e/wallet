package rest

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	types "walletApp/internal/api/rest/types"
	"walletApp/internal/domain"
	servicetypes "walletApp/internal/service/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var _ types.Handler = (*handler)(nil)

type handler struct {
	service servicetypes.Service
	logger  *slog.Logger
}

func (h *handler) SetupRouter(r chi.Router) chi.Router {
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/wallets", h.CreateWallet)
		r.Post("/wallet", h.CreateOperation)
		r.Get("/wallets/{id}", h.Balance)
	})
	return r
}

func (h *handler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	input, err := decodeJSON[types.CreateWalletRequest](w, r)
	if err != nil {
		h.writeError(w, r, fmt.Errorf("restHandler.CreateWallet (decode request): %w", err))
		return
	}

	wallet, err := h.service.Create(r.Context(), servicetypes.CreateInput{Currency: input.Currency})
	if err != nil {
		h.writeError(w, r, fmt.Errorf("restHandler.CreateWallet (create wallet): %w", err))
		return
	}

	w.Header().Set("Location", "/api/v1/wallets/"+wallet.ID.String())
	writeJSON(w, r, http.StatusCreated, walletResponse(wallet))
}

func (h *handler) Balance(w http.ResponseWriter, r *http.Request) {
	walletID, err := walletIDFromRequest(r)
	if err != nil {
		h.writeError(w, r, fmt.Errorf("restHandler.GetWallet (parse wallet ID): %w", err))
		return
	}

	wallet, err := h.service.Get(r.Context(), walletID)
	if err != nil {
		h.writeError(w, r, fmt.Errorf("restHandler.GetWallet (walletID=%s): %w", walletID, err))
		return
	}

	writeJSON(w, r, http.StatusOK, walletResponse(wallet))
}

func New(logger *slog.Logger, service servicetypes.Service) types.Handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

func walletResponse(wallet *domain.Wallet) types.WalletResponse {
	return types.WalletResponse{
		WalletID: wallet.ID.String(),
		Balance:  wallet.Balance,
		Currency: wallet.Currency,
	}
}

package types

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	SetupRouter(router chi.Router) chi.Router
	Balance(w http.ResponseWriter, r *http.Request)
	CreateOperation(w http.ResponseWriter, r *http.Request)
}

package url

import (
	"net/http"

	"github.com/labib0x9/short/internal/transport/http/middleware"
)

func (h *Handler) RegisterRoutes(mux *http.ServeMux, manager *middleware.Manager) {

	mux.Handle(
		"GET /{code}/stat",
		manager.With(
			http.HandlerFunc(h.Analysis),
		),
	)

	mux.Handle(
		"GET /{code}",
		manager.With(
			http.HandlerFunc(h.Get),
		),
	)

	mux.Handle(
		"POST /short",
		manager.With(
			http.HandlerFunc(h.Shorten),
		),
	)
}

package url

import "github.com/labib0x9/short/internal/app/url"

type Handler struct {
	srv *url.Service
}

func NewHandler(srv *url.Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

package url

import (
	"github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/transport/http/middleware"
)

type Handler struct {
	srv        *url.Service
	middlwares *middleware.Middlewares
}

func NewHandler(srv *url.Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

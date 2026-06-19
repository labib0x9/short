package url

import (
	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/internal/app/url"
)

type Handler struct {
	srv      url.Service
	validate *validator.Validate
}

func NewHandler(srv url.Service, validate *validator.Validate) *Handler {
	return &Handler{
		srv:      srv,
		validate: validate,
	}
}

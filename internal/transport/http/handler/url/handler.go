package url

import (
	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/domain/queue"
)

type Handler struct {
	srv      url.Service
	queue    queue.Queue
	validate *validator.Validate
}

func NewHandler(srv url.Service, queue queue.Queue, validate *validator.Validate) *Handler {
	return &Handler{
		srv:      srv,
		queue:    queue,
		validate: validate,
	}
}

package url

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	urldomain "github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/utils"
)

type urlRequest struct {
	Url      string     `json:"url" validate:"required,url"`
	ExpireAt *time.Time `json:"expire_at"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req urlRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		utils.SendError(w, "bad request", http.StatusBadRequest)
		slog.Warn("Shorten: bad json body", "error", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.SendError(w, "bad request", http.StatusBadRequest)
		slog.Warn("Shorten: struct validation failed", "error", err)
		return
	}

	result, err := h.srv.Shorten(r.Context(), req.Url, req.ExpireAt, r.Header.Get("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, urldomain.ErrShortCodeExpired):
			utils.SendError(w, "url expired", http.StatusGone)
		case errors.Is(err, urldomain.ErrShortCodeCollision):
			fallthrough
		default:
			utils.SendError(w, "internl server error", http.StatusInternalServerError)
		}
		slog.Warn("Shorten: srv.Shorten()", "error", err)
		return
	}

	utils.SendJson(w, result, http.StatusCreated)
}

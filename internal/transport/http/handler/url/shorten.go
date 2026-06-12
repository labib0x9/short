package url

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

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
		http.Error(w, "Bad request", http.StatusBadRequest)
		slog.Warn("Shorten: bad json body", "error", err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		slog.Warn("Shorten: struct validation failed", "error", err)
		return
	}

	result, err := h.srv.Shorten(req.Url, req.ExpireAt, r.Header.Get("User-Agent"))
	if err != nil {
		http.Error(w, "internl server error", http.StatusInternalServerError)
		slog.Warn("Shorten: srv.Shorten()", "error", err)
		return
	}

	utils.SendJson(w, result, http.StatusCreated)
}

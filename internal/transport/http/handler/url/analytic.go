package url

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/utils"
)

func (h *Handler) Analysis(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		utils.SendError(w, "bad request", http.StatusBadRequest)
		slog.Warn("Analytics: empty code")
		return
	}

	stat, err := h.srv.Analysis(r.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, url.ErrShortCodeExpired):
			utils.SendError(w, "url expired", http.StatusGone)
		case errors.Is(err, sql.ErrNoRows):
			utils.SendError(w, "not found", http.StatusNotFound)
		default:
			utils.SendError(w, "internl server error", http.StatusInternalServerError)
		}
		slog.Warn("Analytic: srv.Analysis()", "error", err, "ERR", errors.Is(err, sql.ErrNoRows))
		return
	}

	utils.SendJson(w, stat, http.StatusOK)
}

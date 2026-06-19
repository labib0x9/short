package url

import (
	"errors"
	"log/slog"
	"net/http"

	urldomain "github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/utils"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		utils.SendError(w, "bad request", http.StatusBadRequest)
		slog.Warn("Get: empty code")
		return
	}

	url, err := h.srv.Get(r.Context(), code, r.Referer(), r.UserAgent(), r.RemoteAddr)
	if err != nil {
		switch {
		case errors.Is(err, urldomain.ErrShortCodeExpired):
			utils.SendError(w, "url expired", http.StatusGone)
		default:
			utils.SendError(w, "internl server error", http.StatusInternalServerError)
		}
		slog.Warn("Analytic: srv.Get()", "error", err)
		return
	}

	http.Redirect(w, r, url.URL, http.StatusFound)
}

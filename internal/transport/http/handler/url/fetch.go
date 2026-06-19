package url

import (
	"log/slog"
	"net/http"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		// 400
		return
	}

	url, err := h.srv.Get(r.Context(), code, r.Referer(), r.UserAgent(), r.RemoteAddr)
	if err != nil {
		http.Error(w, "internl server error", http.StatusInternalServerError)
		slog.Warn("Get: srv.Get()", "error", err)
		return
	}

	http.Redirect(w, r, url.URL, http.StatusFound)
}

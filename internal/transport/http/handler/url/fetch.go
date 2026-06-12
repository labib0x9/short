package url

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labib0x9/short/internal/infra/queue"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		// 400
		return
	}

	url, err := h.srv.Get(r.Context(), code)
	if err != nil {
		http.Error(w, "internl server error", http.StatusInternalServerError)
		slog.Warn("Get: srv.Get()", "error", err)
		return
	}

	http.Redirect(w, r, url.URL, http.StatusFound)

	h.queue.Publish(r.Context(), queue.ClickEvent{
		ShortCode: code,
		ClickedAt: time.Now(),
		Referer:   r.Referer(),
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
		Retries:   2,
	})
}

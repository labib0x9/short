package url

import (
	"log/slog"
	"net/http"

	"github.com/labib0x9/short/internal/utils"
)

func (h *Handler) Analysis(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		// 400
		return
	}

	stat, err := h.srv.Analysis(code)
	if err != nil {
		http.Error(w, "internl server error", http.StatusInternalServerError)
		slog.Warn("Get: srv.Get()", "error", err)
		return
	}

	utils.SendJson(w, stat, http.StatusOK)
}

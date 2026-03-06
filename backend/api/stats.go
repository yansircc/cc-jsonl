package api

import (
	"net/http"

	"cc-jsonl/model"
)

func (h *handler) handleStats(w http.ResponseWriter, r *http.Request) {
	count, totalSize := h.store.Stats()
	writeJSON(w, http.StatusOK, model.StatsResponse{
		Sessions:  count,
		TotalSize: totalSize,
	})
}

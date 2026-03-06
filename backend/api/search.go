package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cc-jsonl/model"
)

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数 q 不能为空"})
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	project := r.URL.Query().Get("project")

	result := h.engine.Search(model.SearchRequest{
		Query:   q,
		Project: project,
		Offset:  offset,
		Limit:   limit,
	})

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

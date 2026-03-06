package api

import (
	"net/http"
	"strconv"
)

func (h *handler) handleContext(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionId")
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数 uuid 不能为空"})
		return
	}

	around := 1
	if v := r.URL.Query().Get("around"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			around = n
		}
	}

	ctx, err := h.engine.ReadContext(sessionID, uuid, around)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if ctx == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "消息不存在"})
		return
	}

	writeJSON(w, http.StatusOK, ctx)
}

package api

import (
	"io/fs"
	"net/http"
	"strings"

	"cc-jsonl/config"
	"cc-jsonl/index"
	"cc-jsonl/search"
)

// NewRouter 创建 HTTP 路由
func NewRouter(cfg config.Config, store *index.Store, engine *search.Engine, staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	h := &handler{cfg: cfg, store: store, engine: engine}

	mux.HandleFunc("GET /api/search", h.handleSearch)
	mux.HandleFunc("GET /api/context/{sessionId}", h.handleContext)
	mux.HandleFunc("GET /api/stats", h.handleStats)

	// 静态文件服务（SPA fallback）
	fileServer := http.FileServer(http.FS(staticFS))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// 尝试提供静态文件
		path := r.URL.Path
		if path == "/" {
			fileServer.ServeHTTP(w, r)
			return
		}

		// 检查文件是否存在
		f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// SPA fallback: 返回 index.html
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})

	return withMiddleware(mux)
}

type handler struct {
	cfg    config.Config
	store  *index.Store
	engine *search.Engine
}

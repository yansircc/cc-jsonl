package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"cc-jsonl/api"
	"cc-jsonl/config"
	"cc-jsonl/index"
	"cc-jsonl/search"
)

//go:embed all:ui/dist
var uiFS embed.FS

// startServer 启动 HTTP 服务
func startServer(cfg config.Config, store *index.Store) error {
	engine := search.NewEngine(store, cfg)

	// 提取嵌入的静态文件
	distFS, err := fs.Sub(uiFS, "ui/dist")
	if err != nil {
		return fmt.Errorf("嵌入文件系统错误: %w", err)
	}

	router := api.NewRouter(cfg, store, engine, distFS)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
}

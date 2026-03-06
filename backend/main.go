package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"cc-jsonl/config"
	"cc-jsonl/index"
	"cc-jsonl/model"
	"cc-jsonl/search"
)

func main() {
	cfg := config.Default()

	port := flag.Int("port", cfg.Port, "HTTP 监听端口")
	dataDir := flag.String("data", cfg.DataDir, "Claude 数据目录")
	query := flag.String("query", "", "CLI 搜索模式：直接搜索并打印结果")
	project := flag.String("project", "", "按项目过滤")
	flag.Parse()

	cfg.Port = *port
	cfg.DataDir = *dataDir

	log.Printf("数据目录: %s", cfg.DataDir)
	log.Printf("构建索引中...")

	store := index.Build(cfg)

	if *query != "" {
		// CLI 搜索模式
		engine := search.NewEngine(store, cfg)
		result := engine.Search(model.SearchRequest{
			Query:   *query,
			Project: *project,
			Limit:   20,
		})
		fmt.Printf("找到 %d 条结果（耗时 %dms）\n\n", result.Total, result.QueryTimeMs)
		for i, hit := range result.Results {
			fmt.Printf("[%d] %s | %s | %s\n", i+1, hit.Role, hit.Timestamp.Format("2006-01-02 15:04"), hit.Project)
			fmt.Printf("    %s\n\n", hit.Snippet)
		}
		os.Exit(0)
	}

	// 启动文件监视
	watcher := index.NewWatcher(store, cfg)
	watcher.Start()
	defer watcher.Stop()

	log.Printf("启动 HTTP 服务: http://localhost:%d", cfg.Port)
	if err := startServer(cfg, store); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

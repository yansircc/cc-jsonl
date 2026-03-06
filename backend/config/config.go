package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config 应用配置
type Config struct {
	// HTTP 监听端口
	Port int
	// Claude 数据目录（~/.claude）
	DataDir string
	// 搜索并发数
	Workers int
	// 单行最大读取字节数（防 OOM）
	MaxLineBytes int
	// 搜索结果默认 limit
	DefaultLimit int
	// 文件变更检测周期（秒）
	WatchInterval int
}

// Default 返回默认配置
func Default() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Port:          3456,
		DataDir:       filepath.Join(home, ".claude"),
		Workers:       runtime.NumCPU(),
		MaxLineBytes:  10 * 1024 * 1024, // 10MB
		DefaultLimit:  20,
		WatchInterval: 30,
	}
}

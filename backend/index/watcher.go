package index

import (
	"log"
	"os"
	"time"

	"cc-jsonl/config"
)

// Watcher 轮询检测文件变更，更新索引
type Watcher struct {
	store    *Store
	cfg      config.Config
	fileMod  map[string]time.Time // filePath -> modTime
	stopCh   chan struct{}
}

// NewWatcher 创建文件变更监视器
func NewWatcher(store *Store, cfg config.Config) *Watcher {
	w := &Watcher{
		store:   store,
		cfg:     cfg,
		fileMod: make(map[string]time.Time),
		stopCh:  make(chan struct{}),
	}
	// 初始化已知文件的修改时间
	for _, info := range store.AllFilePaths("") {
		if stat, err := os.Stat(info.FilePath); err == nil {
			w.fileMod[info.FilePath] = stat.ModTime()
		}
	}
	return w
}

// Start 启动轮询
func (w *Watcher) Start() {
	go func() {
		ticker := time.NewTicker(time.Duration(w.cfg.WatchInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.check()
			case <-w.stopCh:
				return
			}
		}
	}()
	log.Printf("文件监视已启动（%ds 周期）", w.cfg.WatchInterval)
}

// Stop 停止轮询
func (w *Watcher) Stop() {
	close(w.stopCh)
}

func (w *Watcher) check() {
	sessions := w.store.AllFilePaths("")
	updated := 0
	for _, info := range sessions {
		stat, err := os.Stat(info.FilePath)
		if err != nil {
			continue
		}
		lastMod, known := w.fileMod[info.FilePath]
		if !known || stat.ModTime().After(lastMod) {
			// 文件已变更，重新提取元数据
			newInfo := extractSessionInfo(info.FilePath)
			if newInfo != nil {
				w.store.Update(*newInfo)
				updated++
			}
			w.fileMod[info.FilePath] = stat.ModTime()
		}
	}
	if updated > 0 {
		log.Printf("索引更新: %d 个会话", updated)
	}
}

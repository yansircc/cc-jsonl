package index

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cc-jsonl/config"
	"cc-jsonl/model"
)

// jsonlRecord 用于快速提取元数据的最小结构
type jsonlRecord struct {
	Type      string    `json:"type"`
	SessionID string    `json:"sessionId"`
	Cwd       string    `json:"cwd"`
	Timestamp string    `json:"timestamp"`
}

// Build 遍历所有项目目录，构建会话索引
func Build(cfg config.Config) *Store {
	store := NewStore()
	projectsDir := filepath.Join(cfg.DataDir, "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		log.Printf("读取项目目录失败: %v", err)
		return store
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.Workers)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(projectsDir, entry.Name())
		// 扫描目录下的 JSONL 文件
		jsonlFiles, _ := filepath.Glob(filepath.Join(dirPath, "*.jsonl"))
		// 也扫描子 agent 目录
		subFiles, _ := filepath.Glob(filepath.Join(dirPath, "*/subagents/*.jsonl"))
		jsonlFiles = append(jsonlFiles, subFiles...)

		for _, f := range jsonlFiles {
			wg.Add(1)
			sem <- struct{}{}
			go func(filePath string) {
				defer wg.Done()
				defer func() { <-sem }()
				info := extractSessionInfo(filePath)
				if info != nil {
					store.Add(*info)
				}
			}(f)
		}
	}

	wg.Wait()
	log.Printf("索引构建完成: %d 个会话", store.Count())
	return store
}

// extractSessionInfo 从 JSONL 文件中提取会话元数据
// 只读取前几行和最后几行，避免全文扫描
func extractSessionInfo(filePath string) *model.SessionInfo {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil
	}

	// 从文件名提取 session ID
	base := filepath.Base(filePath)
	sessionID := strings.TrimSuffix(base, ".jsonl")

	info := &model.SessionInfo{
		ID:       sessionID,
		FilePath: filePath,
		FileSize: stat.Size(),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	msgCount := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var rec jsonlRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			continue
		}

		// 提取项目路径
		if info.Project == "" && rec.Cwd != "" {
			info.Project = rec.Cwd
		}

		// 只统计 user/assistant 消息
		if rec.Type == "user" || rec.Type == "assistant" {
			msgCount++
			ts := parseTimestamp(rec.Timestamp)
			if !ts.IsZero() {
				if info.StartTime.IsZero() || ts.Before(info.StartTime) {
					info.StartTime = ts
				}
				if ts.After(info.LastTime) {
					info.LastTime = ts
				}
			}
		}
	}

	info.MessageCount = msgCount
	if msgCount == 0 {
		return nil // 跳过空会话
	}

	return info
}

// parseTimestamp 解析 ISO 8601 时间戳
func parseTimestamp(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}

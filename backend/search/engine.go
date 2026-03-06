package search

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"cc-jsonl/config"
	"cc-jsonl/index"
	"cc-jsonl/model"
)

// Engine 搜索引擎，编排并发文件扫描
type Engine struct {
	store *index.Store
	cfg   config.Config
}

// NewEngine 创建搜索引擎
func NewEngine(store *index.Store, cfg config.Config) *Engine {
	return &Engine{store: store, cfg: cfg}
}

// Search 执行搜索，返回分页结果
func (e *Engine) Search(req model.SearchRequest) model.SearchResult {
	start := time.Now()

	if req.Limit <= 0 {
		req.Limit = e.cfg.DefaultLimit
	}

	queryLower := []byte(strings.ToLower(req.Query))
	sessions := e.store.AllFilePaths(req.Project)

	// 并发扫描所有文件
	var mu sync.Mutex
	var allHits []model.SearchHit
	var wg sync.WaitGroup
	sem := make(chan struct{}, e.cfg.Workers)

	for _, session := range sessions {
		wg.Add(1)
		sem <- struct{}{}
		go func(s model.SessionInfo) {
			defer wg.Done()
			defer func() { <-sem }()
			hits := ScanFile(s.FilePath, queryLower, e.cfg.MaxLineBytes)
			if len(hits) > 0 {
				mu.Lock()
				allHits = append(allHits, hits...)
				mu.Unlock()
			}
		}(session)
	}
	wg.Wait()

	// 按时间倒序排列
	sort.Slice(allHits, func(i, j int) bool {
		return allHits[i].Timestamp.After(allHits[j].Timestamp)
	})

	total := len(allHits)
	hasMore := false
	var page []model.SearchHit

	if req.Offset < total {
		end := req.Offset + req.Limit
		if end > total {
			end = total
		}
		page = allHits[req.Offset:end]
		hasMore = end < total
	}

	return model.SearchResult{
		Results:     page,
		Total:       total,
		HasMore:     hasMore,
		QueryTimeMs: time.Since(start).Milliseconds(),
	}
}

// ReadContext 读取指定消息的上下文（前后 around 条消息）
func (e *Engine) ReadContext(sessionID, uuid string, around int) (*model.ContextResponse, error) {
	info, ok := e.store.Get(sessionID)
	if !ok {
		return nil, nil
	}

	messages := readSessionMessages(info.FilePath, e.cfg.MaxLineBytes)

	// 找到目标消息
	hitIndex := -1
	for i, m := range messages {
		if m.UUID == uuid {
			hitIndex = i
			break
		}
	}
	if hitIndex < 0 {
		return nil, nil
	}

	// 计算上下文窗口
	start := hitIndex - around
	if start < 0 {
		start = 0
	}
	end := hitIndex + around + 1
	if end > len(messages) {
		end = len(messages)
	}

	return &model.ContextResponse{
		Messages: messages[start:end],
		HitIndex: hitIndex - start,
	}, nil
}

// readSessionMessages 读取文件中所有 user/assistant 消息
func readSessionMessages(filePath string, maxLineBytes int) []model.Message {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	reader := bufio.NewReaderSize(f, 64*1024)
	var messages []model.Message

	for {
		line, err := readLine(reader, maxLineBytes)
		if len(line) == 0 && err != nil {
			break
		}

		// 字节级类型过滤
		prefix := line
		if len(prefix) > 1024 {
			prefix = prefix[:1024]
		}
		isUser := bytes.Contains(prefix, typeUser) || bytes.Contains(prefix, typeUserSp)
		isAssistant := bytes.Contains(prefix, typeAssistant) || bytes.Contains(prefix, typeAssistantSp)
		if !isUser && !isAssistant {
			continue
		}

		var msg jsonlMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg.Type != "user" && msg.Type != "assistant" {
			continue
		}

		text := extractText(&msg)
		if text == "" {
			continue
		}

		ts := parseTimestamp(msg.Timestamp)

		messages = append(messages, model.Message{
			UUID:      msg.UUID,
			Role:      msg.Type,
			Text:      text,
			Timestamp: ts,
		})
	}

	return messages
}

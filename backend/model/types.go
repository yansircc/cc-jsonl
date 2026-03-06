package model

//go:generate go run github.com/gzuidhof/tygo@latest generate

import "time"

// SessionInfo 会话元数据（索引中存储）
type SessionInfo struct {
	// 会话 ID（文件名中的 UUID）
	ID string `json:"id"`
	// 项目真实路径（从 cwd 字段获取）
	Project string `json:"project"`
	// JSONL 文件路径
	FilePath string `json:"-"`
	// 会话开始时间
	StartTime time.Time `json:"startTime"`
	// 会话最后活动时间
	LastTime time.Time `json:"lastTime"`
	// 消息数量（user + assistant）
	MessageCount int `json:"messageCount"`
	// 文件大小（字节）
	FileSize int64 `json:"fileSize"`
}

// SearchRequest 搜索请求参数
type SearchRequest struct {
	Query   string `json:"q"`
	Project string `json:"project,omitempty"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Results    []SearchHit `json:"results"`
	Total      int         `json:"total"`
	HasMore    bool        `json:"hasMore"`
	QueryTimeMs int64      `json:"queryTimeMs"`
}

// SearchHit 单条搜索命中
type SearchHit struct {
	SessionID   string      `json:"sessionId"`
	MessageUUID string      `json:"messageUuid"`
	Project     string      `json:"project"`
	Role        string      `json:"role"`
	Timestamp   time.Time   `json:"timestamp"`
	Snippet     string      `json:"snippet"`
	Highlights  []Highlight `json:"highlights"`
}

// ContextResponse 上下文消息响应
type ContextResponse struct {
	Messages []Message `json:"messages"`
	HitIndex int       `json:"hitIndex"`
}

// StatsResponse 统计信息响应
type StatsResponse struct {
	Sessions  int   `json:"sessions"`
	TotalSize int64 `json:"totalSize"`
}

// Highlight 高亮位置
type Highlight struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Message 单条消息
type Message struct {
	UUID      string    `json:"uuid"`
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

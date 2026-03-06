package search

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"time"

	"cc-jsonl/model"
)

// 类型过滤标记（字节级快速检查）
var (
	typeUser      = []byte(`"type":"user"`)
	typeAssistant = []byte(`"type":"assistant"`)
	// 兼容带空格的 JSON 格式
	typeUserSp      = []byte(`"type": "user"`)
	typeAssistantSp = []byte(`"type": "assistant"`)
)

// jsonlMessage 用于解析搜索匹配行的结构
type jsonlMessage struct {
	Type      string          `json:"type"`
	UUID      string          `json:"uuid"`
	SessionID string          `json:"sessionId"`
	Cwd       string          `json:"cwd"`
	Timestamp string          `json:"timestamp"`
	Message   json.RawMessage `json:"message"`
}

// messageContent 解析 message 字段
type messageContent struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// contentBlock assistant 消息的 content 数组元素
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ScanFile 扫描单个 JSONL 文件，返回匹配结果
// 三层快速路径：类型过滤 → 关键词预检 → JSON 解析确认
func ScanFile(filePath string, queryLower []byte, maxLineBytes int) []model.SearchHit {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	reader := bufio.NewReaderSize(f, 64*1024)
	var hits []model.SearchHit

	for {
		line, err := readLine(reader, maxLineBytes)
		if len(line) == 0 && err != nil {
			break
		}

		// 第一层：字节级类型过滤（检查前 1KB）
		prefix := line
		if len(prefix) > 1024 {
			prefix = prefix[:1024]
		}
		isUser := bytes.Contains(prefix, typeUser) || bytes.Contains(prefix, typeUserSp)
		isAssistant := bytes.Contains(prefix, typeAssistant) || bytes.Contains(prefix, typeAssistantSp)
		if !isUser && !isAssistant {
			continue
		}

		// 第二层：字节级关键词预检
		if !bytes.Contains(bytes.ToLower(line), queryLower) {
			continue
		}

		// 第三层：JSON 解析确认
		hit := parseLine(line, queryLower)
		if hit != nil {
			hits = append(hits, *hit)
		}
	}

	return hits
}

// readLine 读取一行，超过 maxBytes 则截断
func readLine(reader *bufio.Reader, maxBytes int) ([]byte, error) {
	var line []byte
	for {
		part, isPrefix, err := reader.ReadLine()
		line = append(line, part...)
		if !isPrefix || err != nil {
			return line, err
		}
		if len(line) > maxBytes {
			// 截断超长行，消费剩余部分
			for isPrefix && err == nil {
				_, isPrefix, err = reader.ReadLine()
			}
			return line[:maxBytes], err
		}
	}
}

// parseLine 解析 JSON 行，提取文本并确认匹配
func parseLine(line []byte, queryLower []byte) *model.SearchHit {
	var msg jsonlMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil
	}

	if msg.Type != "user" && msg.Type != "assistant" {
		return nil
	}

	text := extractText(&msg)
	if text == "" {
		return nil
	}

	// 确认文本中包含关键词
	textLower := strings.ToLower(text)
	queryStr := string(queryLower)
	idx := strings.Index(textLower, queryStr)
	if idx < 0 {
		return nil
	}

	// 提取片段和高亮
	snippet, highlights := ExtractSnippet(text, queryStr, 150)

	ts := parseTimestamp(msg.Timestamp)

	return &model.SearchHit{
		SessionID:   msg.SessionID,
		MessageUUID: msg.UUID,
		Project:     msg.Cwd,
		Role:        msg.Type,
		Timestamp:   ts,
		Snippet:     snippet,
		Highlights:  highlights,
	}
}

// extractText 从消息中提取纯文本
func extractText(msg *jsonlMessage) string {
	if len(msg.Message) == 0 {
		return ""
	}

	var content messageContent
	if err := json.Unmarshal(msg.Message, &content); err != nil {
		return ""
	}

	// user 消息：content 可能是字符串
	var textStr string
	if err := json.Unmarshal(content.Content, &textStr); err == nil {
		return textStr
	}

	// assistant 消息：content 是数组
	var blocks []contentBlock
	if err := json.Unmarshal(content.Content, &blocks); err != nil {
		return ""
	}

	var sb strings.Builder
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(b.Text)
		}
	}
	return sb.String()
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}

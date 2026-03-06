package search

import (
	"os"
	"path/filepath"
	"testing"
)

// 构造测试用 JSONL 文件
func writeTempJSONL(t *testing.T, lines ...string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-session.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		f.WriteString(line + "\n")
	}
	f.Close()
	return path
}

func TestScanFile_匹配用户消息(t *testing.T) {
	path := writeTempJSONL(t,
		`{"type":"user","uuid":"u1","sessionId":"s1","cwd":"/proj","timestamp":"2025-01-01T00:00:00Z","message":{"role":"user","content":"hello world"}}`,
		`{"type":"assistant","uuid":"u2","sessionId":"s1","cwd":"/proj","timestamp":"2025-01-01T00:01:00Z","message":{"role":"assistant","content":[{"type":"text","text":"hi there"}]}}`,
		`{"type":"tool_result","sessionId":"s1","timestamp":"2025-01-01T00:02:00Z"}`,
	)

	hits := ScanFile(path, []byte("hello"), 10*1024*1024)
	if len(hits) != 1 {
		t.Fatalf("期望 1 条命中，实际 %d", len(hits))
	}
	if hits[0].Role != "user" {
		t.Errorf("role 应为 user，实际 %s", hits[0].Role)
	}
	if hits[0].MessageUUID != "u1" {
		t.Errorf("uuid 应为 u1，实际 %s", hits[0].MessageUUID)
	}
	if hits[0].SessionID != "s1" {
		t.Errorf("sessionId 应为 s1，实际 %s", hits[0].SessionID)
	}
}

func TestScanFile_匹配助手消息(t *testing.T) {
	path := writeTempJSONL(t,
		`{"type":"assistant","uuid":"a1","sessionId":"s1","cwd":"/proj","timestamp":"2025-01-01T00:00:00Z","message":{"role":"assistant","content":[{"type":"text","text":"这是一段关于搜索的回复"}]}}`,
	)

	hits := ScanFile(path, []byte("搜索"), 10*1024*1024)
	if len(hits) != 1 {
		t.Fatalf("期望 1 条命中，实际 %d", len(hits))
	}
	if hits[0].Role != "assistant" {
		t.Errorf("role 应为 assistant，实际 %s", hits[0].Role)
	}
}

func TestScanFile_不匹配工具消息(t *testing.T) {
	path := writeTempJSONL(t,
		`{"type":"tool_result","sessionId":"s1","cwd":"/proj","timestamp":"2025-01-01T00:00:00Z","message":{"content":"hello"}}`,
	)

	hits := ScanFile(path, []byte("hello"), 10*1024*1024)
	if len(hits) != 0 {
		t.Fatalf("tool_result 不应被搜索，但找到 %d 条", len(hits))
	}
}

func TestScanFile_无匹配(t *testing.T) {
	path := writeTempJSONL(t,
		`{"type":"user","uuid":"u1","sessionId":"s1","cwd":"/proj","timestamp":"2025-01-01T00:00:00Z","message":{"role":"user","content":"hello world"}}`,
	)

	hits := ScanFile(path, []byte("不存在的关键词"), 10*1024*1024)
	if len(hits) != 0 {
		t.Fatalf("不应有匹配，但找到 %d 条", len(hits))
	}
}

func TestScanFile_空文件(t *testing.T) {
	path := writeTempJSONL(t)
	hits := ScanFile(path, []byte("anything"), 10*1024*1024)
	if len(hits) != 0 {
		t.Fatalf("空文件不应有匹配")
	}
}

package search

import (
	"os"
	"path/filepath"
	"testing"

	"cc-jsonl/config"
	"cc-jsonl/index"
	"cc-jsonl/model"
)

// 创建带完整 JSONL 文件的测试环境
func setupTestEngine(t *testing.T) (*Engine, string) {
	t.Helper()
	dir := t.TempDir()

	// 写测试 JSONL
	sessionFile := filepath.Join(dir, "test-session.jsonl")
	f, err := os.Create(sessionFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{
		`{"type":"user","uuid":"msg-1","sessionId":"sess-1","cwd":"/project/foo","timestamp":"2025-06-01T10:00:00Z","message":{"role":"user","content":"如何实现全文搜索？"}}`,
		`{"type":"assistant","uuid":"msg-2","sessionId":"sess-1","cwd":"/project/foo","timestamp":"2025-06-01T10:01:00Z","message":{"role":"assistant","content":[{"type":"text","text":"全文搜索可以用倒排索引或逐行扫描实现"}]}}`,
		`{"type":"user","uuid":"msg-3","sessionId":"sess-1","cwd":"/project/foo","timestamp":"2025-06-01T10:02:00Z","message":{"role":"user","content":"那搜索性能怎么优化？"}}`,
		`{"type":"assistant","uuid":"msg-4","sessionId":"sess-1","cwd":"/project/foo","timestamp":"2025-06-01T10:03:00Z","message":{"role":"assistant","content":[{"type":"text","text":"可以用三层快速路径过滤来优化搜索性能"}]}}`,
	}
	for _, line := range lines {
		f.WriteString(line + "\n")
	}
	f.Close()

	// 构建 store
	store := index.NewStore()
	stat, _ := os.Stat(sessionFile)
	store.Add(model.SessionInfo{
		ID:       "sess-1",
		Project:  "/project/foo",
		FilePath: sessionFile,
		FileSize: stat.Size(),
	})

	cfg := config.Config{
		Workers:      2,
		MaxLineBytes: 10 * 1024 * 1024,
		DefaultLimit: 20,
	}
	engine := NewEngine(store, cfg)
	return engine, dir
}

func TestEngine_Search(t *testing.T) {
	engine, _ := setupTestEngine(t)

	result := engine.Search(model.SearchRequest{Query: "搜索", Limit: 10})
	if result.Total == 0 {
		t.Fatal("应至少有 1 条结果")
	}
	// "搜索" 出现在多条消息中
	if result.Total < 2 {
		t.Errorf("'搜索' 应至少匹配 2 条，实际 %d", result.Total)
	}
}

func TestEngine_Search_分页(t *testing.T) {
	engine, _ := setupTestEngine(t)

	// 先查总数
	full := engine.Search(model.SearchRequest{Query: "搜索", Limit: 100})

	// limit=1 只返回 1 条
	page1 := engine.Search(model.SearchRequest{Query: "搜索", Limit: 1, Offset: 0})
	if len(page1.Results) != 1 {
		t.Errorf("limit=1 应返回 1 条，实际 %d", len(page1.Results))
	}
	if !page1.HasMore {
		t.Error("应有更多结果")
	}
	if page1.Total != full.Total {
		t.Error("Total 应保持一致")
	}
}

func TestEngine_Search_无结果(t *testing.T) {
	engine, _ := setupTestEngine(t)

	result := engine.Search(model.SearchRequest{Query: "完全不存在的关键词xyz", Limit: 10})
	if result.Total != 0 {
		t.Errorf("不应有结果，实际 %d", result.Total)
	}
}

func TestEngine_ReadContext(t *testing.T) {
	engine, _ := setupTestEngine(t)

	ctx, err := engine.ReadContext("sess-1", "msg-2", 1)
	if err != nil {
		t.Fatal(err)
	}
	if ctx == nil {
		t.Fatal("context 不应为 nil")
	}

	// around=1，msg-2 前后各 1 条 → msg-1, msg-2, msg-3
	if len(ctx.Messages) != 3 {
		t.Errorf("around=1 应返回 3 条消息，实际 %d", len(ctx.Messages))
	}
	if ctx.Messages[ctx.HitIndex].UUID != "msg-2" {
		t.Errorf("hitIndex 指向的消息 UUID 应为 msg-2，实际 %s", ctx.Messages[ctx.HitIndex].UUID)
	}
}

func TestEngine_ReadContext_不存在的会话(t *testing.T) {
	engine, _ := setupTestEngine(t)

	ctx, err := engine.ReadContext("nonexistent", "msg-1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if ctx != nil {
		t.Error("不存在的会话应返回 nil")
	}
}

func TestEngine_ReadContext_不存在的UUID(t *testing.T) {
	engine, _ := setupTestEngine(t)

	ctx, err := engine.ReadContext("sess-1", "nonexistent-uuid", 1)
	if err != nil {
		t.Fatal(err)
	}
	if ctx != nil {
		t.Error("不存在的 UUID 应返回 nil")
	}
}

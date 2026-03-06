package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"cc-jsonl/config"
	"cc-jsonl/index"
	"cc-jsonl/model"
	"cc-jsonl/search"
)

// 创建带测试数据的完整 API 服务
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()

	// 写测试 JSONL
	sessionFile := filepath.Join(dir, "test-session.jsonl")
	f, err := os.Create(sessionFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{
		`{"type":"user","uuid":"msg-1","sessionId":"sess-1","cwd":"/project/test","timestamp":"2025-06-01T10:00:00Z","message":{"role":"user","content":"hello world"}}`,
		`{"type":"assistant","uuid":"msg-2","sessionId":"sess-1","cwd":"/project/test","timestamp":"2025-06-01T10:01:00Z","message":{"role":"assistant","content":[{"type":"text","text":"hi there, how can I help?"}]}}`,
		`{"type":"user","uuid":"msg-3","sessionId":"sess-1","cwd":"/project/test","timestamp":"2025-06-01T10:02:00Z","message":{"role":"user","content":"tell me about search"}}`,
	}
	for _, line := range lines {
		f.WriteString(line + "\n")
	}
	f.Close()

	store := index.NewStore()
	stat, _ := os.Stat(sessionFile)
	store.Add(model.SessionInfo{
		ID:       "sess-1",
		Project:  "/project/test",
		FilePath: sessionFile,
		FileSize: stat.Size(),
	})

	cfg := config.Config{
		Workers:      2,
		MaxLineBytes: 10 * 1024 * 1024,
		DefaultLimit: 20,
	}
	engine := search.NewEngine(store, cfg)

	// 用空的 fstest 作为静态文件
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}

	router := NewRouter(cfg, store, engine, staticFS)
	return httptest.NewServer(router)
}

func TestAPI_Search(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/search?q=hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("期望 200，实际 %d", resp.StatusCode)
	}

	var result model.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result.Total == 0 {
		t.Error("应至少有 1 条结果")
	}
	if result.Results[0].SessionID != "sess-1" {
		t.Error("sessionId 应为 sess-1")
	}
}

func TestAPI_Search_缺少参数(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/search")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("缺少 q 参数应返回 400，实际 %d", resp.StatusCode)
	}
}

func TestAPI_Context(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/context/sess-1?uuid=msg-2&around=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("期望 200，实际 %d", resp.StatusCode)
	}

	var ctx model.ContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&ctx); err != nil {
		t.Fatal(err)
	}

	if len(ctx.Messages) != 3 {
		t.Errorf("around=1 应返回 3 条消息，实际 %d", len(ctx.Messages))
	}
	if ctx.Messages[ctx.HitIndex].UUID != "msg-2" {
		t.Error("hitIndex 应指向 msg-2")
	}
}

func TestAPI_Context_缺少UUID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/context/sess-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("缺少 uuid 参数应返回 400，实际 %d", resp.StatusCode)
	}
}

func TestAPI_Context_不存在(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/context/nonexistent?uuid=msg-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("不存在的会话应返回 404，实际 %d", resp.StatusCode)
	}
}

func TestAPI_Stats(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("期望 200，实际 %d", resp.StatusCode)
	}

	var stats model.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}

	if stats.Sessions != 1 {
		t.Errorf("sessions 应为 1，实际 %d", stats.Sessions)
	}
	if stats.TotalSize <= 0 {
		t.Error("totalSize 应大于 0")
	}
}

func TestAPI_StaticFallback(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// 访问根路径应返回 index.html
	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("根路径应返回 200，实际 %d", resp.StatusCode)
	}
}

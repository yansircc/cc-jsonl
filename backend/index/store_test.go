package index

import (
	"testing"

	"cc-jsonl/model"
)

func TestStore_AddAndGet(t *testing.T) {
	s := NewStore()
	info := model.SessionInfo{ID: "abc", Project: "/proj", FileSize: 1024}
	s.Add(info)

	got, ok := s.Get("abc")
	if !ok {
		t.Fatal("应能获取刚添加的会话")
	}
	if got.Project != "/proj" {
		t.Errorf("project 应为 /proj，实际 %s", got.Project)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("不存在的 ID 不应返回 ok")
	}
}

func TestStore_Count(t *testing.T) {
	s := NewStore()
	if s.Count() != 0 {
		t.Error("空 store 的 count 应为 0")
	}
	s.Add(model.SessionInfo{ID: "a", Project: "/p1"})
	s.Add(model.SessionInfo{ID: "b", Project: "/p2"})
	if s.Count() != 2 {
		t.Errorf("count 应为 2，实际 %d", s.Count())
	}
}

func TestStore_Stats(t *testing.T) {
	s := NewStore()
	s.Add(model.SessionInfo{ID: "a", Project: "/p", FileSize: 100})
	s.Add(model.SessionInfo{ID: "b", Project: "/p", FileSize: 200})

	count, totalSize := s.Stats()
	if count != 2 {
		t.Errorf("count 应为 2，实际 %d", count)
	}
	if totalSize != 300 {
		t.Errorf("totalSize 应为 300，实际 %d", totalSize)
	}
}

func TestStore_AllFilePaths(t *testing.T) {
	s := NewStore()
	s.Add(model.SessionInfo{ID: "a", Project: "/p1", FilePath: "/a.jsonl"})
	s.Add(model.SessionInfo{ID: "b", Project: "/p2", FilePath: "/b.jsonl"})
	s.Add(model.SessionInfo{ID: "c", Project: "/p1", FilePath: "/c.jsonl"})

	// 无过滤
	all := s.AllFilePaths("")
	if len(all) != 3 {
		t.Errorf("应返回 3 条，实际 %d", len(all))
	}

	// 按项目过滤
	p1 := s.AllFilePaths("/p1")
	if len(p1) != 2 {
		t.Errorf("p1 应有 2 条，实际 %d", len(p1))
	}

	p3 := s.AllFilePaths("/p3")
	if len(p3) != 0 {
		t.Errorf("p3 应有 0 条，实际 %d", len(p3))
	}
}

func TestStore_Update(t *testing.T) {
	s := NewStore()
	s.Add(model.SessionInfo{ID: "a", Project: "/p1", FileSize: 100})

	s.Update(model.SessionInfo{ID: "a", Project: "/p1", FileSize: 200})
	got, _ := s.Get("a")
	if got.FileSize != 200 {
		t.Errorf("更新后 fileSize 应为 200，实际 %d", got.FileSize)
	}
}

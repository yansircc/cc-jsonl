package index

import (
	"sync"

	"cc-jsonl/model"
)

// Store 线程安全的内存会话索引
type Store struct {
	mu       sync.RWMutex
	sessions map[string]model.SessionInfo // sessionID -> SessionInfo
	byProject map[string][]string         // project -> []sessionID
}

// NewStore 创建空索引
func NewStore() *Store {
	return &Store{
		sessions:  make(map[string]model.SessionInfo),
		byProject: make(map[string][]string),
	}
}

// Add 添加会话信息到索引
func (s *Store) Add(info model.SessionInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[info.ID] = info
	s.byProject[info.Project] = append(s.byProject[info.Project], info.ID)
}

// Get 获取单个会话信息
func (s *Store) Get(id string) (model.SessionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, ok := s.sessions[id]
	return info, ok
}

// Count 返回会话总数
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// Stats 返回会话数量和总文件大小
func (s *Store) Stats() (count int, totalSize int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count = len(s.sessions)
	for _, info := range s.sessions {
		totalSize += info.FileSize
	}
	return
}

// AllFilePaths 返回所有会话的文件路径（搜索用）
func (s *Store) AllFilePaths(project string) []model.SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.SessionInfo
	for _, info := range s.sessions {
		if project != "" && info.Project != project {
			continue
		}
		result = append(result, info)
	}
	return result
}

// Update 更新已有会话信息（文件变化时）
func (s *Store) Update(info model.SessionInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, exists := s.sessions[info.ID]
	s.sessions[info.ID] = info
	if !exists {
		s.byProject[info.Project] = append(s.byProject[info.Project], info.ID)
	} else if old.Project != info.Project {
		// 项目路径变化（不太可能但防御性处理）
		s.removeFromProject(old.Project, info.ID)
		s.byProject[info.Project] = append(s.byProject[info.Project], info.ID)
	}
}

func (s *Store) removeFromProject(project, id string) {
	ids := s.byProject[project]
	for i, sid := range ids {
		if sid == id {
			s.byProject[project] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

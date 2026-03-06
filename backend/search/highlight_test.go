package search

import (
	"testing"
)

func TestExtractSnippet_基本匹配(t *testing.T) {
	text := "这是一段测试文本，包含关键词"
	snippet, highlights := ExtractSnippet(text, "测试", 10)

	if snippet == "" {
		t.Fatal("snippet 不应为空")
	}
	if len(highlights) == 0 {
		t.Fatal("应有高亮位置")
	}
	// 高亮范围应该覆盖 "测试" 两个字符
	h := highlights[0]
	runes := []rune(snippet)
	matched := string(runes[h.Start:h.End])
	if matched != "测试" {
		t.Errorf("高亮内容应为 '测试'，实际为 %q", matched)
	}
}

func TestExtractSnippet_长文本截断(t *testing.T) {
	// 构造一段长文本，关键词在中间
	long := make([]rune, 500)
	for i := range long {
		long[i] = 'A'
	}
	copy(long[250:], []rune("hello"))
	text := string(long)

	snippet, highlights := ExtractSnippet(text, "hello", 20)

	if len(highlights) == 0 {
		t.Fatal("应有高亮位置")
	}
	// snippet 应包含省略号
	runes := []rune(snippet)
	if runes[0] != '.' || runes[1] != '.' || runes[2] != '.' {
		t.Error("长文本截断应有前缀省略号")
	}
}

func TestExtractSnippet_未匹配(t *testing.T) {
	snippet, highlights := ExtractSnippet("hello world", "xyz", 10)
	if snippet == "" {
		t.Error("未匹配时应返回文本开头")
	}
	if highlights != nil {
		t.Error("未匹配时高亮应为 nil")
	}
}

func TestExtractSnippet_中文多次出现(t *testing.T) {
	text := "搜索是搜索的本质，搜索即组织"
	_, highlights := ExtractSnippet(text, "搜索", 50)

	if len(highlights) < 2 {
		t.Errorf("应至少匹配 2 次，实际 %d 次", len(highlights))
	}
}

func TestRuneIndex(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		want     int
	}{
		{"hello", "ll", 2},
		{"你好世界", "世界", 2},
		{"abc", "xyz", -1},
		{"abc", "", 0},
		{"", "a", -1},
	}
	for _, tt := range tests {
		got := runeIndex([]rune(tt.haystack), []rune(tt.needle))
		if got != tt.want {
			t.Errorf("runeIndex(%q, %q) = %d, want %d", tt.haystack, tt.needle, got, tt.want)
		}
	}
}

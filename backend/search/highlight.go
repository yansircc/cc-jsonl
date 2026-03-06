package search

import (
	"strings"

	"cc-jsonl/model"
)

// ExtractSnippet 提取包含关键词的文本片段，计算高亮位置（rune 语义）
// contextChars 是关键词前后各保留的字符数
// 返回的 Highlight.Start/End 是字符偏移（与 JS string.slice 一致）
func ExtractSnippet(text string, queryLower string, contextChars int) (string, []model.Highlight) {
	runes := []rune(text)
	runesLower := []rune(strings.ToLower(text))
	queryRunes := []rune(queryLower)
	queryLen := len(queryRunes)

	idx := runeIndex(runesLower, queryRunes)
	if idx < 0 {
		// 没找到，返回文本开头
		if len(runes) > contextChars*2 {
			return string(runes[:contextChars*2]) + "...", nil
		}
		return text, nil
	}

	// 计算片段范围（rune 级别）
	start := idx - contextChars
	if start < 0 {
		start = 0
	}
	end := idx + queryLen + contextChars
	if end > len(runes) {
		end = len(runes)
	}

	snippet := string(runes[start:end])
	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(runes) {
		suffix = "..."
	}

	result := prefix + snippet + suffix

	// 在结果中查找所有匹配位置（rune 级别）
	resultRunes := []rune(strings.ToLower(result))
	var highlights []model.Highlight
	searchFrom := 0
	for {
		pos := runeIndex(resultRunes[searchFrom:], queryRunes)
		if pos < 0 {
			break
		}
		absPos := searchFrom + pos
		highlights = append(highlights, model.Highlight{
			Start: absPos,
			End:   absPos + queryLen,
		})
		searchFrom = absPos + queryLen
	}

	return result, highlights
}

// runeIndex 在 rune 切片中查找子序列，返回起始 rune 偏移，未找到返回 -1
func runeIndex(haystack, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(haystack) {
		return -1
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

# cc-jsonl

[![CI](https://github.com/yansircc/cc-jsonl/actions/workflows/ci.yml/badge.svg)](https://github.com/yansircc/cc-jsonl/actions/workflows/ci.yml)

Claude Code 对话记录搜索工具。

```
V = C / F
```

JSONL 记录了一切，搜索就是组织。不存储、不转换、不整理，只聚焦。

## 功能

- 全文搜索 `~/.claude/projects/` 下所有 JSONL 对话记录
- 点击结果就地展开上下文消息
- 单二进制部署，内嵌前端

## 快速开始

```bash
make build
./cc-jsonl
# 浏览器打开 http://localhost:3456
```

## 开发

```bash
# 同时启动前后端（热重载）
make dev

# 单独构建前端
cd frontend && bun install && bun run dev

# 生成 TypeScript 类型（Go → TS，SSOT）
make types
```

## 技术栈

- 后端：Go（标准库 HTTP server，零依赖）
- 前端：SvelteKit + SASS
- 类型同步：tygo（Go struct → TypeScript interface）

## 搜索策略

按需扫描，不建索引。三层快速路径过滤：

1. **字节级类型过滤** — 前 1KB 检查 `"type":"user"` / `"type":"assistant"`，跳过 94% 数据
2. **字节级关键词预检** — `bytes.Contains(lower(line), query)`
3. **JSON 解析确认** — 仅对通过前两层的行做完整解析

## License

MIT

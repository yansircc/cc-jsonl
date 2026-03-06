.PHONY: dev dev-frontend dev-backend build clean types

# 开发模式：同时启动前后端
dev: dev-backend dev-frontend

dev-backend:
	cd backend && go run . &

dev-frontend:
	cd frontend && bun run dev

# 构建前端
frontend-build:
	cd frontend && bun run build

# 复制前端产物到 Go embed 目录
copy-ui: frontend-build
	rm -rf backend/ui/dist
	cp -r frontend/build backend/ui/dist

# 构建单二进制
build: copy-ui
	cd backend && go build -o ../cc-jsonl .

# 生成 TypeScript 类型（需要 tygo）
types:
	cd backend && go run github.com/gzuidhof/tygo@latest generate

clean:
	rm -f cc-jsonl
	rm -rf backend/ui/dist
	rm -rf frontend/build
	rm -rf frontend/.svelte-kit

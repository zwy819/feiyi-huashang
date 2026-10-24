# CLAUDE.md

本文件为在本仓库工作的 AI 编码助手提供上下文。面向使用者的说明见 `README.md`。

## 项目概述

**非遗华裳** — 中国非物质文化遗产服饰 3D 数字博物馆 + AI 导览 Agent「织语」。

- 前端：Vue 3 + TypeScript + Vue Router + Vite + TailwindCSS + Three.js + ECharts。
- 后端：Go 1.22+ / Gin / SQLite / JWT / bcrypt / gofpdf。
- AI：OpenAI 兼容模型；Go 进程内 RAG（字符二元组 BM25 + Embedding + RRF）；SSE 流式回答；引用与 3D 动作服务端校验。
- 数据：40 件馆藏、26 个民族、22 个省级地区、36 门工艺、119 个 3D 热点、26 篇带来源资料。

## 常用命令

```bash
# 后端（在 server/ 下执行）
cp .env.example .env
go mod download
go run ./cmd/api -seed      # 空库初始化：40 件馆藏 + 119 个热点
go run ./cmd/api            # 端口 3001
go run ./cmd/api -ingest    # 强制重建 AI 知识库/向量
go test ./...
go build ./cmd/api

# 前端（在 client/ 下执行，必须用 pnpm）
pnpm install
pnpm run dev                # 端口 5173
pnpm run build
pnpm run typecheck
pnpm run verify:ai          # 需要后端 + Vite 已启动
pnpm run eval:ai -- --limit 10
```

## 架构

```text
Browser (5173)
  ├─ /api/* ── Vite proxy ──▶ Gin (3001)
  └─ Vue SPA                     ├─ handlers → services → stores
                                 ├─ internal/ai: ingest / retriever / tools / provider
                                 └─ SQLite: data/nonlegacy.db
```

AI 请求流程：

1. 前端带 `message + history + context` 调用 `POST /api/ai/chat`。
2. 服务端限流、预算检查、页面上下文校验。
3. 结构化过滤 + 馆藏工具 + 混合检索，生成 `[S1]…[S8]` 证据。
4. 模型流式生成；SSE 发送 `meta/tool/sources/delta/action/done/error`。
5. 服务端清理非法引用与非法 3D 动作；`done.content` 为最终文本。

## 关键约束

- API Key 只能放 `server/.env` 或部署 Secret，禁止进入前端或 Git。
- `.env`、`*.db`、`*.glb`、`dist/` 已被忽略；模型通过 GitHub Release 分发。
- AI 回答不是文物鉴定；涉及传说/图腾要写“馆藏说明/常见解释认为”。
- 3D 热点 ID、服饰 ID 必须先查库校验；不能信任模型返回值。
- 前端流式消息必须从 `messages.value` 中取回响应式对象再修改，直接改 `push()` 前的原始对象不会触发 Vue 更新。
- 知识库文章只做事实摘要并保留来源链接，不整段复制原文。

## API 约定

统一响应：`{ "code": 0, "message": null, "data": {} }`；分页从 1 开始；时间格式 `2026-07-01 14:30:00`。AI SSE 是例外，使用 `text/event-stream`。

主要接口：

- `GET /api/ai/status`
- `POST /api/ai/chat`
- `POST /api/ai/feedback`
- `POST /api/admin/ai/reindex`（管理员）
- 原有服饰、统计、收藏、报告、评论、答题接口保持不变。

## 提交规范

使用 Angular 风格：`<type>(<scope>): <subject>`，一个提交只做一件语义独立的事。每次改动后：

```bash
git add <files>
git commit -m "feat(ai): ..."
git push origin HEAD
```

当前远程：`git@github.com:zwy819/feiyi-huashang.git`。

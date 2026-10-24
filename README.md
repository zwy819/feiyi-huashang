# 非遗华裳 · 中国传统服饰 3D 数字博物馆

以数字技术重现千年织造智慧，探索中国非物质文化遗产服饰的纹样之美、工艺之精、文化之魂。

项目包含一个可浏览的 3D 数字博物馆，以及一位会讲解、能指认、可溯源的 AI 导览员「织语」。AI 回答基于馆藏档案、3D 热点讲解和 26 篇带来源的工艺/民族资料，关键结论以 `[S1]` 形式标注来源；当资料不足时会明确拒答，而不是自由发挥。

## 项目亮点

- **40 件非遗服饰**：覆盖 26 个民族、22 个省级地区、36 门工艺、5 个朝代。
- **3D 数字展厅**：41 个 GLB 模型、119 个可点击热点，支持旋转、缩放、背景切换和分享图。
- **AI 导览 Agent**：混合检索（字符二元组 BM25 + Embedding + RRF）、馆藏工具调用、SSE 流式回答、引用校验、3D 热点指认。
- **可信与可恢复**：无 Key、限流、超时、空检索、SSE 中断均有独立状态；API Key 只保存在服务端。
- **可评测**：Go 单元测试、前端类型检查、Playwright 浏览器验收脚本。

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + TypeScript + Vue Router + Vite + TailwindCSS |
| 3D | Three.js（GLTF/GLB，原生渲染） |
| 图表 | ECharts 5 |
| 后端 | Go 1.22+ / Gin / SQLite / JWT / bcrypt / gofpdf |
| AI | OpenAI 兼容接口，默认阿里云百炼 Qwen；进程内 RAG，无外部向量库 |

## 快速开始

### 一键启动（Windows）

安装 Go、Node.js、pnpm 后，双击仓库根目录的 `start.bat`。脚本会自动：

1. 从 `.env.example` 创建 `server/.env`（如果不存在）；
2. 启动 Go 后端，空数据库时自动 `-seed`；
3. 安装前端依赖（首次）并启动 Vite；
4. 等待健康检查通过后打开 `http://localhost:5173/`。

模型资产不会自动下载；需要 3D 展示时先执行 `scripts/fetch-models.ps1`。

### 手动启动

#### 环境要求

- Go 1.22+
- Node.js 18+
- pnpm 8+
- Windows 下编译 `mattn/go-sqlite3` 需要 CGO 与 GCC（MinGW-w64）

### 1. 启动后端

```bash
cd server
cp .env.example .env
# 编辑 .env，至少确认 PORT / JWT_SECRET；需要 AI 时填写 AI_API_KEY
go mod download
go run ./cmd/api -seed
```

首次运行会建表、插入 40 件馆藏与示例账号，并自动建立 AI 知识库索引。默认端口 `3001`。

- 管理员：`admin` / `admin123`
- 普通用户：`user` / `user123`

### 2. 启动前端

```bash
cd client
pnpm install
pnpm run dev
```

打开 `http://localhost:5173/`。Vite 会把 `/api` 代理到 `http://localhost:3001`。

### 3. 获取 3D 模型（可选，但推荐）

模型约 230 MB，不在 Git 历史中。在仓库根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\fetch-models.ps1
```

脚本从 GitHub Release 下载并校验 41 个 GLB。没有模型时，浏览、统计、AI 和报告仍可用，只是 3D 视口显示占位。详见 [docs/ASSETS.md](docs/ASSETS.md)。

### 4. 配置 AI

在 `server/.env` 中填写（该文件已被 `.gitignore` 忽略）：

```dotenv
AI_ENABLED=true
AI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
AI_API_KEY=你的Key
AI_CHAT_MODEL=qwen-plus
AI_EMBEDDING_MODEL=text-embedding-v3
AI_MOCK_MODE=false
```

没有 Key 时可以设置 `AI_MOCK_MODE=true`，用内置假模型验证前后端链路；评审演示应使用真实模型。

重建知识库（修改资料后）：

```bash
go run ./cmd/api -ingest
# 或管理员调用 POST /api/admin/ai/reindex
```

## AI 导览工作流

```text
用户问题 + 页面上下文 + 最近历史
        │
        ▼
服务端校验 / 限流 / 日预算
        │
        ├─ 结构化过滤（民族/省份/工艺/朝代）
        ├─ 馆藏工具（检索、详情、热点、相似、统计）
        └─ 混合检索（BM25 + 向量 + RRF）
        │
        ▼
证据编号 [S1]…[S8] → 提示词 → 模型流式生成
        │
        ▼
引用校验 → 3D 动作校验 → SSE 事件
```

SSE 事件：`meta`、`tool`、`sources`、`delta`、`action`、`done`、`error`。前端用 `fetch` + `ReadableStream` 解析，不依赖 `EventSource`。

安全设计：

- Key 只读服务端环境变量，不进入前端构建产物。
- 匿名限流（默认 10 次/分钟/IP）与全站日 Token 预算。
- 只记录会话哈希、模型、Token、延迟、工具数量，不记录完整对话。
- 引用编号必须能在本轮来源中找到，否则服务端删除；3D 热点 ID 必须真实存在且属于对应服饰。
- 检索资料只作为资料处理，提示词明确抵抗资料中的注入指令。

## API 一览

所有 JSON 接口沿用 `{ "code": 0, "message": null, "data": {} }` 结构。

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/api/health` | 健康检查 |
| `GET` | `/api/garments` | 服饰列表（分页/筛选/搜索） |
| `GET` | `/api/garments/:id` | 服饰详情 |
| `GET` | `/api/garments/:id/annotations` | 3D 热点标注 |
| `GET` | `/api/statistics/overview` | 馆藏概览 |
| `GET` | `/api/statistics/geo` | 地理分布 |
| `POST` | `/api/report/generate` | 生成 PDF 收藏报告 |
| `GET` | `/api/ai/status` | AI 配置与知识库状态 |
| `POST` | `/api/ai/chat` | SSE 流式导览 |
| `POST` | `/api/ai/feedback` | 回答点赞/点踩 |
| `POST` | `/api/admin/ai/reindex` | 管理员重建知识库 |

完整接口见 [docs/API.md](docs/API.md)。

## 目录结构

```text
nonlegacy-garment-3d/
├── client/                 Vue 3 前端
│   ├── src/components/AI/  AI 面板、消息、来源抽屉、工具轨迹
│   ├── src/composables/    useAiChat 会话状态
│   ├── src/services/ai.ts  SSE 解析
│   └── scripts/            Playwright 验收脚本
├── server/                 Go 后端
│   ├── cmd/api/            入口、种子数据、AI 初始化
│   ├── internal/ai/        知识摄取、BM25、检索、工具、Provider、SSE
│   ├── internal/handler/   HTTP 处理
│   ├── knowledge/articles/ 26 篇带来源资料
│   └── data/nonlegacy.db   本地 SQLite（不提交）
├── docs/                   计划、API、资产说明
└── tools/                  标注与开发辅助脚本
```

## 测试与验收

```bash
# 后端
cd server
go test ./...
go build ./cmd/api

# 前端
cd client
pnpm run typecheck
pnpm run build

# 浏览器验收（需要本地后端与 Vite 已启动）
pnpm run verify:ai
```

验收脚本会检查：

1. 首页 AI 面板能打开并返回带引用的回答；
2. 来源抽屉能展示来源，且 3D 可定位来源有标记；
3. 详情页点击「在 3D 中查看」后，对应热点被激活并显示讲解卡。

## 资产与许可

- 代码以 MIT 许可发布，详见 [LICENSE](LICENSE)。
- 3D 模型约 230 MB，未纳入 Git；复现 3D 展示时需将 GLB 放入 `client/public/models/`。模型为数字复原/艺术化展示，不是文物级扫描。
- 知识库文章只做事实性摘要并链接原始发布方，不整段复制原文；原文版权归各发布方所有。
- 字体、模型生成工具和第三方素材的授权需在提交前逐项核对，详见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

## 常见问题

### AI 状态显示“未配置”

确认 `server/.env` 中 `AI_API_KEY` 非空，并重启后端。没有 Key 时普通浏览、统计、报告仍可用，只是 AI 入口会提示未配置。

### 模型返回 401 Incorrect API key

- **百炼（DashScope）Key**：通常以 `sk-` 开头，Base URL 用 `https://dashscope.aliyuncs.com/compatible-mode/v1`。
- **QwenCloud Key**：以 `sk-ws-` 开头，必须使用创建 Key 的同一区域端点，例如新加坡：
  `https://ap-southeast-1.maas.qwencloudapi.com/compatible-mode/v1`。
- 不要把 Qwen Code 的 OAuth/登录令牌当成 API Key；请到控制台的 API Keys 页面重新创建。
- 区域不匹配、Key 复制不完整或已删除都会返回 401。修改后执行 `go run ./cmd/api -ingest` 重建向量索引。

### 3D 视口显示“暂无 3D 模型”

模型资产未下载。执行 `scripts/fetch-models.ps1`，确认 `client/public/models/` 下有 41 个 GLB。

### 首次启动很慢

`go run` 需要编译 CGO/SQLite，首次可能 1–2 分钟；前端首次 `pnpm install` 也需要时间。之后启动会明显加快。

## 已知边界

- 未配置真实模型 Key 时，AI 入口会显示“未配置”，普通浏览不受影响。
- 向量检索依赖 Embedding 接口；接口不可用时自动退化为 BM25，不影响回答流程。
- 当前不提供照片识衣（P1）；按比赛计划，P0 稳定后再评估。
- 高峰期或无网络环境下，模型首 Token 可能超过 3 秒；服务端提供超时与重试状态。

## 比赛信息

第九届传智杯 AI WEB 网页开发挑战赛作品。详细改进计划见 [docs/传智杯-AI-WEB-改进计划.md](docs/传智杯-AI-WEB-改进计划.md)。

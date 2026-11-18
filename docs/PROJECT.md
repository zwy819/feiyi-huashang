# 非遗华裳系统架构

更新日期：2026-11-16。

非遗华裳面向服饰文化学习与浏览，把馆藏检索、带来源的 AI 问答和 3D 热点讲解连接起来。访客可以匿名浏览与提问；登录后使用收藏、报告、评论和投稿；管理员维护馆藏、模型与审核内容。

## 运行架构

```text
浏览器：Vue 3 + TypeScript
  ├─ 路由页面、AI 面板、Three.js、ECharts
  └─ /api → Vite 代理 → Go / Gin :3001
                           ├─ Handler → Service → Store
                           ├─ AI：工具规划、混合检索、SSE
                           ├─ TTS：语音缓存与浏览器语音回退
                           └─ SQLite、模型、报告与语音文件
```

前端开发端口为 5173，固定端口被占用时启动失败并提示。后端主要承担数据、鉴权与模型接入，前端承担可视化、会话状态和 3D 交互。

## 代码边界

| 目录 | 职责 |
| --- | --- |
| `client/src/pages/` | 页面取数、路由上下文与用户操作协调 |
| `client/src/components/3D/ModelViewer.vue` | Three.js 场景、模型加载、相机飞行、资源释放与公开控制接口 |
| `client/src/composables/useModelHotspots.ts` | 热点投影、遮挡检测、选择状态 |
| `client/src/components/3D/TourControls.vue` | 导览控制条展示与交互事件 |
| `client/src/components/AI/` | 会话、来源抽屉、回答与工具轨迹展示 |
| `client/src/composables/useAiChat.ts` | 会话、流式状态、上下文和中断 |
| `client/src/composables/useNarration.ts` | 同一页面共享的语音播放状态，协调服务端 TTS 与浏览器语音 |
| `client/src/services/` | REST、SSE 与路线接口 |
| `server/internal/handler/` | 参数、鉴权上下文、HTTP 返回 |
| `server/internal/service/` | 认证、报告、模型、统计等业务 |
| `server/internal/model/` | 实体与参数化 SQL 数据访问 |
| `server/internal/ai/` | 知识摄取、检索、模型接入、工具执行与引用编号校验 |
| `server/internal/tour/` | 根据馆藏与热点规划参观路线 |
| `server/internal/tts/` | 语音生成和缓存 |

详情页保留导览状态、路由和语音的统一协调；模型场景负责创建与释放 Three.js 资源，热点投影模块通过 getter 读取这些对象，不把渲染器放入 Vue 深层响应式状态。

## AI 问答

1. 校验问题、历史和页面上下文；检查频率与日 Token 预算。
2. 开启 `AI_AGENT_ENABLED` 时，聊天模型通过 Function Calling 自主选择只读工具，最多 2 轮、4 次执行；工具结果回传给规划模型，参数错误时允许调整查询。关闭后继续使用原有 RAG 问答流程。
3. 服务端补充明确的馆藏过滤、SQL 统计、详情和热点证据。
4. 以字符二元组 BM25 与 Embedding 检索，使用 RRF 融合；去重后保留最多 8 项证据。
5. 编号 `[S1]` 至 `[S8]`，连同用户问题和页面上下文发送给模型，流式生成回答。
6. 校验来源编号；校验热点 ID 是否存在于本轮证据。模型明确指定热点或回答提及热点标题时，才生成 3D 动作。
7. 前端呈现来源、工具完成/失败状态与动作；用户点击后定位对应模型细节。

模型可选工具为 `search_garments`、`get_garment_detail`、`get_annotations`、`find_similar`、`get_statistics`。工具只读取本馆数据库，模型不能执行 SQL 或访问任意 URL。模型需支持 OpenAI 兼容 Function Calling；Mock 模式用于开发链路检查。

来源编号校验检查引用句柄存在于本轮证据。事实正确性与来源是否支持相邻论断，通过独立评测和人工核验记录，方法见 [EVALUATION.md](EVALUATION.md)。

### 向量索引

知识来自馆藏字段、3D 热点讲解和 `server/knowledge/articles/` 的带来源摘要。现有种子馆藏为 40 件、119 个热点、26 篇资料，当前运行索引为 185 个知识块；实际数量随内容变化。

文档内容或 Embedding 模型变化、向量缺失时，摄取流程重新生成相应分块与向量。查询只比较同模型、同维度的向量；向量接口失败时使用 BM25。`-ingest` 和管理员重建接口强制刷新知识库，不重置馆藏。

### 统计含义

`total` 计算服饰总数或非空字段的不同类别数，支持民族/省份/工艺/朝代过滤；`ranking` 返回按件数排序的前 8 个分类。工艺按完整字段分组，“蜡染+刺绣”不会被自动拆成两个工艺记录。页面图表和 AI 统计均描述本项目馆藏。

## 3D 与语音

模型优先通过 `/api/models/:id/file` 读取；旧 `model_url` 入口继续可用。模型查看器支持旋转、缩放、热点、背景切换、镜头飞行和分享截图。导览控制条与详情页共用语音状态，避免简介和导览同时播放。

Qwen-TTS 音频缓存在 `server/data/tts/`，失败时回退到浏览器中文语音。热点投影复用 Vector3 和 Raycaster，每帧更新屏幕坐标；遮挡或视口外的点位禁用点击。

## 启动与数据

Windows 免安装发行版将 Vue 生产页面、模型、知识库、馆藏初始化数据与 Go 服务打成一个 EXE，使用 `CGO_ENABLED=0`，直接在本机 3001 端口提供页面和 API。双击中文 `一键启动.bat` 或 EXE，无需 Go、Node.js 或 pnpm。运行数据和可编辑配置释放到 EXE 旁，首次建立 40 件馆藏、40 条模型关联与 119 个热点。见 [免安装启动](免安装启动.md)。

源码开发需要 Go 1.22+、Node.js 18+ 和 pnpm。SQLite 使用纯 Go 驱动 `modernc.org/sqlite`，保留 WAL 和外键约束。`开发启动.bat` 以 `CGO_ENABLED=0` 启动 Go 服务与 Vite，并等待健康检查。仓库和外层的 `一键启动.bat` 优先发行版，缺少发行版时转入源码开发；`start.bat` 是兼容入口。

手动运行：

```powershell
cd server
go run ./cmd/api

# 另开终端
cd client
pnpm install --frozen-lockfile
pnpm run dev
```

首次从 `.env.example` 创建 `.env`，配置 JWT 与 AI。`-seed` 会重建馆藏、收藏、模型数据与模型目录，只在首次初始化或明确重置时使用。

数据库在 `server/data/nonlegacy.db`，上传模型在 `server/data/models/`，模型源文件在 `client/public/models/`，报告在 `server/reports/`。本地配置、数据库、模型原件、缓存和构建产物不进入 Git；资产获取见 [ASSETS.md](ASSETS.md)。

## 验证入口

```powershell
# server/，与 CI 一样关闭 CGO
$env:CGO_ENABLED = "0"
go test ./...
go build ./cmd/api

# client/
pnpm run lint
pnpm run typecheck
pnpm run build
pnpm verify:ai
```

CI 执行 Go 测试与编译、前端 lint/类型检查/构建。浏览器验收需要运行中的前后端；AI 质量评测使用 [EVALUATION.md](EVALUATION.md) 中的 60 题流程。

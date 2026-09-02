# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

**非遗华裳** — 中国非物质文化遗产服饰 3D 数字博物馆。以 Vue 3 + Three.js 前端配合 Go + Gin + SQLite 后端，展示非遗服饰的纹样、工艺与地理分布。

## 常用命令

```bash
# 后端（server/ 目录）— Go + Gin
cd server
cp .env.example .env          # 首次运行前必须配置
go mod tidy                   # 安装后端依赖
go run ./cmd/api -seed        # 插入种子数据（5 件服饰）
go run ./cmd/api              # 启动后端开发服务器 (端口 3001)
go test ./...                 # 运行后端测试

# 前端（client/ 目录）— 必须使用 pnpm
cd client
pnpm install                  # 安装依赖
pnpm run dev                  # 启动 Vite 开发服务器 (端口 5173)
pnpm run build                # 生产构建 → dist/
pnpm run preview              # 预览生产构建
pnpm run lint                 # ESLint 检查（注意：ESLint 未安装为 devDependency，当前无法运行）
```

前端开发时 Vite 代理 `/api` → `http://localhost:3001`，无需单独配置 CORS。生产环境中 Gin 可通过 `r.Static()` 托管前端构建产物。

## 整体架构

```
Browser (localhost:5173)
   │
   ├─ /api/*  ──(Vite proxy)──▶ Gin API Server (localhost:3001)
   │                                      │
   └─ /* (Vue SPA)                        ├─ middleware: CORS / JWT / Recovery
                                          ├─ handlers: auth / garments / daily
                                          │            favorites / statistics / report
                                          ├─ services: 业务逻辑
                                          └─ models + SQLite
                                                       文件: server/data/nonlegacy.db
```

- **前端**: Vue 3 + Vue Router 4（7 个路由页面）+ Vue Transition / CSS 动画 + ECharts 5 图表
- **3D 渲染**: Three.js 原生（非 React Three Fiber），`ModelViewer.vue` 封装场景/相机/光照，`ModelLoader.js` 手动管理 GLTF/GLB 加载生命周期
- **后端**: Go + Gin 经典三层结构（handler → service → model）
- **数据库**: Go SQLite driver 直接读写，无需 WASM 兼容层或手动 `saveToDisk()`

## API 通用约定

所有 JSON 响应统一使用三字段结构：

```json
{
  "code": 0,
  "message": null,
  "data": {}
}
```

- `code: 0` → 成功；非 0 → 业务错误（参见下方业务码表）
- `message` → 成功时为 `null`（创建/更新/删除操作可返回提示文案），错误时为具体错误信息
- `data` → 业务数据，无数据时返回 `null`

### 业务码一览

| code | HTTP 状态码 | 含义 |
|------|-------------|------|
| `0` | `2xx` | 成功 |
| `40001` | `400` | 参数缺失或格式错误 |
| `40002` | `400` | 参数值不合法 |
| `40101` | `401` | 未提供认证令牌 |
| `40102` | `401` | 认证令牌无效或已过期 |
| `40103` | `401` | 用户名或密码错误 |
| `40301` | `403` | 无权操作该资源 |
| `40401` | `404` | 资源不存在 |
| `40901` | `409` | 资源冲突或重复操作 |
| `50001` | `500` | 服务器内部错误 |

### 分页规范

- 页码从 1 开始（1-based），`limit` 默认 12，上限 100
- 响应始终包含 `total`、`page`、`limit`
- 前端通过 `Math.ceil(total / limit)` 计算总页数

### 日期时间格式

全部使用 ISO 字符串，精确到秒，本地时间：`2026-07-01 14:30:00`

## 接口汇总表

| # | 方法 | 路径 | 说明 | 鉴权 | 状态 |
|---|------|------|------|------|------|
| 1 | `POST` | `/api/auth/register` | 用户注册 | — | ✅ |
| 2 | `POST` | `/api/auth/login` | 用户登录 | — | ✅ |
| 3 | `GET` | `/api/auth/profile` | 获取用户信息 | JWT | ✅ |
| 4 | `GET` | `/api/garments` | 服饰列表（分页/筛选/搜索） | — | ✅ |
| 5 | `GET` | `/api/garments/:id` | 服饰详情 | — | ✅ |
| 6 | `GET` | `/api/garments/:id/recommend` | 智能推荐链 | — | ✅ |
| 7 | `GET` | `/api/daily` | 每日一服 | — | ✅ |
| 8 | `GET` | `/api/favorites` | 收藏列表 | JWT | ✅ |
| 9 | `POST` | `/api/favorites` | 添加收藏 | JWT | ✅ |
| 10 | `DELETE` | `/api/favorites/:garment_id` | 移除收藏 | JWT | ✅ |
| 11 | `GET` | `/api/favorites/check/:garment_id` | 检查收藏状态 | JWT | ✅ |
| 12 | `GET` | `/api/statistics/geo` | 地理分布数据 | — | ✅ |
| 13 | `GET` | `/api/statistics/craft-flow` | 工艺流向数据 | — | ✅ |
| 14 | `GET` | `/api/statistics/overview` | 概览统计 | — | ✅ |
| 15 | `GET` | `/api/report/generate?ids=1,2,3` | 生成 PDF 报告 | — | ✅ |
| 16 | `GET` | `/api/health` | 健康检查 | — | ✅ |
| 17 | `GET` | `/api/comments` | 服饰评论列表 | — | 📋 规划中 |
| 18 | `POST` | `/api/comments` | 发表评论 | JWT | 📋 规划中 |
| 19 | `PUT` | `/api/comments/:id` | 编辑评论 | JWT | 📋 规划中 |
| 20 | `DELETE` | `/api/comments/:id` | 删除评论 | JWT | 📋 规划中 |
| 21 | `POST` | `/api/comments/:id/like` | 点赞评论 | JWT | 📋 规划中 |
| 22 | `DELETE` | `/api/comments/:id/like` | 取消点赞 | JWT | 📋 规划中 |
| 23 | `GET` | `/api/statistics/province/:name` | 省份下钻数据 | — | 📋 规划中 |
| 24 | `GET` | `/api/statistics/sankey` | 朝代-民族-工艺流向 | — | 📋 规划中 |
| 25 | `GET` | `/api/statistics/by-ethnic` | 按民族统计 | — | 📋 规划中 |
| 26 | `GET` | `/api/statistics/by-dynasty` | 按朝代统计 | — | 📋 规划中 |
| 27 | `GET` | `/api/garments/:id/image` | 服饰预览图 | — | 📋 规划中 |

> 详细请求/响应格式参见 `docs/API(1).md`。

## 项目目录结构

```
nonlegacy-garment-3d/
├── client/                          # 前端 Vue SPA（使用 pnpm）
│   ├── public/
│   │   └── models/                  # 3D 模型文件 (.glb)
│   │       ├── miao-skirt.glb       #   苗族百褶裙 3D 模型
│   │       └── dai-skirt.glb        #   傣族筒裙 3D 模型
│   ├── src/
│   │   ├── main.js                  # 应用入口：createApp + router 挂载
│   │   ├── App.vue                  # 根组件：全局布局出口
│   │   ├── router/
│   │   │   └── index.js             # Vue Router 路由定义 + 登录守卫
│   │   ├── components/
│   │   │   ├── Layout/              # Layout.vue / Navbar.vue / Footer.vue
│   │   │   ├── Common/              # Card.vue / Modal.vue / Skeleton.vue / EmptyState.vue / Toast.vue
│   │   │   └── 3D/                  # ModelViewer.vue / ModelLoader.js
│   │   ├── pages/
│   │   │   ├── Home/HomePage.vue    # 首页：Hero + 每日一服 + 模块入口
│   │   │   ├── Explore/ExplorePage.vue  # 服饰浏览：搜索 + 筛选 + 分页网格
│   │   │   ├── Detail/DetailPage.vue    # 服饰详情：3D 模型 + 属性 + 推荐链
│   │   │   ├── Favorites/FavoritesPage.vue  # 我的收藏
│   │   │   ├── Statistics/StatisticsPage.vue # 地理分布：气泡图 + 饼图 + 概览
│   │   │   ├── Report/ReportPage.vue    # 导出报告：预览 + PDF 生成下载
│   │   │   └── Auth/AuthPage.vue        # 登录注册页
│   │   ├── composables/
│   │   │   └── index.js             # useFetch / useAuth / useDebounce / useToast
│   │   ├── services/
│   │   │   └── api.js               # Axios 封装 + 6 组 API 函数 + 拦截器
│   │   ├── utils/
│   │   │   └── index.js             # formatDate / 省坐标 / 色系常量
│   │   └── styles/
│   │       └── index.css            # TailwindCSS + 自定义组件样式
│   ├── index.html
│   ├── vite.config.js               # Vite 配置（代理 + 别名）
│   ├── tailwind.config.js           # TailwindCSS 配置（莫高窟色系）
│   └── package.json
├── server/                          # Go + Gin 后端
│   ├── cmd/api/main.go              # Gin 启动入口
│   ├── internal/
│   │   ├── config/                  # 环境变量读取
│   │   ├── db/                      # SQLite 连接与迁移
│   │   ├── handler/                 # Gin handlers
│   │   ├── middleware/              # CORS / JWT / Recovery
│   │   ├── model/                   # 数据结构与查询
│   │   └── service/                 # 业务逻辑
│   ├── data/nonlegacy.db            # SQLite 数据库文件（已被 .gitignore 忽略）
│   ├── reports/                     # PDF 报告输出目录（已被 .gitignore 忽略）
│   ├── .env                         # 环境变量（不可提交）
│   ├── .env.example                 # 环境变量模板
│   └── go.mod
├── clothes/                         # 3D 模型源文件目录
│   ├── miao-skirt.glb
│   └── dai-skirt.glb
├── docs/                            # 项目文档
│   ├── PROJECT(1).md                # 项目文档
│   └── API(1).md                    # API 接口文档
├── CLAUDE.md                        # 本文件
└── README.md
```

## 数据库

四张表：`users`、`garments`、`favorites`、`daily_picks`。规划中还有 `comments` 和 `comment_likes` 表。

- **users** — `id` / `username` (UNIQUE) / `email` (UNIQUE) / `password_hash` (bcrypt) / `avatar_url` / `created_at` / `updated_at`
- **garments** — `id` / `name` / `dynasty` / `ethnic` / `craft` / `material` / `pattern` / `cultural_meaning` / `province` / `description` / `image_url` / `model_url` / `created_at` / `updated_at`
- **favorites** — `id` / `user_id` (FK→users) / `garment_id` (FK→garments) / `created_at`，UNIQUE(user_id, garment_id)
- **daily_picks** — `id` / `garment_id` (FK→garments) / `pick_date` (UNIQUE) / `created_at`

Go 后端通过 SQLite driver 直接读写数据库文件，写操作由 SQLite 自身持久化。

## 前端数据流

1. **API 层** (`services/api.js`) — Axios 实例，baseURL `/api`，请求拦截器自动附加 JWT（从 localStorage 读取），响应拦截器解包 `.data`，401 时清除 token 并派发 `auth:required` 事件弹出登录框
2. **Composables** (`composables/index.js`) — `useFetch(fetcher, deps)` 通用异步获取；`useAuth()` 管理登录状态（localStorage 存取 token + user）；`useDebounce(value, delay)` 搜索防抖；`useToast()` 通知管理
3. **页面组件** (`pages/`) — 每个页面直接调用 composables 和 API 函数，状态提升到页面级别
4. **通用组件** (`components/Common/`) — Card、Modal、Skeleton、EmptyState、Toast 五种 UI 组件
5. **3D 组件** (`components/3D/`) — `ModelViewer.vue` 封装 Three.js 场景（相机、光照、OrbitControls），`ModelLoader.js` 手动管理 GLB 加载生命周期

## 鉴权流程

- JWT 密钥配置在 `server/.env` → `JWT_SECRET`，默认 `nonlegacy_dev_secret_key`，有效期 `168h`（7 天）
- Gin JWT 中间件 → 解析 Token 并将用户信息写入 `gin.Context`
- 强制鉴权接口（收藏相关）→ 无 Token 或 Token 无效返回 401
- 公开接口可选择性读取 Token 识别已登录用户
- 客户端 `useAuth()` 的 `login(token, userData)` 将 token 和用户信息存入 localStorage
- Axios 请求拦截器自动从 localStorage 读取 token 附加到 `Authorization: Bearer <token>` 请求头

### 路由鉴权矩阵

| 路由 | 鉴权方式 | 未登录行为 |
|------|----------|-----------|
| `/api/auth/register` `/api/auth/login` | 无 | — |
| `/api/auth/profile` | JWT 强制 | 返回 401 |
| `/api/garments` `/api/garments/:id` `/api/garments/:id/recommend` | 无 | — |
| `/api/daily` | 无 | — |
| `/api/favorites` 全部 | JWT 强制 | 返回 401 |
| `/api/statistics/*` | 无 | — |
| `/api/report/generate` | 无（需提供 ids） | — |

## 关键业务规则

- **每日一服** (`/api/daily`) — 按日期唯一，每天仅推送一件；先查 `daily_picks` 表今日是否有记录，有则返回已推送的，无则随机选一件并记录
- **智能推荐链** (`/api/garments/:id/recommend`) — 基于"民族 + 工艺"匹配度计算相似性：两者都匹配 relevance=2，任一匹配=1，都不匹配=0。按 relevance 降序、ID 降序排列，最多返回 4 条
- **收藏** — 同一用户对同一服饰仅可收藏一次（UNIQUE(user_id, garment_id)），重复收藏返回 409
- **PDF 报告** — 由 Go 后端（gofpdf 或同类库）根据收藏列表生成，API 返回 PDF 二进制流（`Content-Type: application/pdf`），前端触发下载

## 前端错误处理

Axios 响应拦截器统一处理：
- **401** → 清除 localStorage token + user → 派发 `auth:required` 事件 → Toast 提示 "登录已过期" → 弹窗要求重新登录
- **其他错误** → 优先使用 `response.data.message`，可按 `response.data.code` 做分支处理
- 各页面另有 `catch` 降级逻辑，在 API 不可用时使用本地硬编码数据保证页面基本可用

## 配色方案

敦煌莫高窟色系，在 `tailwind.config.js` 和 `utils/index.js` 中定义（两处需保持同步）：

| 色键 | 色值 | 应用场景 |
|------|------|----------|
| `cinnabar` (朱砂红) | `#C43A31` | 标题、主按钮、强调、收藏心形 |
| `malachite` (石绿) | `#2E7D5A` | 辅助点缀、成功状态 |
| `ochre` (土黄) | `#C89B5E` | 背景、边框、卡片、次按钮 |
| `lapis` (青金石蓝) | `#2E4A7A` | 正文、深色背景、图表文字 |
| `leadwhite` (铅白) | `#F5F0E8` | 页面底色 |
| `gold` (金) | `#B8862D` | 高亮强调、分隔线 |

每个色键包含 50-900 色阶。

## 种子数据

当前 5 件服饰：

| ID | 名称 | 民族 | 朝代 | 工艺 | 省份 | 3D 模型 |
|----|------|------|------|------|------|---------|
| 1 | 苗族百褶裙 | 苗族 | 清 | 蜡染+刺绣 | 贵州 | ✅ miao-skirt.glb |
| 2 | 苏绣旗袍 | 汉族 | 现代 | 苏绣 | 江苏 | — |
| 3 | 藏族氆氇袍 | 藏族 | 清 | 氆氇织造 | 西藏 | — |
| 4 | 蒙古族长袍 | 蒙古族 | 元 | 镶边刺绣 | 内蒙古 | — |
| 5 | 傣族筒裙 | 傣族 | 清 | 织锦 | 云南 | ✅ dai-skirt.glb |

规划扩充至 30 件，覆盖更多民族（维吾尔族、壮族、彝族等）、工艺（缂丝、扎染、挑花等）和朝代（唐、宋、明），每件配套 `.glb` 3D 模型 + 静态预览图。

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | 3001 | 后端服务端口 |
| `APP_ENV` | development | 运行环境（development 时详细错误只写服务端日志） |
| `DB_PATH` | ./data/nonlegacy.db | SQLite 数据库文件路径（相对 server/） |
| `JWT_SECRET` | nonlegacy_dev_secret_key | JWT 签名密钥（生产必须更换） |
| `JWT_EXPIRES_IN` | 168h | JWT 有效期 |
| `CORS_ORIGIN` | http://localhost:5173 | 允许的跨域来源 |

## 注意事项

- **后端必须在 `server/` 目录下运行** — `.env` 中的相对路径依赖此前提
- **前端必须使用 pnpm** — 禁止 npm 或 yarn
- **数据库文件在 `.gitignore` 中** — `server/data/nonlegacy.db` 和 `server/reports/` 不提交
- **`.env` 不可提交** — `.env.example` 是模板文件
- **后端使用 Go + Gin** — 不再使用 Express / sql.js
- **JSON 响应统一结构** — `{ "code": 0, "message": null, "data": {} }`，错误使用业务码并保留正确 HTTP 状态码
- **ESLint 未安装为 devDependency** — lint 脚本暂时无法运行（已知问题）
- **当前无测试框架、无 Docker 配置**
- **3D 模型开发环境由 Vite 静态文件服务提供** — 路径 `/models/<name>.glb` 对应 `client/public/models/`

## 包管理器约束

- **前端 (`client/`)**：必须使用 **pnpm**，禁止 npm 或 yarn
- **后端 (`server/`)**：使用 **Go modules**（`go mod tidy`）

## Git 提交规范

每次与 AI 对话结束时，必须将当次所有改动提交到 Git，遵循以下规则：

- **提交风格**：遵循 [Angular Commit Message Convention](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)，格式为 `<type>(<scope>): <subject>`
  - `type`: feat / fix / refactor / style / docs / test / chore
  - `scope`: client / server / shared（按改动范围选择）
  - `subject`: 中文简述改动内容
- **最小化提交**：每次只提交一个语义独立的改动，**禁止**将多个不相关的改动积攒为一次大批量提交。一个改动 = 一个 commit
- **示例**：`feat(client): 添加服饰详情页 3D 模型预览`、`fix(server): 修复每日一服日期比较时区问题`
- **提交前检查**：确保代码无语法错误，再执行 `git add` + `git commit`
- 如果项目尚未初始化 Git 仓库，首次对话时需执行 `git init` 并完成初始提交

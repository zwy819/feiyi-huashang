# 非遗华裳 — 项目文档

> 中国非物质文化遗产服饰 3D 数字博物馆
>
> 最后更新: 2026-07-06

---

## 目录

- [1. 项目概述](#1-项目概述)
- [2. 技术架构](#2-技术架构)
- [3. 项目结构](#3-项目结构)
- [4. 后端架构](#4-后端架构)
- [5. 前端架构](#5-前端架构)
- [6. 数据库设计](#6-数据库设计)
- [7. 后台管理面板](#7-后台管理面板)
- [8. 3D 模型渲染管线](#8-3d-模型渲染管线)
- [9. 认证与鉴权](#9-认证与鉴权)
- [10. 配色方案](#10-配色方案)
- [11. 开发指南](#11-开发指南)
- [12. 部署说明](#12-部署说明)

---

## 1. 项目概述

**非遗华裳**是一个以数字技术重现中国传统非物质文化遗产服饰的 3D 数字博物馆。项目以 Vue 3 + Three.js 前端配合 Go + Gin + SQLite 后端，展示非遗服饰的纹样、工艺与地理分布，并提供 3D 模型交互预览、智能推荐链、PDF 报告导出等核心功能。

### 核心功能

| 功能模块 | 路由 | 说明 |
|----------|------|------|
| 首页 | `/` | Hero 展示 + 每日一服 3D 预览 + 模块入口 |
| 服饰浏览 | `/explore` | 按朝代/民族/工艺筛选，关键词搜索，分页展示 |
| 服饰详情 | `/detail/:id` | 3D 模型预览 + 属性卡片 + 智能推荐链 + 收藏 |
| 收藏夹 | `/favorites` | 收藏/取消收藏管理（需登录） |
| 地理分布 | `/statistics` | 全国非遗服饰地理分布气泡图 + 工艺饼图 + 概览 |
| 导出报告 | `/report` | 基于收藏列表生成 PDF 报告下载 |
| 登录注册 | `/login` | 用户注册与 JWT 登录 |
| 后台管理 | `/admin` | 仪表盘 / 模型上传 / 服饰 CRUD（需管理员） |

---

## 2. 技术架构

### 2.1 整体架构图

```
Browser (localhost:5173)
   │
   ├─ /api/*  ──(Vite proxy)──▶ Gin API Server (localhost:3001)
   │                                      │
   └─ /* (Vue SPA)                        ├─ middleware 层
                                          │   ├─ Recovery (panic 恢复)
                                          │   ├─ CORS (跨域)
                                          │   └─ JWT + AdminOnly (认证鉴权)
                                          │
                                          ├─ handler 层 (HTTP 处理)
                                          │   ├─ AuthHandler
                                          │   ├─ GarmentHandler
                                          │   ├─ FavoriteHandler
                                          │   ├─ StatisticsHandler
                                          │   ├─ ReportHandler
                                          │   ├─ ModelHandler
                                          │   ├─ AdminGarmentHandler
                                          │   └─ AdminDashboardHandler
                                          │
                                          ├─ service 层 (业务逻辑)
                                          │   ├─ AuthService
                                          │   ├─ GarmentService
                                          │   ├─ DailyService
                                          │   ├─ ReportService
                                          │   ├─ ModelService
                                          │   └─ StatisticsService
                                          │
                                          ├─ model 层 (数据访问)
                                          │   ├─ UserStore
                                          │   ├─ GarmentStore
                                          │   ├─ FavoriteStore
                                          │   ├─ DailyPickStore
                                          │   └─ ModelStore
                                          │
                                          └─ SQLite (server/data/nonlegacy.db)
```

### 2.2 技术栈

| 层 | 技术 | 版本要求 |
|---|------|---------|
| **前端框架** | Vue 3 (Composition API + `<script setup>`) | ^3.4 |
| **前端语言** | TypeScript (strict) | ^5.4 |
| **构建工具** | Vite | ^5.4 |
| **路由** | Vue Router 4 | ^4.3 |
| **3D 渲染** | Three.js (原生 GLTF/GLB 加载) | ^0.160 |
| **样式** | TailwindCSS 3 (敦煌莫高窟色系) | ^3.4 |
| **图表** | ECharts 5 | ^5.5 |
| **HTTP 客户端** | Axios | ^1.6 |
| **后端框架** | Go + Gin | Go ≥ 1.21 |
| **数据库** | SQLite (mattn/go-sqlite3) | — |
| **认证** | JWT (golang-jwt/jwt/v5) + bcrypt | — |
| **PDF 生成** | gofpdf | — |
| **包管理器** | pnpm (前端) / Go modules (后端) | pnpm ≥ 8 |

---

## 3. 项目结构

```
nonlegacy-garment-3d/
├── client/                          # 前端 Vue SPA（使用 pnpm）
│   ├── public/
│   │   └── models/                  # 3D 模型源文件 (.glb)
│   │       ├── miao-skirt.glb       #   苗族百褶裙
│   │       └── dai-skirt.glb        #   傣族筒裙
│   ├── src/
│   │   ├── main.ts                  # 应用入口：createApp + router + Toast
│   │   ├── App.vue                  # 根组件：<router-view> + 全局动画
│   │   ├── router/
│   │   │   └── index.ts             # 路由定义 + 登录守卫 + admin 守卫
│   │   ├── components/
│   │   │   ├── Layout/              # Layout.vue / Navbar.vue / Footer.vue
│   │   │   ├── Common/              # Card.vue / Modal.vue / Skeleton.vue / EmptyState.vue / Toast.vue
│   │   │   └── 3D/                  # ModelViewer.vue / ModelLoader.ts
│   │   ├── pages/
│   │   │   ├── Home/HomePage.vue    # 首页：Hero + 每日一服 3D + 模块入口
│   │   │   ├── Explore/ExplorePage.vue  # 服饰浏览：搜索 + 筛选 + 分页网格
│   │   │   ├── Detail/DetailPage.vue    # 服饰详情：3D 模型 + 属性卡片 + 推荐链 + 收藏
│   │   │   ├── Favorites/FavoritesPage.vue  # 收藏夹：列表 + 移除 + 导出入口
│   │   │   ├── Statistics/StatisticsPage.vue # 统计：气泡图 + 饼图 + 概览
│   │   │   ├── Report/ReportPage.vue    # 导出报告：预览 + PDF 生成下载
│   │   │   ├── Auth/AuthPage.vue        # 登录注册页
│   │   │   └── Admin/                   # 后台管理页面
│   │   │       ├── AdminLayout.vue      #   侧边栏布局
│   │   │       ├── DashboardPage.vue    #   仪表盘统计卡片
│   │   │       ├── ModelsPage.vue       #   模型上传/列表/删除
│   │   │       ├── GarmentsPage.vue     #   服饰列表/编辑/删除
│   │   │       └── GarmentEditPage.vue  #   服饰创建/编辑表单
│   │   ├── composables/
│   │   │   └── index.ts             # useFetch<T> / useAuth / useDebounce<T> / useToast
│   │   ├── services/
│   │   │   └── api.ts               # Axios 封装 + 7 组 API 函数 + 拦截器
│   │   ├── types/                   # TypeScript 类型定义
│   │   │   ├── api.ts               #   ApiResponse<T> / PaginatedData<T> / GarmentQueryParams
│   │   │   ├── garment.ts           #   Garment / GarmentFormData / RecommendItem
│   │   │   ├── model.ts             #   ModelInfo / ModelListItem / ModelUploadResult
│   │   │   ├── user.ts              #   User / AuthResponse / LoginForm / RegisterForm
│   │   │   ├── statistics.ts        #   GeoDataItem / CraftFlowItem / OverviewStats / DashboardStats
│   │   │   ├── favorite.ts          #   FavoriteItem / FavoriteCheckResult
│   │   │   └── global.d.ts          #   Window.__showToast 类型扩展
│   │   ├── utils/
│   │   │   └── index.ts             # formatDate / formatFileSize / 省坐标 / 色系常量
│   │   └── styles/
│   │       └── index.css            # TailwindCSS + 自定义组件样式
│   ├── index.html
│   ├── vite.config.ts               # Vite 配置（代理 + @ 别名）
│   ├── tsconfig.json                # TypeScript 严格模式配置
│   ├── tailwind.config.js           # TailwindCSS 配置（莫高窟色系）
│   └── package.json
├── server/                          # Go + Gin 后端
│   ├── cmd/
│   │   └── api/
│   │       └── main.go              # 入口：配置加载 + DB 初始化 + 迁移 + -seed 种子 + 路由
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go            # 环境变量读取（PORT/DB_PATH/JWT_SECRET 等）
│   │   ├── database/
│   │   │   ├── sqlite.go            # SQLite 连接（WAL 模式 + 外键启用）
│   │   │   └── migrate.go           # 14 条迁移 SQL（5 表 + 9 索引）
│   │   ├── model/                   # 数据模型 + Store 模式数据访问
│   │   │   ├── response.go          #   Response / PaginatedData / SuccessResponse / ErrorResponse
│   │   │   ├── user.go              #   User + UserStore（CRUD + Count）
│   │   │   ├── garment.go           #   Garment + GarmentStore（List/GetByID/Create/Update/Delete + 推荐查询 + 统计）
│   │   │   ├── favorite.go          #   Favorite + FavoriteStore（Add/Remove/ListByUser/IsFavorited/Count）
│   │   │   ├── daily_pick.go        #   DailyPick + DailyPickStore（GetByDate/Create）
│   │   │   └── model_3d.go          #   Model3D + ModelStore（Create/GetByID/List/Update/Delete/Count）
│   │   ├── handler/                 # HTTP 处理层（Gin handlers）
│   │   │   ├── auth.go              #   Register / Login / Profile
│   │   │   ├── garment.go           #   List / GetByID / GetRecommendChain / GetDaily
│   │   │   ├── favorite.go          #   List / Add / Remove / Check
│   │   │   ├── statistics.go        #   GetGeoDistribution / GetCraftFlow / GetOverview
│   │   │   ├── report.go            #   Generate (PDF)
│   │   │   ├── model_handler.go     #   ListPublic / ServeFile / Admin CRUD
│   │   │   ├── admin_garment.go     #   Create / Update / Delete (garment)
│   │   │   └── admin_dashboard.go   #   GetDashboard (统计概览)
│   │   ├── middleware/
│   │   │   ├── cors.go              # CORS 中间件
│   │   │   ├── jwt.go               # JWT 生成/解析 + JWT() 守卫 + AdminOnly() 守卫
│   │   │   └── recovery.go          # Panic 恢复中间件
│   │   ├── service/                 # 业务逻辑层
│   │   │   ├── auth_service.go      #   注册/登录逻辑 + bcrypt 验证
│   │   │   ├── garment_service.go   #   推荐链算法（民族+工艺匹配）
│   │   │   ├── daily_service.go     #   每日一服（随机选 + 写入记录）
│   │   │   ├── report_service.go    #   PDF 报告生成（gofpdf）
│   │   │   ├── model_service.go     #   模型上传/删除（文件系统 + DB）
│   │   │   └── statistics_service.go #  地理/工艺/概览统计
│   │   └── router/
│   │       └── router.go            # Gin 路由注册（公开 + 管理两组）
│   ├── data/
│   │   ├── nonlegacy.db             # SQLite 数据库文件（.gitignore）
│   │   └── models/                  # 上传的 3D 模型文件（按 ID 分目录，.gitignore）
│   ├── reports/                     # PDF 报告输出目录（.gitignore）
│   ├── .env                         # 环境变量（不可提交）
│   ├── .env.example                 # 环境变量模板
│   ├── go.mod
│   └── go.sum
├── docs/                            # 项目文档
│   ├── API.md                       # API 接口文档（27 个端点）
│   └── PROJECT.md                   # 本文档
├── clothes/                         # 3D 模型源文件目录
├── CLAUDE.md                        # Claude Code 项目指令
└── README.md                        # 项目说明
```

---

## 4. 后端架构

### 4.1 分层设计

后端采用经典三层架构（Store → Service → Handler），各层职责清晰：

| 层 | 目录 | 职责 |
|---|------|------|
| **Handler** | `internal/handler/` | HTTP 请求解析、参数校验、调用 Service、JSON 序列化响应 |
| **Service** | `internal/service/` | 业务逻辑（认证、推荐算法、文件上传、PDF 生成） |
| **Model/Store** | `internal/model/` | 数据结构定义 + Store 模式数据访问（原生 SQL，无 ORM） |

### 4.2 Store 模式

每个数据库表对应一个 `XxxStore` 结构体，持有 `*sql.DB` 指针，方法直接写 SQL：

```go
type GarmentStore struct {
    db *sql.DB
}

func (s *GarmentStore) List(params GarmentQueryParams) ([]Garment, int, error) {
    // 动态构建 WHERE + 参数化查询 + COUNT 分页
}
```

**优点**：无 ORM 抽象层，SQL 可见可控，零运行时反射开销。

### 4.3 路由注册

路由分为两个组：

- **公开组** (`/api/*`)：无需认证或可选认证，19 个端点
- **管理组** (`/api/admin/*`)：JWT + AdminOnly 中间件保护，8 个端点

```go
// 公开路由
api := r.Group("/api")
{
    api.GET("/garments", garmentHandler.List)      // 无需认证
    api.POST("/favorites", ..., favoriteHandler.Add) // JWT 强制
    api.GET("/models/:id/file", modelHandler.ServeFile)
}

// 管理路由
admin := r.Group("/api/admin")
admin.Use(middleware.JWT(cfg.JWTSecret, true))
admin.Use(middleware.AdminOnly())
{
    admin.POST("/models", modelHandler.AdminUpload)
    admin.DELETE("/garments/:id", adminGarmentHandler.Delete)
}
```

### 4.4 中间件链

请求处理经过以下中间件：

1. **Recovery** — 全局 panic 恢复，返回 500 而非崩溃
2. **CORS** — 允许前端跨域（开发环境 `localhost:5173`）
3. **JWT** — 解析 `Authorization: Bearer <token>`，可选/强制两种模式
4. **AdminOnly** — 校验 `user_role == "admin"`，非管理员返回 403

### 4.5 配置管理

通过环境变量配置（`server/.env`），`config.Load()` 自动读取并设默认值：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `3001` | 后端服务端口 |
| `APP_ENV` | `development` | 运行环境（`production` 时 Gin 关闭 debug 日志） |
| `DB_PATH` | `./data/nonlegacy.db` | SQLite 数据库路径（相对 `server/`） |
| `JWT_SECRET` | `nonlegacy_dev_secret_key` | JWT 签名密钥（生产必须更换） |
| `JWT_EXPIRES_IN` | `168h` | JWT 有效期（7 天） |
| `CORS_ORIGIN` | `http://localhost:5173` | 允许的跨域来源 |
| `UPLOAD_DIR` | `./data/models` | 模型文件上传目录 |
| `MAX_UPLOAD_SIZE` | `52428800` | 上传文件大小限制（50MB） |

---

## 5. 前端架构

### 5.1 数据流

```
用户交互
    │
    ▼
Page Component (pages/*.vue)
    │  调用 API 层 + composables
    ▼
services/api.ts (Axios 实例)
    │  请求拦截器：附加 JWT Token
    │  响应拦截器：解包 .data + 401 处理
    ▼
Go Gin Server (:3001)
    │
    ▼
Page Component 更新状态 → Vue 响应式更新 DOM
```

### 5.2 API 层设计

`services/api.ts` 分为 7 组 API 函数：

| 组 | 变量名 | 端点数量 | 说明 |
|---|--------|---------|------|
| 服饰 | `garmentAPI` | 3 | list / getById / getRecommendChain / getDaily |
| 收藏 | `favoriteAPI` | 4 | list / add / remove / check |
| 统计 | `statisticsAPI` | 3 | getGeoDistribution / getCraftFlow / getOverview |
| 报告 | `reportAPI` | 1 | generate (blob 响应) |
| 认证 | `authAPI` | 3 | register / login / getProfile |
| 模型 | `modelAPI` | 2 | list / getFileUrl |
| 管理 | `admin*API` | 8 | 仪表盘 + 模型 CRUD + 服饰 CRUD |

所有函数返回类型为 `ApiResponse<T>`，拦截器自动脱去 axios 外层。

### 5.3 Composables

| Composable | 说明 |
|-----------|------|
| `useFetch<T>(fetcher, deps?)` | 通用异步获取，自动管理 loading/error/data 状态 |
| `useAuth()` | 登录状态管理：token + user 的 localStorage 读写，响应式 `isLoggedIn` |
| `useDebounce<T>(value, delay)` | 输入防抖，默认 300ms |
| `useToast()` | 全局通知管理，支持 `success` / `error` / `info` 三种类型 |

### 5.4 路由守卫

`router/index.ts` 的 `beforeEach` 守卫：

1. **已登录访问 `/login`** → 重定向到 `/`
2. **`requiresAuth` 且未登录** → 重定向到 `/login`
3. **`requiresAdmin` 且非 admin** → 重定向到 `/`

### 5.5 组件设计

**通用组件**（`components/Common/`）：

| 组件 | Props | 说明 |
|------|-------|------|
| `Card` | `to`, `title`, `subtitle`, `image?` | 可点击卡片，用于列表展示 |
| `Modal` | `visible`, `title`, `size?` | 弹窗容器（支持 sm/md/lg） |
| `Skeleton` | `type`, `count?` | 骨架屏加载态（card/detail 两种） |
| `EmptyState` | `icon`, `title`, `description`, `action?` | 空状态提示 + 可选操作按钮 |
| `Toast` | (通过 provide/inject) | 全局通知浮层，自动消失 |

**3D 组件**（`components/3D/`）：

| 组件 | 说明 |
|------|------|
| `ModelViewer.vue` | 封装 Three.js 场景：Scene + PerspectiveCamera + AmbientLight + DirectionalLight + OrbitControls |
| `ModelLoader.ts` | GLTFLoader 加载器，管理 load/progress/error 生命周期 |

---

## 6. 数据库设计

### 6.1 ER 关系图

```
┌──────────┐       ┌──────────────┐       ┌──────────┐
│   users  │       │   favorites  │       │ garments │
├──────────┤       ├──────────────┤       ├──────────┤
│ id (PK)  │──<    │ id (PK)      │    >──│ id (PK)  │──<
│ username │       │ user_id (FK) │       │ name     │   │
│ email    │       │ garment_id(FK)│      │ dynasty  │   │
│ role     │       │ created_at   │       │ ...      │   │
└──────────┘       └──────────────┘       │ model_id(FK)─┐
                                           └──────────┘   │
┌──────────────┐                                           │
│ daily_picks  │       ┌──────────┐                        │
├──────────────┤       │  models  │                        │
│ id (PK)      │       ├──────────┤                        │
│ garment_id(FK)──<    │ id (PK)  │<───────────────────────┘
│ pick_date    │       │ filename │
└──────────────┘       │ ...      │
                       └──────────┘
```

### 6.2 关键约束

| 约束 | 表 | 说明 |
|------|-----|------|
| `UNIQUE(user_id, garment_id)` | favorites | 一人一服饰仅可收藏一次 |
| `UNIQUE(pick_date)` | daily_picks | 每天只有一条推荐 |
| `ON DELETE CASCADE` | favorites, daily_picks | 删除服饰/用户时级联删除关联记录 |
| `ON DELETE SET NULL` | garments.model_id, models.garment_id | 删除关联时保留记录，解除关联 |

### 6.3 索引

数据库迁移共创建 9 个索引，覆盖所有高频查询字段：

- `garments`: dynasty, ethnic, craft, province, model_id
- `models`: garment_id
- `favorites`: user_id, garment_id
- `daily_picks`: pick_date

### 6.4 种子数据

运行 `go run ./cmd/api -seed` 插入：

- **2 个用户**：admin (admin/admin123, role=admin) + user (user/user123, role=user)
- **5 件服饰**：苗族百褶裙、苏绣旗袍、藏族氆氇袍、蒙古族长袍、傣族筒裙
- **2 个模型关联**：苗族百褶裙 ← 苗族.glb、傣族筒裙 ← 傣族.glb
- **1 条每日推荐**：当日日期关联苗族百褶裙

---

## 7. 后台管理面板

### 7.1 路由结构

```
/admin                   → AdminLayout.vue (侧边栏布局)
  /admin                 → DashboardPage.vue (仪表盘)
  /admin/models          → ModelsPage.vue (模型管理)
  /admin/garments        → GarmentsPage.vue (服饰管理)
  /admin/garments/new    → GarmentEditPage.vue (创建服饰)
  /admin/garments/:id/edit → GarmentEditPage.vue (编辑服饰)
```

### 7.2 功能详情

#### 仪表盘（DashboardPage）

- 4 个统计卡片：服饰总数（朱砂红）、模型总数（石绿）、用户总数（青金石蓝）、收藏总数（金色）
- 数据来源：`GET /api/admin/dashboard`

#### 模型管理（ModelsPage）

- **列表**：ID、文件名、大小、关联服饰、上传时间、操作（删除）
- **上传**：模态窗口，选择 `.glb/.gltf` 文件（前端校验格式+大小），可选关联服饰
- **删除**：二次确认弹窗，物理删除文件 + 数据库记录 + 解除服饰关联
- **分页**：每页 20 条，上一页/下一页导航

#### 服饰管理（GarmentsPage）

- **列表**：ID、名称、朝代、民族、工艺、省份、3D 模型关联状态、操作（编辑/删除）
- **删除**：二次确认，提示会级联删除收藏和每日推荐
- 点击「编辑」跳转到编辑页面

#### 服饰编辑（GarmentEditPage）

- **创建模式**：空表单，填写后 POST `/api/admin/garments`
- **编辑模式**：预填现有数据，PUT `/api/admin/garments/:id`
- **表单字段**：名称*、朝代*（下拉选择）、民族*（下拉选择）、工艺、材质、纹样、省份*、文化寓意（textarea）、描述（textarea）、图片 URL、关联 3D 模型（下拉选择）
- **朝代选项**：唐、宋、元、明、清、现代
- **民族选项**：汉族、苗族、藏族、蒙古族、傣族、维吾尔族、壮族、彝族、回族、满族、侗族、瑶族、白族、土家族、哈尼族、哈萨克族、黎族、京族、畲族、朝鲜族
- **模型下拉**：调用 `GET /api/models` 加载可用模型列表，显示原始文件名及已关联提示

---

## 8. 3D 模型渲染管线

### 8.1 模型文件获取流程

```
Garment 数据 (含 model_id)
       │
       ▼
computed modelPath:
  if model_id → /api/models/{model_id}/file    ← 新方式（优先）
  if model_url → model_url                     ← 旧方式（向后兼容）
       │
       ▼
GLTFLoader.load(modelPath)
       │
       ▼
Three.js Scene → ModelViewer 渲染
```

### 8.2 模型存储

- **存储路径**: `server/data/models/{id}/{timestamp}_{sanitized_name}.glb`
- **访问方式**: `GET /api/models/{id}/file` → Gin `c.File()` 流式输出二进制
- **缓存策略**: `Cache-Control: public, max-age=86400` (24 小时)
- **Range 支持**: `Accept-Ranges: bytes` — 支持断点续传

### 8.3 后向兼容

`garments` 表保留 `model_url` 字段，前端 `modelPath` computed 优先使用 `model_id`，降级使用 `model_url`：

```typescript
const modelPath = computed(() => {
  if (!garment.value) return ''
  if (garment.value.model_id) return `/api/models/${garment.value.model_id}/file`
  if (garment.value.model_url) return garment.value.model_url
  return ''
})
```

### 8.4 前端错误降级

当模型文件加载失败时，前端显示占位符（👗 emoji + 提示文字 "3D 服饰展示 / 暂无 3D 模型"），不阻塞页面。

---

## 9. 认证与鉴权

### 9.1 JWT 流程

```
注册/登录 → 后端生成 JWT（含 user_id + username + role，7 天有效）
              │
              ▼
         前端存入 localStorage.token + localStorage.user
              │
              ▼
         Axios 拦截器自动附加 Authorization: Bearer <token>
              │
              ▼
         Gin JWT 中间件解析 → 注入 c.Set("user_id", ...)
              │
              ▼
         Handler 通过 c.GetInt64("user_id") 获取当前用户
```

### 9.2 角色权限

| 角色 | 权限范围 |
|------|---------|
| `user` (默认) | 浏览服饰、收藏管理、生成报告 |
| `admin` | 上述全部 + 访问 `/admin/*` + 模型上传/删除 + 服饰 CRUD |

### 9.3 前端守卫逻辑

- **路由守卫** (`router.beforeEach`)：检查 `requiresAuth` / `requiresAdmin` meta 字段
- **Navbar**：通过 `user.role === 'admin'` 判断是否显示「管理」入口
- **401 拦截** (axios 响应拦截器)：清除 token → Toast 提示 → 延时跳转 `/login`

---

## 10. 配色方案

采用**敦煌莫高窟色系**，在 `tailwind.config.js` 和 `utils/index.ts` 中定义（两处需保持同步）：

| 色键 | 色值 | Tailwind 类前缀 | 应用场景 |
|------|------|----------------|----------|
| **朱砂红** (cinnabar) | `#C43A31` | `mogao-cinnabar` | 标题、主按钮、强调色、收藏心形、删除按钮 |
| **石绿** (malachite) | `#2E7D5A` | `mogao-malachite` | 辅助点缀、成功状态、关联标识 |
| **土黄** (ochre) | `#C89B5E` | `mogao-ochre` | 背景、边框、卡片、次按钮、分隔线 |
| **青金石蓝** (lapis) | `#2E4A7A` | `mogao-lapis` | 正文、深色背景、图表文字、侧边栏 |
| **铅白** (leadwhite) | `#F5F0E8` | `mogao-leadwhite` | 页面底色、浅色背景 |
| **金色** (gold) | `#B8862D` | `mogao-gold` | 高亮强调、特殊分隔线 |

每个色键包含 50-900 色阶（如 `mogao-cinnabar-500`）。

---

## 11. 开发指南

### 11.1 环境要求

- **Go** ≥ 1.21
- **Node.js** ≥ 18
- **pnpm** ≥ 8（前端必须使用 pnpm，禁止 npm/yarn）

### 11.2 本地开发启动

```bash
# 1. 后端
cd server
cp .env.example .env        # 首次：复制环境变量配置
go mod tidy                 # 安装 Go 依赖
go run ./cmd/api -seed      # 首次：插入种子数据并启动服务
go run ./cmd/api            # 后续：直接启动（无需 -seed）

# 2. 前端（另开终端）
cd client
pnpm install                # 安装前端依赖
pnpm run dev                # 启动 Vite 开发服务器
```

访问 `http://localhost:5173`

### 11.3 前端类型检查与构建

```bash
cd client
pnpm run type-check         # vue-tsc --noEmit 类型检查
pnpm run build              # 生产构建 → dist/
pnpm run preview            # 预览生产构建
```

### 11.4 后端测试

```bash
cd server
go test ./...               # 运行所有测试
```

### 11.5 开发注意事项

- **后端必须在 `server/` 目录下运行** — `.env` 中的相对路径依赖此前提
- **数据库和模型文件在 `.gitignore` 中** — `server/data/nonlegacy.db`、`server/data/models/`、`server/reports/` 不提交
- **`.env` 不可提交** — `.env.example` 是配置模板
- **ESLint 未安装** — lint 脚本暂时无法运行（已知问题）
- **Vite 代理自动转发 `/api/*` 到 `localhost:3001`** — 开发时无需手动配置 CORS

---

## 12. 部署说明

### 12.1 生产构建

```bash
# 前端构建
cd client
pnpm run build              # 输出到 client/dist/

# 后端编译
cd server
GOOS=linux GOARCH=amd64 go build -o api ./cmd/api
```

### 12.2 生产模式

在 `server/.env` 中设置：

```env
APP_ENV=production
CORS_ORIGIN=https://yourdomain.com
JWT_SECRET=<强随机密钥>
```

Gin 在 production 模式下关闭 debug 日志，且可通过 `r.Static()` 直接托管前端构建产物：

```go
// 在 router.go 中添加
r.Static("/assets", "./client/dist/assets")
r.StaticFile("/", "./client/dist/index.html")
```

### 12.3 数据库备份

SQLite 为单文件数据库，备份只需复制 `server/data/nonlegacy.db`：

```bash
cp server/data/nonlegacy.db server/data/nonlegacy.db.backup.$(date +%Y%m%d)
```

### 12.4 模型文件迁移

上传的模型文件存储在 `server/data/models/`，迁移时需将整个目录随数据库一起迁移。

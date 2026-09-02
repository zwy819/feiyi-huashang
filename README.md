# 非遗华裳 — 中国非物质文化遗产服饰 3D 数字博物馆

以数字技术重现千年织造智慧，探索中国非物质文化遗产服饰的纹样之美、工艺之精、文化之魂。

## 技术栈

| 层 | 技术 |
|---|---|
| **前端** | Vue 3 + Vue Router 4 + Vite + TypeScript |
| **3D 渲染** | Three.js（原生 GLTF/GLB 加载） |
| **样式** | TailwindCSS 3（敦煌莫高窟色系） |
| **图表** | ECharts 5 |
| **后端** | Go + Gin |
| **数据库** | SQLite |
| **认证** | JWT（golang-jwt + bcrypt） |
| **报告** | gofpdf（PDF 生成） |

## 项目结构

```
nonlegacy-garment-3d/
├── client/                          # 前端 Vue SPA（使用 pnpm）
│   ├── public/
│   │   └── models/                  # 3D 模型源文件 (.glb)
│   ├── src/
│   │   ├── main.ts                  # 应用入口
│   │   ├── App.vue                  # 根组件
│   │   ├── router/index.ts          # 路由配置（含 admin 路由）
│   │   ├── components/
│   │   │   ├── Layout/              # Layout / Navbar / Footer
│   │   │   ├── Common/              # Card / Modal / Skeleton / EmptyState / Toast
│   │   │   └── 3D/                  # ModelViewer / ModelLoader
│   │   ├── pages/
│   │   │   ├── Home/                # 首页
│   │   │   ├── Explore/             # 服饰浏览
│   │   │   ├── Detail/              # 服饰详情 + 3D
│   │   │   ├── Favorites/           # 收藏夹
│   │   │   ├── Statistics/          # 地理分布统计
│   │   │   ├── Report/              # PDF 报告
│   │   │   ├── Auth/                # 登录/注册
│   │   │   └── Admin/               # 后台管理面板
│   │   ├── composables/index.ts     # useFetch / useAuth / useDebounce / useToast
│   │   ├── services/api.ts          # Axios API 封装
│   │   ├── utils/index.ts           # 工具函数 + 常量
│   │   ├── types/                   # TypeScript 类型定义
│   │   └── styles/index.css         # 全局样式
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
├── server/                          # Go + Gin 后端
│   ├── cmd/api/main.go              # 入口 + 种子数据
│   ├── internal/
│   │   ├── config/                  # 环境变量配置
│   │   ├── database/                # SQLite 连接 + 迁移
│   │   ├── model/                   # 数据模型 + 数据访问
│   │   ├── handler/                 # HTTP 处理层
│   │   ├── middleware/              # CORS / JWT / Admin / Recovery
│   │   ├── service/                 # 业务逻辑层
│   │   └── router/                  # 路由注册
│   ├── data/
│   │   ├── nonlegacy.db             # SQLite 数据库
│   │   └── models/                  # 上传的 3D 模型文件
│   ├── reports/                     # 生成的 PDF 报告
│   ├── .env / .env.example
│   └── go.mod
├── docs/                            # 项目文档
│   ├── API.md                       # API 接口文档
│   └── PROJECT.md                   # 项目说明
├── clothes/                         # 3D 模型源文件
└── README.md
```

## 核心功能模块

| 模块 | 路由 | 功能描述 |
|---|---|---|
| **首页** | `/` | Hero 展示 + 每日一服 3D 预览 + 模块入口 |
| **服饰浏览** | `/explore` | 按朝代/民族/工艺筛选，关键词搜索，分页展示 |
| **服饰详情** | `/detail/:id` | 服饰详情 + 3D 模型预览 + 智能推荐链 + 收藏 |
| **收藏夹** | `/favorites` | 收藏/取消收藏管理（需登录） |
| **地理分布** | `/statistics` | 全国非遗服饰地理分布气泡图 + 工艺饼图 + 概览 |
| **导出报告** | `/report` | 基于收藏列表生成 PDF 报告下载（需登录） |
| **登录/注册** | `/login` | 用户注册与登录 |
| **后台管理** | `/admin` | 仪表盘 / 模型管理 / 服饰管理（需管理员） |

## 快速开始

### 环境要求

- Go >= 1.21
- Node.js >= 18
- pnpm >= 8

### 1. 启动后端

```bash
cd server

# 复制环境变量配置
cp .env.example .env

# 安装 Go 依赖
go mod tidy

# 插入种子数据并启动服务器（端口 3001）
go run ./cmd/api -seed
```

种子数据包含：
- 管理员账号：`admin` / `admin123`
- 普通用户：`user` / `user123`
- 5 件非遗服饰（苗族百褶裙、苏绣旗袍、藏族氆氇袍、蒙古族长袍、傣族筒裙）

### 2. 启动前端

```bash
cd client

# 安装依赖（必须使用 pnpm）
pnpm install

# 启动开发服务器（端口 5173）
pnpm run dev
```

### 3. 访问

浏览器打开 `http://localhost:5173`

## API 接口概览

所有接口响应格式：`{ "code": 0, "message": null, "data": {} }`

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/auth/register` | 用户注册 |
| `POST` | `/api/auth/login` | 用户登录 |
| `GET` | `/api/auth/profile` | 获取用户信息（JWT） |
| `GET` | `/api/garments` | 服饰列表（分页/筛选/搜索） |
| `GET` | `/api/garments/:id` | 服饰详情 |
| `GET` | `/api/garments/:id/recommend` | 智能推荐链 |
| `GET` | `/api/daily` | 每日一服 |
| `GET` | `/api/favorites` | 收藏列表（JWT） |
| `POST` | `/api/favorites` | 添加收藏（JWT） |
| `DELETE` | `/api/favorites/:garment_id` | 移除收藏（JWT） |
| `GET` | `/api/favorites/check/:garment_id` | 检查收藏状态（JWT） |
| `GET` | `/api/statistics/geo` | 地理分布数据 |
| `GET` | `/api/statistics/craft-flow` | 工艺流向数据 |
| `GET` | `/api/statistics/overview` | 概览统计 |
| `GET` | `/api/report/generate?ids=1,2,3` | 生成 PDF 报告 |
| `GET` | `/api/models` | 公开模型列表（分页） |
| `GET` | `/api/models/:id/file` | 下载模型文件（供 3D 渲染） |

### 管理接口（需要 JWT + admin 角色）

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/admin/dashboard` | 仪表盘统计 |
| `GET` | `/api/admin/models` | 模型列表 |
| `GET` | `/api/admin/models/:id` | 模型详情 |
| `POST` | `/api/admin/models` | 上传模型文件 |
| `PUT` | `/api/admin/models/:id` | 更新模型信息 |
| `DELETE` | `/api/admin/models/:id` | 删除模型 |
| `POST` | `/api/admin/garments` | 创建服饰 |
| `PUT` | `/api/admin/garments/:id` | 更新服饰 |
| `DELETE` | `/api/admin/garments/:id` | 删除服饰 |

详细接口文档见 [docs/API.md](docs/API.md)。

## 后台管理面板

管理员登录后可通过导航栏「管理」入口进入后台面板（`/admin`），功能包括：

- **仪表盘**：服饰/模型/用户/收藏统计概览
- **模型管理**：上传 GLB/GLTF 文件，关联服饰，删除模型
- **服饰管理**：创建/编辑/删除服饰信息，关联 3D 模型

管理员账号：`admin` / `admin123`

## 配色方案

采用**敦煌莫高窟色系**：

| 色名 | 色值 | 用途 |
|------|------|------|
| 朱砂红 | `#C43A31` | 标题、按钮主色、强调 |
| 石绿 | `#2E7D5A` | 辅助点缀、成功状态 |
| 土黄 | `#C89B5E` | 背景、边框、次按钮 |
| 青金石蓝 | `#2E4A7A` | 正文、深色背景 |
| 铅白 | `#F5F0E8` | 页面底色 |
| 金色 | `#B8862D` | 高亮强调、分隔线 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | 3001 | 后端服务端口 |
| `APP_ENV` | development | 运行环境 |
| `DB_PATH` | ./data/nonlegacy.db | SQLite 数据库路径 |
| `JWT_SECRET` | nonlegacy_dev_secret_key | JWT 签名密钥 |
| `JWT_EXPIRES_IN` | 168h | JWT 有效期 |
| `CORS_ORIGIN` | http://localhost:5173 | 允许的跨域来源 |
| `UPLOAD_DIR` | ./data/models | 模型文件上传目录 |
| `MAX_UPLOAD_SIZE` | 52428800 | 上传文件大小限制（50MB） |

## License

MIT

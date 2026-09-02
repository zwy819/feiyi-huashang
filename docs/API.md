# 非遗华裳 API 接口文档

> **Base URL**: `http://localhost:3001/api`
>
> 最后更新: 2026-07-06

---

## 目录

- [1. 通用约定](#1-通用约定)
- [2. 认证模块](#2-认证模块)
- [3. 服饰模块](#3-服饰模块)
- [4. 收藏模块](#4-收藏模块)
- [5. 统计模块](#5-统计模块)
- [6. 报告模块](#6-报告模块)
- [7. 模型模块（公开）](#7-模型模块公开)
- [8. 管理模块](#8-管理模块)
- [9. 健康检查](#9-健康检查)
- [10. 鉴权说明](#10-鉴权说明)
- [11. 数据库结构](#11-数据库结构)

---

## 1. 通用约定

### 1.1 响应格式

所有 JSON 响应统一使用以下三字段结构：

```json
{
  "code": 0,
  "message": null,
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | `number` | `0` = 成功；非 `0` = 业务错误（见下方业务码表） |
| `message` | `string \| null` | 成功时为 `null`（创建/更新/删除可返回提示文案），错误时为具体错误信息 |
| `data` | `any \| null` | 业务数据，无数据时返回 `null` |

### 1.2 业务码一览

| code | HTTP 状态码 | 含义 |
|------|------------|------|
| `0` | `2xx` | 成功 |
| `40001` | `400` | 参数缺失或格式错误 |
| `40002` | `400` | 参数值不合法（如格式错误、文件类型不支持、超大小限制） |
| `40101` | `401` | 未提供认证令牌（请求头缺少 `Authorization`） |
| `40102` | `401` | 认证令牌无效或已过期 |
| `40103` | `401` | 用户名或密码错误 |
| `40301` | `403` | 无权操作该资源（需要 `admin` 角色） |
| `40401` | `404` | 资源不存在（服饰/模型/用户/文件） |
| `40901` | `409` | 资源冲突（用户名/邮箱重复、重复收藏） |
| `50001` | `500` | 服务器内部错误 |

### 1.3 分页规范

- 页码从 **1** 开始（1-based）
- `limit` 默认值视接口而定（12 或 20），上限均为 100
- 响应始终包含 `total`、`page`、`limit`
- 前端通过 `Math.ceil(total / limit)` 计算总页数

### 1.4 日期时间格式

统一使用 ISO 字符串，本地时间，精确到秒：

```
2026-07-06 14:30:00
```

---

## 2. 认证模块

### 2.1 用户注册

```
POST /api/auth/register
Content-Type: application/json
```

**请求体：**

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "123456"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 用户名，全局唯一 |
| `email` | string | 是 | 邮箱，全局唯一 |
| `password` | string | 是 | 密码，最少 6 位 |

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "id": 3,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user"
  }
}
```

**返回字段说明：**

| 字段 | 说明 |
|------|------|
| `token` | JWT 令牌，有效期 7 天，后续请求需通过 `Authorization: Bearer <token>` 携带 |
| `role` | 角色：`admin` 或 `user` |

**错误码：**

| code | 场景 |
|------|------|
| `40001` | 参数缺失或格式错误 |
| `40002` | 密码长度不能少于 6 位 |
| `40901` | 用户名已存在 / 邮箱已被注册 |

---

### 2.2 用户登录

```
POST /api/auth/login
Content-Type: application/json
```

**请求体：**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "id": 1,
    "username": "admin",
    "email": "admin@nonlegacy.cn",
    "role": "admin"
  }
}
```

**错误码：**

| code | 场景 |
|------|------|
| `40001` | 参数缺失或格式错误 |
| `40103` | 用户名或密码错误 |

---

### 2.3 获取用户信息

```
GET /api/auth/profile
Authorization: Bearer <token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@nonlegacy.cn",
    "avatar_url": "",
    "role": "admin",
    "created_at": "2026-07-06 10:00:00",
    "updated_at": "2026-07-06 10:00:00"
  }
}
```

**错误码：**

| code | 场景 |
|------|------|
| `40101` | 未提供认证令牌 |
| `40102` | 认证令牌无效或已过期 |
| `40401` | 用户不存在 |

---

## 3. 服饰模块

### 3.1 服饰列表

```
GET /api/garments
```

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | number | `1` | 页码（1-based） |
| `limit` | number | `12` | 每页条数（上限 100） |
| `search` | string | — | 关键词搜索（匹配名称/描述/工艺/民族） |
| `dynasty` | string | — | 朝代精确筛选 |
| `ethnic` | string | — | 民族精确筛选 |
| `craft` | string | — | 工艺精确筛选 |

**示例请求：**

```
GET /api/garments?page=1&limit=12&search=苗&dynasty=清&ethnic=苗族
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "苗族百褶裙",
        "dynasty": "清",
        "ethnic": "苗族",
        "craft": "蜡染+刺绣",
        "material": "棉麻",
        "pattern": "几何纹+花卉纹",
        "cultural_meaning": "象征苗族人民对自然和生命的崇敬...",
        "province": "贵州",
        "description": "苗族百褶裙是苗族女性传统服饰中...",
        "image_url": "",
        "model_url": "",
        "model_id": 1,
        "created_at": "2026-07-06 10:00:00",
        "updated_at": "2026-07-06 10:00:00"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 12
  }
}
```

---

### 3.2 服饰详情

```
GET /api/garments/:id
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "id": 1,
    "name": "苗族百褶裙",
    "dynasty": "清",
    "ethnic": "苗族",
    "craft": "蜡染+刺绣",
    "material": "棉麻",
    "pattern": "几何纹+花卉纹",
    "cultural_meaning": "象征苗族人民对自然和生命的崇敬...",
    "province": "贵州",
    "description": "苗族百褶裙是苗族女性传统服饰中最具代表性的下装...",
    "image_url": "",
    "model_url": "",
    "model_id": 1,
    "created_at": "2026-07-06 10:00:00",
    "updated_at": "2026-07-06 10:00:00"
  }
}
```

**服饰字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | number | 服饰 ID |
| `name` | string | 服饰名称 |
| `dynasty` | string | 朝代（展示时拼接"代"，如 "清代"） |
| `ethnic` | string | 民族 |
| `craft` | string | 工艺（可含多种，用 "+" 分隔） |
| `material` | string | 材质 |
| `pattern` | string | 纹样描述 |
| `cultural_meaning` | string | 文化寓意 |
| `province` | string | 省份 |
| `description` | string | 详细描述 |
| `image_url` | string | 静态预览图 URL |
| `model_url` | string | **[向后兼容]** 旧版模型路径 |
| `model_id` | number \| null | 关联的 3D 模型 ID，前端通过 `/api/models/{model_id}/file` 加载 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

**错误码：**

| code | 场景 |
|------|------|
| `40002` | `id` 参数值不合法 |
| `40401` | 服饰不存在 |

---

### 3.3 智能推荐链

```
GET /api/garments/:id/recommend
```

**推荐算法：**

基于「民族 + 工艺」匹配度计算相似性：
- 民族和工艺都匹配 → `similarity = 2`
- 仅民族或工艺匹配 → `similarity = 1`
- 都不匹配 → `similarity = 0`

按 `similarity` 降序排列，返回最多 4 件（排除当前服饰自身）。

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "items": [
      { "id": 3, "name": "藏族氆氇袍", "similarity": 1 },
      { "id": 5, "name": "傣族筒裙", "similarity": 1 }
    ]
  }
}
```

---

### 3.4 每日一服

```
GET /api/daily
```

**业务规则：**

- 按日期唯一，每天仅推送一件
- 先查 `daily_picks` 表今日是否有记录 → 有则返回已推送的
- 无则通过 `ORDER BY RANDOM()` 随机选一件，记录到 `daily_picks`，下次请求直接返回

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "id": 2,
    "name": "苏绣旗袍",
    "dynasty": "现代",
    "ethnic": "汉族",
    "craft": "苏绣",
    "material": "丝绸",
    "pattern": "花鸟纹",
    "cultural_meaning": "体现江南水乡的柔美与精致...",
    "province": "江苏",
    "description": "...",
    "image_url": "",
    "model_url": "",
    "model_id": 2,
    "created_at": "2026-07-06 10:00:00",
    "updated_at": "2026-07-06 10:00:00"
  }
}
```

---

## 4. 收藏模块

> 全部需要 JWT 认证（`Authorization: Bearer <token>`）

### 4.1 收藏列表

```
GET /api/favorites
Authorization: Bearer <token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "favorites": [
      {
        "id": 1,
        "user_id": 1,
        "garment_id": 1,
        "created_at": "2026-07-06 12:00:00",
        "garment_name": "苗族百褶裙",
        "garment_dynasty": "清",
        "garment_ethnic": "苗族",
        "garment_craft": "蜡染+刺绣",
        "garment_province": "贵州",
        "garment_image_url": ""
      }
    ]
  }
}
```

> `favorites` 是 JOIN 查询结果，包含 `garment_*` 前缀字段供前端直接展示，无需额外请求。

---

### 4.2 添加收藏

```
POST /api/favorites
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体：**

```json
{
  "garment_id": 1
}
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": "收藏成功",
  "data": null
}
```

**错误码：**

| code | 场景 |
|------|------|
| `40001` | 参数缺失 |
| `40002` | `garment_id` 不合法（≤ 0） |
| `40901` | 已收藏该服饰（UNIQUE 约束冲突） |

---

### 4.3 移除收藏

```
DELETE /api/favorites/:garment_id
Authorization: Bearer <token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": "已取消收藏",
  "data": null
}
```

---

### 4.4 检查收藏状态

```
GET /api/favorites/check/:garment_id
Authorization: Bearer <token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "isFavorited": true
  }
}
```

---

## 5. 统计模块

### 5.1 地理分布

```
GET /api/statistics/geo
```

按省份聚合服饰数量，供前端 ECharts 气泡图使用。

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": [
    { "province": "贵州", "count": 1 },
    { "province": "江苏", "count": 1 },
    { "province": "西藏", "count": 1 },
    { "province": "内蒙古", "count": 1 },
    { "province": "云南", "count": 1 }
  ]
}
```

---

### 5.2 工艺流向

```
GET /api/statistics/craft-flow
```

按工艺聚合服饰数量，供前端饼图/流程图使用。

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": [
    { "craft": "蜡染+刺绣", "count": 1 },
    { "craft": "苏绣", "count": 1 },
    { "craft": "氆氇织造", "count": 1 },
    { "craft": "镶边刺绣", "count": 1 },
    { "craft": "织锦", "count": 1 }
  ]
}
```

---

### 5.3 概览统计

```
GET /api/statistics/overview
```

返回服饰总量、民族数、工艺数、省份数，供仪表盘/首页展示。

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "totalGarments": 5,
    "totalEthnics": 5,
    "totalCrafts": 5,
    "totalProvinces": 5
  }
}
```

---

## 6. 报告模块

### 6.1 生成 PDF 报告

```
GET /api/report/generate?ids=1,2,3
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ids` | string | 是 | 逗号分隔的服饰 ID 列表（如 `1,2,3`） |

**成功响应 (200)：**

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="nonlegacy-report-3.pdf"
```

响应体为二进制 PDF 流。前端通过 `axios({ responseType: 'blob' })` 获取后创建 Blob URL 触发下载。

**PDF 内容：**
- 封面（标题 "非遗华裳" + 服饰数量）
- 每件服饰独立一页（名称、朝代、民族、工艺、材质、纹样、省份、文化寓意、描述）

**错误码：**

| code | 场景 |
|------|------|
| `40002` | 未提供有效的服饰 ID 列表 |
| `50001` | PDF 生成失败 / 未找到指定的服饰 |

---

## 7. 模型模块（公开）

### 7.1 模型列表

```
GET /api/models
```

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | number | `1` | 页码（1-based） |
| `limit` | number | `20` | 每页条数（上限 100） |
| `garment_id` | number | — | 按关联服饰筛选 |
| `search` | string | — | 按原始文件名搜索 |

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "items": [
      {
        "id": 1,
        "filename": "苗族.glb",
        "original_name": "苗族.glb",
        "file_size": 3125036,
        "mime_type": "model/gltf-binary",
        "garment_id": 1,
        "garment_name": "苗族百褶裙",
        "created_at": "2026-07-06 10:00:00",
        "updated_at": "2026-07-06 10:00:00"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 20
  }
}
```

> ⚠️ **安全提示**：公开接口**不返回** `file_path` 字段，防止暴露服务器文件路径。

---

### 7.2 下载模型文件

```
GET /api/models/:id/file
```

供前端 Three.js GLTFLoader 加载 3D 模型。

**成功响应 (200)：**

```
Content-Type: model/gltf-binary
Content-Disposition: inline; filename="苗族.glb"
Accept-Ranges: bytes
Cache-Control: public, max-age=86400
```

响应体为二进制 GLB/GLTF 数据。支持 HTTP Range 请求，缓存 24 小时。

**前端使用方式：**

```typescript
// 优先使用 model_id，向后兼容 model_url
const modelPath = computed(() => {
  if (garment.value.model_id)
    return `/api/models/${garment.value.model_id}/file`
  if (garment.value.model_url)
    return garment.value.model_url
  return ''
})
```

**错误码：**

| code | 场景 |
|------|------|
| `40002` | `id` 参数值不合法 |
| `40401` | 模型文件不存在（数据库记录或物理文件缺失） |

---

## 8. 管理模块

> 全部需要 JWT 认证 + `admin` 角色。普通用户访问返回 `40301`。
>
> **测试管理员**：`admin` / `admin123`

### 8.1 仪表盘统计

```
GET /api/admin/dashboard
Authorization: Bearer <admin_token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "total_garments": 5,
    "total_models": 5,
    "total_users": 2,
    "total_favorites": 3,
    "recent_uploads": [
      {
        "id": 5,
        "filename": "傣族.glb",
        "original_name": "傣族.glb",
        "file_size": 2567890,
        "mime_type": "model/gltf-binary",
        "file_path": "data/models/5/傣族.glb",
        "garment_id": 5,
        "garment_name": "傣族筒裙",
        "created_at": "2026-07-06 10:00:00",
        "updated_at": "2026-07-06 10:00:00"
      }
    ]
  }
}
```

| 字段 | 说明 |
|------|------|
| `total_garments` | 服饰总数 |
| `total_models` | 模型文件总数 |
| `total_users` | 注册用户总数 |
| `total_favorites` | 收藏记录总数 |
| `recent_uploads` | 最近上传的 5 个模型（含 `file_path`） |

---

### 8.2 模型列表（管理员）

```
GET /api/admin/models?page=1&limit=20
Authorization: Bearer <admin_token>
```

管理接口**包含 `file_path` 字段**，供后台查看/管理文件存储路径。

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": {
    "items": [
      {
        "id": 1,
        "filename": "苗族.glb",
        "original_name": "苗族.glb",
        "file_size": 3125036,
        "mime_type": "model/gltf-binary",
        "file_path": "data/models/1/苗族.glb",
        "garment_id": 1,
        "garment_name": "苗族百褶裙",
        "created_at": "2026-07-06 10:00:00",
        "updated_at": "2026-07-06 10:00:00"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 20
  }
}
```

---

### 8.3 模型详情

```
GET /api/admin/models/:id
Authorization: Bearer <admin_token>
```

返回单个模型的完整信息（含 `file_path`）。

---

### 8.4 上传模型文件

```
POST /api/admin/models
Authorization: Bearer <admin_token>
Content-Type: multipart/form-data
```

**表单字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | File | **是** | 模型文件，仅支持 `.glb` / `.gltf`，最大 50MB |
| `garment_id` | number | 否 | 关联的服饰 ID |

**上传流程：**

1. 验证文件扩展名（`.glb` / `.gltf`）
2. 验证文件大小（≤ 50MB）
3. 验证 `garment_id`（如提供，需存在对应服饰）
4. 创建 `models` 数据库记录（获取自增 ID）
5. 创建存储目录 `{UPLOAD_DIR}/{id}/`
6. 写入文件（`{timestamp}_{sanitized_name}`）
7. 更新数据库记录（写入 `filename`、`file_path`）
8. 如关联了服饰，同步更新 `garments.model_id`

**成功响应 (201)：**

```json
{
  "code": 0,
  "message": "模型上传成功",
  "data": {
    "id": 6,
    "filename": "1710123456_new-model.glb",
    "original_name": "new-model.glb",
    "file_size": 1024000,
    "mime_type": "model/gltf-binary",
    "file_path": "data/models/6/1710123456_new-model.glb",
    "garment_id": 1,
    "garment_name": "苗族百褶裙",
    "created_at": "2026-07-06 14:30:00",
    "updated_at": "2026-07-06 14:30:00"
  }
}
```

**错误码：**

| code | 场景 |
|------|------|
| `40001` | 未提供文件（`file` 字段为空） |
| `40002` | 不支持的文件格式 / 文件大小超过限制（50MB） / 关联的服饰不存在 |

---

### 8.5 更新模型信息

```
PUT /api/admin/models/:id
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体（部分更新）：**

```json
{
  "original_name": "renamed-model.glb",
  "garment_id": 1
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `original_name` | string | 新的显示名称 |
| `garment_id` | number | 关联的服饰 ID（`null` 表示解除关联） |

**副作用：** 如更新了 `garment_id`，会同步更新 `garments` 表的 `model_id`。

---

### 8.6 删除模型

```
DELETE /api/admin/models/:id
Authorization: Bearer <admin_token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": "模型已删除",
  "data": null
}
```

**副作用：**
1. 如关联了服饰，解除 `garments.model_id`（设为 `NULL`）
2. 删除数据库 `models` 记录
3. 删除服务器上的模型文件及目录（`{UPLOAD_DIR}/{id}/`）

> ⚠️ **不可撤销**：文件物理删除后无法恢复。

---

### 8.7 创建服饰

```
POST /api/admin/garments
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体：**

```json
{
  "name": "测试服饰",
  "dynasty": "清",
  "ethnic": "苗族",
  "craft": "蜡染",
  "material": "棉麻",
  "pattern": "几何纹",
  "cultural_meaning": "测试文化寓意",
  "province": "贵州",
  "description": "测试描述",
  "image_url": "",
  "model_id": null
}
```

**必填字段：** `name`、`dynasty`、`ethnic`、`province`

**成功响应 (201)：**

```json
{
  "code": 0,
  "message": "服饰创建成功",
  "data": { "... 创建的完整服饰对象（含生成的 id） ..." }
}
```

---

### 8.8 更新服饰

```
PUT /api/admin/garments/:id
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**请求体（部分更新，只发送需要修改的字段）：**

```json
{
  "name": "更新后的名称",
  "craft": "新工艺",
  "model_id": 2
}
```

可更新字段：`name`、`dynasty`、`ethnic`、`craft`、`material`、`pattern`、`cultural_meaning`、`province`、`description`、`image_url`、`model_url`、`model_id`

---

### 8.9 删除服饰

```
DELETE /api/admin/garments/:id
Authorization: Bearer <admin_token>
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": "服饰已删除",
  "data": null
}
```

**副作用：**
1. 删除 `garments` 记录
2. 级联删除关联的 `favorites` 和 `daily_picks`（CASCADE）
3. 关联的 `models.garment_id` 设为 `NULL`（ON DELETE SET NULL）

---

## 9. 健康检查

```
GET /api/health
```

**成功响应 (200)：**

```json
{
  "code": 0,
  "message": null,
  "data": "ok"
}
```

---

## 10. 鉴权说明

### 10.1 JWT Token 流程

1. 登录/注册成功 → 返回 JWT Token（有效期 168h / 7 天）
2. 客户端存入 `localStorage.token`
3. 后续请求通过 `Authorization: Bearer <token>` 请求头发送
4. Token 包含 `user_id`、`username`、`role` 三个声明
5. 过期后需重新登录

### 10.2 路由鉴权矩阵

| 路由 | 鉴权方式 | 未登录/无权限行为 |
|------|----------|-------------------|
| `POST /api/auth/register` | 无 | — |
| `POST /api/auth/login` | 无 | — |
| `GET /api/auth/profile` | JWT | 返回 401 |
| `GET /api/garments` 系列 | 无（可选 JWT） | — |
| `GET /api/daily` | 无 | — |
| `GET/POST/DELETE /api/favorites` | JWT 强制 | 返回 401 |
| `GET /api/statistics/*` | 无 | — |
| `GET /api/report/generate` | 无 | — |
| `GET /api/models` 公开 | 无 | — |
| `GET /api/admin/*` | JWT + admin 角色 | 401（无 token）/ 403（非 admin） |

### 10.3 测试账号

| 用户名 | 密码 | 角色 | 权限 |
|--------|------|------|------|
| `admin` | `admin123` | admin | 全部权限 + 后台管理 |
| `user` | `user123` | user | 收藏、报告生成 |

---

## 11. 数据库结构

### 11.1 ER 关系

```
users (1) ──< (N) favorites (N) >── (1) garments
                    │                       │
                    │                       ├──< models (N:1, SET NULL on delete)
                    │                       │
                    └── daily_picks (N:1, CASCADE on delete)
```

### 11.2 users 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | 用户 ID |
| `username` | TEXT | UNIQUE NOT NULL | 用户名 |
| `email` | TEXT | UNIQUE NOT NULL | 邮箱 |
| `password_hash` | TEXT | NOT NULL | bcrypt 密码哈希 |
| `avatar_url` | TEXT | DEFAULT '' | 头像 URL |
| `role` | TEXT | NOT NULL DEFAULT 'user' | 角色：`admin` / `user` |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

### 11.3 garments 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | 服饰 ID |
| `name` | TEXT | NOT NULL | 服饰名称 |
| `dynasty` | TEXT | DEFAULT '' | 朝代 |
| `ethnic` | TEXT | DEFAULT '' | 民族 |
| `craft` | TEXT | DEFAULT '' | 工艺 |
| `material` | TEXT | DEFAULT '' | 材质 |
| `pattern` | TEXT | DEFAULT '' | 纹样 |
| `cultural_meaning` | TEXT | DEFAULT '' | 文化寓意 |
| `province` | TEXT | DEFAULT '' | 省份 |
| `description` | TEXT | DEFAULT '' | 描述 |
| `image_url` | TEXT | DEFAULT '' | 预览图 URL |
| `model_url` | TEXT | DEFAULT '' | **[向后兼容]** 旧版模型路径 |
| `model_id` | INTEGER | FK→models(id) ON DELETE SET NULL | 关联 3D 模型 |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：** `dynasty`、`ethnic`、`craft`、`province`、`model_id`

### 11.4 models 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | 模型 ID |
| `filename` | TEXT | NOT NULL | 存储文件名（`{timestamp}_{sanitized}`） |
| `original_name` | TEXT | NOT NULL | 原始上传文件名 |
| `file_size` | INTEGER | NOT NULL DEFAULT 0 | 文件大小（字节） |
| `mime_type` | TEXT | NOT NULL DEFAULT 'model/gltf-binary' | MIME 类型 |
| `file_path` | TEXT | NOT NULL DEFAULT '' | 文件相对路径 |
| `garment_id` | INTEGER | FK→garments(id) ON DELETE SET NULL | 关联服饰 |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引：** `garment_id`

**存储路径：** `{UPLOAD_DIR}/{id}/{timestamp}_{sanitized_name}.glb`

### 11.5 favorites 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | 记录 ID |
| `user_id` | INTEGER | FK→users(id) ON DELETE CASCADE | 用户 ID |
| `garment_id` | INTEGER | FK→garments(id) ON DELETE CASCADE | 服饰 ID |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 收藏时间 |

**约束：** `UNIQUE(user_id, garment_id)` — 同一用户对同一服饰仅可收藏一次

### 11.6 daily_picks 表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | 记录 ID |
| `garment_id` | INTEGER | FK→garments(id) ON DELETE CASCADE | 服饰 ID |
| `pick_date` | DATE | UNIQUE NOT NULL | 推荐日期 |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

---

## 附录 A：接口速查表

| # | 方法 | 路径 | 说明 | 鉴权 | 分页 |
|---|------|------|------|------|------|
| 1 | `POST` | `/api/auth/register` | 用户注册 | — | — |
| 2 | `POST` | `/api/auth/login` | 用户登录 | — | — |
| 3 | `GET` | `/api/auth/profile` | 获取用户信息 | JWT | — |
| 4 | `GET` | `/api/garments` | 服饰列表 | — | ✅ |
| 5 | `GET` | `/api/garments/:id` | 服饰详情 | — | — |
| 6 | `GET` | `/api/garments/:id/recommend` | 智能推荐链 | — | — |
| 7 | `GET` | `/api/daily` | 每日一服 | — | — |
| 8 | `GET` | `/api/favorites` | 收藏列表 | JWT | — |
| 9 | `POST` | `/api/favorites` | 添加收藏 | JWT | — |
| 10 | `DELETE` | `/api/favorites/:garment_id` | 移除收藏 | JWT | — |
| 11 | `GET` | `/api/favorites/check/:garment_id` | 检查收藏状态 | JWT | — |
| 12 | `GET` | `/api/statistics/geo` | 地理分布 | — | — |
| 13 | `GET` | `/api/statistics/craft-flow` | 工艺流向 | — | — |
| 14 | `GET` | `/api/statistics/overview` | 概览统计 | — | — |
| 15 | `GET` | `/api/report/generate` | 生成 PDF 报告 | — | — |
| 16 | `GET` | `/api/models` | 模型列表（公开） | — | ✅ |
| 17 | `GET` | `/api/models/:id/file` | 下载模型文件 | — | — |
| 18 | `GET` | `/api/admin/dashboard` | 仪表盘统计 | JWT+admin | — |
| 19 | `GET` | `/api/admin/models` | 模型列表（管理） | JWT+admin | ✅ |
| 20 | `GET` | `/api/admin/models/:id` | 模型详情 | JWT+admin | — |
| 21 | `POST` | `/api/admin/models` | 上传模型 | JWT+admin | — |
| 22 | `PUT` | `/api/admin/models/:id` | 更新模型信息 | JWT+admin | — |
| 23 | `DELETE` | `/api/admin/models/:id` | 删除模型 | JWT+admin | — |
| 24 | `POST` | `/api/admin/garments` | 创建服饰 | JWT+admin | — |
| 25 | `PUT` | `/api/admin/garments/:id` | 更新服饰 | JWT+admin | — |
| 26 | `DELETE` | `/api/admin/garments/:id` | 删除服饰 | JWT+admin | — |
| 27 | `GET` | `/api/health` | 健康检查 | — | — |

## 附录 B：种子数据一览

| ID | 名称 | 朝代 | 民族 | 工艺 | 省份 | 3D 模型 |
|----|------|------|------|------|------|---------|
| 1 | 苗族百褶裙 | 清 | 苗族 | 蜡染+刺绣 | 贵州 | ✅ 苗族.glb |
| 2 | 苏绣旗袍 | 现代 | 汉族 | 苏绣 | 江苏 | — |
| 3 | 藏族氆氇袍 | 清 | 藏族 | 氆氇织造 | 西藏 | — |
| 4 | 蒙古族长袍 | 元 | 蒙古族 | 镶边刺绣 | 内蒙古 | — |
| 5 | 傣族筒裙 | 清 | 傣族 | 织锦 | 云南 | ✅ 傣族.glb |

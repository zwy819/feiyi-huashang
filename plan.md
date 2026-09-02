# 地理分布中国地图 + 工艺流向桑基图 实施计划

## 需求概述

1. **地理分布图**：将现有的 ECharts 散点图改造为真实中国地图底图 + 数据叠加
2. **工艺流向图**：将现有占位符替换为桑基图，展示 朝代 → 民族 → 工艺 三层流向

## 实施清单

### 1. 后端 — 新增桑基图数据接口

**文件**: `server/internal/service/statistics_service.go`

新增数据结构：
```go
// SankeyNode 桑基图节点
type SankeyNode struct {
    Name string `json:"name"`
}

// SankeyLink 桑基图连接
type SankeyLink struct {
    Source string `json:"source"`
    Target string `json:"target"`
    Value  int    `json:"value"`
}

// SankeyData 桑基图完整数据
type SankeyData struct {
    Nodes []SankeyNode `json:"nodes"`
    Links []SankeyLink `json:"links"`
}
```

新增方法 `GetSankeyData() (*SankeyData, error)`，逻辑：
1. 从 garments 表查询所有记录，获取 dynasty / ethnic / craft 三个字段
2. 对 craft 字段做**拆词处理**：将 "蜡染+刺绣" 拆为 ["蜡染", "刺绣"] 两条独立记录
3. 统计朝代→民族的流向（source=dynasty, target=ethnic, value=count）
4. 统计民族→工艺的流向（source=ethnic, target=craft, value=count）
5. 收集所有不重复的节点名，生成 nodes 数组
6. 去重合并 links，返回 SankeyData

**文件**: `server/internal/handler/statistics.go`

新增方法 `GetSankey(c *gin.Context)`，调用 service 返回数据。

**文件**: `server/internal/router/router.go`

在 statistics 路由组中新增：
```go
stats.GET("/sankey", statsHandler.GetSankey)
```

---

### 2. 前端 — 类型定义扩展

**文件**: `client/src/types/statistics.ts`

新增类型：
```typescript
export interface SankeyNode {
  name: string
}

export interface SankeyLink {
  source: string
  target: string
  value: number
}

export interface SankeyData {
  nodes: SankeyNode[]
  links: SankeyLink[]
}
```

---

### 3. 前端 — API 层新增接口

**文件**: `client/src/services/api.ts`

在 statisticsAPI 中新增：
```typescript
getSankey: () =>
  api.get<unknown, ApiResponse<SankeyData>>('/statistics/sankey'),
```

---

### 4. 获取中国 GeoJSON 数据

**文件**: `client/public/geojson/china.json`

从 DataV.GeoAtlas 或 ECharts 官方资源获取省级中国地图 GeoJSON（约 500KB-1MB），存放于 `client/public/geojson/china.json`。

Vite 会将其作为静态文件服务，前端通过 `/geojson/china.json` 加载。

**备选降级 CDN 地址**（方案 A 兜底）：
- `https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json`（DataV 阿里云，含完整省界）

---

### 5. 前端 — StatisticsPage.vue 全面改造

**文件**: `client/src/pages/Statistics/StatisticsPage.vue`

#### 5.1 ECharts 模块注册更新

当前注册：
```typescript
import { PieChart, ScatterChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
```

新增注册：
```typescript
import { PieChart, MapChart, SankeyChart } from 'echarts/charts'
import {
  GridComponent, TooltipComponent, LegendComponent,
  GeoComponent, VisualMapComponent,
} from 'echarts/components'
```

注册调用更新：
```typescript
echarts.use([
  PieChart, MapChart, SankeyChart,
  GridComponent, TooltipComponent, LegendComponent,
  GeoComponent, VisualMapComponent,
  CanvasRenderer,
])
```

#### 5.2 中国地图实现逻辑

新增 Ref：
```typescript
const chinaGeoJSON: Ref<any> = ref(null)
const sankeyData: Ref<SankeyData | null> = ref(null)
```

新增函数 `loadChinaGeoJSON()`：
```
1. 先尝试 fetch('/geojson/china.json') 加载本地 GeoJSON
2. 如果本地加载失败（404/超时），降级 fetch CDN 地址
3. 如果 CDN 也失败，降级为当前的散点图模式（fallback）
4. 将获取到的 GeoJSON 通过 echarts.registerMap('china', geoJSON) 注册
```

新增函数 `buildGeoOption()` 重写：
```
- 如果 chinaGeoJSON 加载成功：
  - series[0]: type: 'map', map: 'china', 底图渲染
    - 使用莫高窟色系：背景 leadwhite，省界 ochre
    - visualMap 组件：按省份服饰数量分级设色（浅→深 cinnabar）
    - 省份标签显示名称
  - series[1]: type: 'effectScatter'（涟漪气泡）
    - 叠加服饰数量散点，位置使用 provinceCoordinates 映射到地图坐标
    - symbolSize 按 sqrt(count) 计算
- 否则：降级为当前 scatter 实现（保留现有代码）
```

#### 5.3 桑基图实现逻辑

新增函数 `buildSankeyOption()`：
```
- 数据来自 sankeyData.value
- series[0]: type: 'sankey'
  - layout: 'none'（手动控制节点位置）
  - layoutIterations: 0
  - nodeAlign: 'left'
  - 三层节点手动分列：
    - 朝代节点 → 左侧列（x: 0）
    - 民族节点 → 中间列（x: 0.5）
    - 工艺节点 → 右侧列（x: 1.0）
  - 每层的 y 按 value 降序排列
  - 颜色方案：
    - 朝代层使用 lapis 蓝色系
    - 民族层使用 cinnabar 暖色系
    - 工艺层使用 malachite 绿色系
  - 标签显示节点名称 + 数值
  - tooltip 显示 source → target: value 件
  - emphasis 高亮关联连线
```

新建 ref `sankeyChartRef` 和对应的 ECharts 实例 `sankeyChart`。

#### 5.4 模板更新

- 地理分布区块标题不变
- 气泡图/地图 chart ref 保持不变（`geoChartRef`）
- 工艺类型分布饼图保持不变
- **工艺流向图占位** → 替换为真实的桑基图容器：
  ```html
  <div ref="sankeyChartRef" style="height: 500px" />
  ```
- 新增 loading 状态（地图 GeoJSON 加载中显示骨架屏）
- 新增 fallback 提示条（GeoJSON 加载失败后显示横幅）

#### 5.5 生命周期更新

`onMounted` 中：
1. 请求 4 个接口（geo, craft-flow, overview, sankey）
2. 调用 `await loadChinaGeoJSON()`
3. 全部完成后调用 `initCharts()`

`initCharts()` 中：
1. 初始化地图图表（geoChartRef）
2. 初始化饼图图表（craftChartRef，保持不变）
3. 初始化桑基图图表（sankeyChartRef）

`disposeCharts()` 中：
1. 新增 sankeyChart 的 dispose

#### 5.6 fallback 降级数据更新

- `craftFlow` 的 fallback 保持现有的饼图数据
- `geoData` 的 fallback 保持现有散点数据
- `sankeyData` 需要新增 fallback 数据（基于 5 件种子数据构建三层关系）

---

### 6. 前端 — composables 扩展（可选）

**文件**: `client/src/composables/index.ts`（或 `.js`）

如果 `index.ts` 不存在则在 `client/src/composables/index.ts` 新增 `useGeoJSON` composable（可选）。也可以直接在 StatisticsPage.vue 中内联实现加载逻辑，避免过度抽象。

当前项目 composables 使用的是 `index.js` 文件，先用 `.ts` 后缀查找实际文件：

如果 composables 文件是 `.js`，则在 StatisticsPage.vue 中内联 GeoJSON 加载逻辑，不创建新文件。

---

### 7. 验证步骤

#### 7.1 后端验证
```bash
cd server && go run ./cmd/api
curl http://localhost:3001/api/statistics/sankey
```
预期返回格式：
```json
{
  "code": 0,
  "message": null,
  "data": {
    "nodes": [
      {"name": "清"}, {"name": "现代"}, {"name": "元"},
      {"name": "苗族"}, {"name": "汉族"}, ...
      {"name": "蜡染"}, {"name": "刺绣"}, ...
    ],
    "links": [
      {"source": "清", "target": "苗族", "value": 3},
      {"source": "苗族", "target": "蜡染", "value": 2},
      ...
    ]
  }
}
```

#### 7.2 前端验证
```bash
cd client && pnpm run dev
```
1. 打开 `http://localhost:5173/statistics`
2. 确认中国地图底图加载成功（省界清晰）
3. 确认省份气泡叠加正确
4. 确认桑基图三层流向正常（朝代→民族→工艺）
5. 断开网络 → 确认本地 GeoJSON 正常加载
6. 删除 `public/geojson/china.json` → 确认降级到 scatter 模式

#### 7.3 交互验证
- 地图 hover 省份显示 tooltip
- visualMap 滑块调整数值范围
- 桑基图 hover 连线高亮关联路径
- 窗口 resize 时三张图表同步响应

---

### 文件改动汇总

| 文件 | 操作 | 说明 |
|------|------|------|
| `server/internal/service/statistics_service.go` | 修改 | 新增 GetSankeyData 方法 + SankeyData 结构体 |
| `server/internal/handler/statistics.go` | 修改 | 新增 GetSankey handler |
| `server/internal/router/router.go` | 修改 | 新增 /statistics/sankey 路由 |
| `client/src/types/statistics.ts` | 修改 | 新增 SankeyNode / SankeyLink / SankeyData 类型 |
| `client/src/services/api.ts` | 修改 | 新增 getSankey API 方法 |
| `client/public/geojson/china.json` | 新增 | 中国省级 GeoJSON（约 500KB-1MB） |
| `client/src/pages/Statistics/StatisticsPage.vue` | 重写 | 地图改 MapChart + 桑基图替换占位符 |

---

### 前置依赖清单

- [ ] 下载中国省级 GeoJSON 文件到 `client/public/geojson/china.json`
  - 推荐来源：[DataV.GeoAtlas](https://datav.aliyun.com/portal/school/atlas/area_selector) → 中国 → 省级 → 下载 JSON
- [ ] 确认 `echarts` 版本 ≥ 5.0（当前 5.6.0 ✅）

### 风险与注意事项

1. **GeoJSON 文件大小**（~500KB-1MB）— 首次加载有网络开销，但本地文件不依赖外部 CDN，加载快。如果文件过大导致页面卡顿，则切换为方案 A（CDN 加载）
2. **`craft` 字段复合值处理** — "蜡染+刺绣" 这类值需要在后端拆分为多条链路。拆分逻辑用 `strings.Split(craft, "+")` 简单可靠
3. **桑基图节点位置** — 5 条数据下桑基图会很稀疏，建议三层节点各 5-10 个即可。后续数据扩充到 30 件时自然填满
4. **ECharts map 坐标系** — 需要确认 `provinceCoordinates` 中的经纬度与 GeoJSON 坐标系一致（都使用 WGS84 / GCJ02 均可，中国省级地图通常使用 GCJ02）
5. **南海诸岛缩略图** — 建议在 GeoJSON 中保留南海诸岛数据，在 `geo.regions` 中配置缩略图位置

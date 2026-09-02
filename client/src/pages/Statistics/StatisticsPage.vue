<template>
  <div class="page-container">
    <div class="mb-8">
      <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">地理分布与工艺流向</h1>
      <p class="mt-2 text-sm text-mogao-ochre-500">非遗服饰的空间分布格局与工艺传承脉络</p>
    </div>

    <!-- 加载中 -->
    <Skeleton v-if="loading" type="detail" />

    <template v-else>
      <!-- GeoJSON 降级提示 -->
      <div
        v-if="geoFallback"
        class="mb-4 p-3 rounded border-l-4 border-mogao-ochre-400 bg-mogao-ochre-50 text-sm text-mogao-ochre-700"
      >
        地图底图加载失败，已切换为散点图模式。
        <span class="text-mogao-ochre-500">（仍可正常查看各省数据）</span>
      </div>

      <!-- 概览统计卡片 -->
      <div v-if="overview" class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div
          v-for="stat in overviewStats"
          :key="stat.label"
          :class="['card-mogao p-5 text-center border-t-2', `border-mogao-${stat.color}`]"
        >
          <p class="text-xs text-mogao-ochre-500">{{ stat.label }}</p>
          <p :class="['text-3xl font-bold font-serif mt-1', `text-mogao-${stat.color}-600`]">
            {{ stat.value }}
          </p>
          <p class="text-xs text-mogao-ochre-400">{{ stat.unit }}</p>
        </div>
      </div>

      <!-- 图表区域：地图 + 工艺饼图 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <!-- 地理分布地图 -->
        <div class="card-mogao p-6">
          <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-4">
            全国非遗服饰地理分布
          </h3>
          <div ref="geoChartRef" style="height: 500px" />
        </div>

        <!-- 工艺柱状图 -->
        <div class="card-mogao p-6">
          <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-4">
            工艺类型分布
          </h3>
          <div ref="craftChartRef" style="height: 600px" />
        </div>
      </div>

      <!-- 工艺流向旭日图 -->
      <div class="card-mogao mt-8 p-6">
        <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-4">
          朝代 → 民族 → 工艺 流向图
        </h3>
        <p class="text-xs text-mogao-ochre-500 -mt-2 mb-4">
          内圈为朝代，中圈为民族，外圈为工艺 — 点击扇区可深入查看
        </p>
        <div
          v-if="!sankeyData || sankeyData.links.length === 0"
          class="aspect-[2/1] bg-gradient-to-br from-mogao-leadwhite-50 to-mogao-ochre-50 rounded flex items-center justify-center"
        >
          <div class="text-center">
            <span class="text-5xl block mb-3">&#x26b3;</span>
            <p class="text-sm text-mogao-ochre-500">暂无足够数据生成流向图</p>
            <p class="text-xs text-mogao-ochre-400 mt-1">扩充种子数据后将自动展示</p>
          </div>
        </div>
        <div v-else ref="sunburstChartRef" style="height: 600px" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, type Ref, type ComputedRef } from 'vue'
import * as echarts from 'echarts/core'
import { PieChart, MapChart, BarChart, SunburstChart } from 'echarts/charts'
import {
  GridComponent, TooltipComponent, LegendComponent,
  GeoComponent, VisualMapComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { statisticsAPI } from '@/services/api'
import { provinceCoordinates, mogaoColors, craftColors } from '@/utils'
import Skeleton from '@/components/Common/Skeleton.vue'
import type {
  GeoDataItem, CraftFlowItem, OverviewStats, SankeyData,
} from '@/types/statistics'
import type { ApiResponse } from '@/types/api'
import type { EChartsOption } from 'echarts'

echarts.use([
  PieChart, MapChart, BarChart, SunburstChart,
  GridComponent, TooltipComponent, LegendComponent,
  GeoComponent, VisualMapComponent,
  CanvasRenderer,
])

// ───────── 类型 ─────────

interface OverviewStatDisplay {
  label: string
  value: number
  unit: string
  color: string
}

// ───────── 省份名称映射（DB短名 → GeoJSON全名）─────────
const shortToFullProvince: Record<string, string> = {
  '北京': '北京市', '天津': '天津市', '上海': '上海市', '重庆': '重庆市',
  '河北': '河北省', '山西': '山西省', '内蒙古': '内蒙古自治区',
  '辽宁': '辽宁省', '吉林': '吉林省', '黑龙江': '黑龙江省',
  '江苏': '江苏省', '浙江': '浙江省', '安徽': '安徽省',
  '福建': '福建省', '江西': '江西省', '山东': '山东省',
  '河南': '河南省', '湖北': '湖北省', '湖南': '湖南省',
  '广东': '广东省', '广西': '广西壮族自治区', '海南': '海南省',
  '四川': '四川省', '贵州': '贵州省', '云南': '云南省',
  '西藏': '西藏自治区', '陕西': '陕西省', '甘肃': '甘肃省',
  '青海': '青海省', '宁夏': '宁夏回族自治区', '新疆': '新疆维吾尔自治区',
  '台湾': '台湾省', '香港': '香港特别行政区', '澳门': '澳门特别行政区',
}

// ───────── 状态 ─────────

const loading: Ref<boolean> = ref(true)
const geoData: Ref<GeoDataItem[] | null> = ref(null)
const craftFlow: Ref<CraftFlowItem[] | null> = ref(null)
const overview: Ref<OverviewStats | null> = ref(null)
const sankeyData: Ref<SankeyData | null> = ref(null)

/** 是否有注册好的 GeoJSON（用于 MapChart） */
const chinaGeoRegistered: Ref<boolean> = ref(false)
/** 是否触发了降级模式（散点图） */
const geoFallback: Ref<boolean> = ref(false)

// ───────── DOM refs ─────────

const geoChartRef: Ref<HTMLElement | null> = ref(null)
const craftChartRef: Ref<HTMLElement | null> = ref(null)
const sunburstChartRef: Ref<HTMLElement | null> = ref(null)

let geoChart: echarts.ECharts | null = null
let craftChart: echarts.ECharts | null = null
let sunburstChart: echarts.ECharts | null = null
let resizeHandler: (() => void) | null = null

// ───────── 概览统计 ─────────

const overviewStats: ComputedRef<OverviewStatDisplay[]> = computed(() => {
  if (!overview.value) return []
  return [
    { label: '服饰总数', value: overview.value.totalGarments, unit: '件', color: 'cinnabar' },
    { label: '涉及民族', value: overview.value.totalEthnics, unit: '个', color: 'malachite' },
    { label: '工艺种类', value: overview.value.totalCrafts, unit: '种', color: 'ochre' },
    { label: '覆盖省份', value: overview.value.totalProvinces, unit: '个', color: 'lapis' },
  ]
})

// ───────── GeoJSON 加载 ─────────

/**
 * 加载中国省界 GeoJSON — 本地为主，CDN 降级
 * 全部失败则标记 geoFallback=true 切换为散点图
 */
async function loadChinaGeoJSON(): Promise<void> {
  const sources = [
    { url: '/geojson/china.json', label: 'local' },
    { url: 'https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json', label: 'CDN' },
  ]

  for (const src of sources) {
    try {
      const resp = await fetch(src.url, { signal: AbortSignal.timeout(8000) })
      if (!resp.ok) continue
      const json = await resp.json()
      echarts.registerMap('china', json)
      chinaGeoRegistered.value = true
      return
    } catch {
      // 尝试下一个来源
      continue
    }
  }

  // 全部失败 → 降级
  geoFallback.value = true
  chinaGeoRegistered.value = false
}

// ───────── 图表 Option 构建 ─────────

function buildGeoOption(): EChartsOption {
  if (!geoData.value) return {}

  // 有地图底图 → MapChart + effectScatter 叠加
  if (chinaGeoRegistered.value) {
    // 构建 GeoJSON全名 → 数量的映射（用于 visualMap 和 tooltip）
    const provinceMap: Record<string, number> = {}
    geoData.value.forEach((item) => {
      const fullName = shortToFullProvince[item.province] || item.province
      provinceMap[fullName] = (provinceMap[fullName] || 0) + item.count
    })

    // 统计 max 供 visualMap 使用
    const maxCount = Math.max(...geoData.value.map((d) => d.count), 1)

    // 涟漪气泡数据 — 使用 provinceCoordinates 坐标，名称映射为全名
    const scatterData = geoData.value
      .filter((item) => provinceCoordinates[item.province])
      .map((item) => {
        const coords = provinceCoordinates[item.province]
        const fullName = shortToFullProvince[item.province] || item.province
        return {
          name: fullName,
          value: [...coords, item.count] as [number, number, number],
          symbolSize: Math.max(8, Math.sqrt(item.count) * 8),
        }
      })

    return {
      visualMap: {
        min: 0,
        max: maxCount,
        left: 20,
        bottom: 10,
        text: ['多', '少'],
        inRange: {
          color: ['#F5F0E8', '#E9988F', '#C43A31'],
        },
        calculable: true,
        textStyle: { color: mogaoColors.lapis, fontSize: 10 },
      },
      geo: {
        map: 'china',
        roam: true,
        label: { show: false },
        itemStyle: {
          areaColor: '#F5F0E8',
          borderColor: '#C89B5E',
          borderWidth: 0.8,
        },
        emphasis: {
          label: { show: true, color: mogaoColors.lapis },
          itemStyle: { areaColor: '#F5EBD8' },
        },
      },
      tooltip: {
        trigger: 'item',
        formatter: (params: any) => {
          if (params.seriesType === 'effectScatter') {
            return `${params.name}<br/>非遗服饰数量：${params.value[2]} 件`
          }
          const count = provinceMap[params.name] || 0
          return `${params.name}<br/>非遗服饰数量：${count} 件`
        },
      },
      series: [
        {
          type: 'map',
          map: 'china',
          geoIndex: 0,
          data: geoData.value.map((item) => ({
            name: shortToFullProvince[item.province] || item.province,
            value: item.count,
          })),
        },
        {
          type: 'effectScatter',
          coordinateSystem: 'geo',
          data: scatterData,
          symbolSize: (val: number[]) => Math.max(8, Math.sqrt(val[2]) * 8),
          showEffectOn: 'render',
          rippleEffect: {
            brushType: 'stroke',
            scale: 2.5,
            period: 4,
          },
          itemStyle: {
            color: mogaoColors.gold,
            shadowBlur: 10,
            shadowColor: 'rgba(184, 134, 45, 0.5)',
          },
          label: {
            show: true,
            formatter: '{b}',
            position: 'top',
            color: mogaoColors.lapis,
            fontSize: 10,
            fontFamily: 'Noto Serif SC, serif',
          },
          emphasis: {
            scale: 1.5,
          },
        },
      ],
    }
  }

  // 降级：散点图（保留旧逻辑）
  return {
    tooltip: {
      trigger: 'item',
      formatter: '{b}<br/>非遗服饰数量：{c} 件',
    },
    grid: { top: 10, right: 20, bottom: 10, left: 20 },
    xAxis: { type: 'value', show: false, min: 75, max: 135 },
    yAxis: { type: 'value', show: false, min: 15, max: 55 },
    series: [{
      type: 'scatter',
      data: geoData.value.map((item) => {
        const coords = provinceCoordinates[item.province] || [0, 0]
        return {
          name: item.province,
          value: [...coords, item.count] as [number, number, number],
          symbolSize: Math.max(8, Math.sqrt(item.count) * 6),
        }
      }),
      itemStyle: {
        color: mogaoColors.cinnabar,
        shadowBlur: 10,
        shadowColor: 'rgba(196, 58, 49, 0.5)',
      },
      label: {
        show: true,
        formatter: '{b}',
        position: 'top',
        color: mogaoColors.lapis,
        fontSize: 11,
        fontFamily: 'Noto Serif SC, serif',
      },
      emphasis: {
        scale: 1.3,
        itemStyle: { color: mogaoColors.gold },
      },
    }],
  }
}

function buildCraftOption(): EChartsOption {
  if (!craftFlow.value) return {}

  // 按数量降序排列，柱状图从下到上显示（最小值在底部 → 数据反转）
  const sorted = [...craftFlow.value].sort((a, b) => a.count - b.count)
  const names = sorted.map((item) => item.craft)
  const values = sorted.map((item) => item.count)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const p = Array.isArray(params) ? params[0] : params
        return `${p.name}<br/>服饰数量：<b>${p.value}</b> 件`
      },
    },
    grid: { left: 100, right: 60, top: 10, bottom: 20 },
    xAxis: {
      type: 'value',
      name: '件',
      nameTextStyle: { color: mogaoColors.ochre, fontSize: 11 },
      axisLabel: { color: mogaoColors.ochre, fontSize: 10 },
      splitLine: { lineStyle: { color: 'rgba(200, 155, 94, 0.15)' } },
    },
    yAxis: {
      type: 'category',
      data: names,
      axisLine: { lineStyle: { color: mogaoColors.ochre } },
      axisLabel: {
        color: mogaoColors.lapis,
        fontSize: 12,
        fontFamily: 'Noto Serif SC, serif',
      },
      axisTick: { show: false },
    },
    series: [{
      type: 'bar',
      data: values.map((v, i) => ({
        value: v,
        itemStyle: {
          color: craftColors[i % craftColors.length],
          borderRadius: [0, 3, 3, 0],
        },
      })),
      barMaxWidth: 22,
      label: {
        show: true,
        position: 'right',
        color: mogaoColors.lapis,
        fontSize: 11,
        fontFamily: 'Noto Serif SC, serif',
      },
      emphasis: {
        itemStyle: { shadowBlur: 8, shadowColor: 'rgba(0,0,0,0.15)' },
      },
    }],
  }
}

/**
 * 将桑基图数据（nodes + links）转换为旭日图层级数据
 * 链路结构：朝代 → 民族 → 工艺
 */
function buildSunburstData(): Record<string, any> | null {
  if (!sankeyData.value) return null

  const { nodes, links } = sankeyData.value

  // 推断节点层级：仅 source → dynasty；source+target → ethnic；仅 target → craft
  const allSources = new Set(links.map((l) => l.source))
  const allTargets = new Set(links.map((l) => l.target))
  const nodeLevel: Record<string, 'dynasty' | 'ethnic' | 'craft'> = {}

  for (const node of nodes) {
    if (allSources.has(node.name) && !allTargets.has(node.name)) {
      nodeLevel[node.name] = 'dynasty'
    } else if (allSources.has(node.name) && allTargets.has(node.name)) {
      nodeLevel[node.name] = 'ethnic'
    } else if (!allSources.has(node.name) && allTargets.has(node.name)) {
      nodeLevel[node.name] = 'craft'
    }
  }

  // 构建 ethnic → craft 映射
  const ethnicCrafts: Record<string, Array<{ name: string; value: number }>> = {}
  for (const l of links) {
    const srcLevel = nodeLevel[l.source]
    const tgtLevel = nodeLevel[l.target]
    if (srcLevel === 'ethnic' && tgtLevel === 'craft') {
      if (!ethnicCrafts[l.source]) ethnicCrafts[l.source] = []
      ethnicCrafts[l.source].push({ name: l.target, value: l.value })
    }
  }

  // 构建 dynasty → ethnic 映射
  const dynastyChildren: Record<string, Array<{ name: string; value: number; children: any[] }>> = {}
  for (const l of links) {
    const srcLevel = nodeLevel[l.source]
    const tgtLevel = nodeLevel[l.target]
    if (srcLevel === 'dynasty' && tgtLevel === 'ethnic') {
      if (!dynastyChildren[l.source]) dynastyChildren[l.source] = []
      dynastyChildren[l.source].push({
        name: l.target,
        value: l.value,
        children: (ethnicCrafts[l.target] || []).map((c) => ({
          name: c.name,
          value: c.value,
        })),
      })
    }
  }

  // 构建顶级 sunburst data
  const result: Record<string, any>[] = []
  for (const dynasty of Object.keys(dynastyChildren)) {
    result.push({
      name: dynasty,
      itemStyle: { color: '#2E4A7A' },
      children: dynastyChildren[dynasty].map((eth) => ({
        name: eth.name,
        itemStyle: { color: '#C43A31' },
        children: eth.children.map((craft) => ({
          name: craft.name,
          value: craft.value,
          itemStyle: { color: '#2E7D5A' },
        })),
      })),
    })
  }

  return result
}

/**
 * 旭日图：朝代 → 民族 → 工艺 三层环形展示
 * 内圈=朝代，中圈=民族，外圈=工艺
 */
function buildSunburstOption(): EChartsOption {
  const data = buildSunburstData()
  if (!data) return {}

  return {
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        // 构建完整路径
        const path: string[] = []
        let node = params
        while (node) {
          if (node.name) path.unshift(node.name)
          node = node.parentNode
        }
        const pct = params.percent ? ` (${params.percent}%)` : ''
        return `${path.join(' → ')}<br/>关联数量：${params.value ?? '-'} 件${pct}`
      },
    },
    series: [{
      type: 'sunburst',
      data: data as any[],
      radius: ['0%', '75%'],
      center: ['50%', '50%'],
      sort: 'desc',
      emphasis: {
        focus: 'ancestor',
        label: { fontSize: 14, fontWeight: 'bold' },
      },
      levels: [
        {},
        {
          // 内圈：朝代
          r0: '10%',
          r: '28%',
          label: {
            rotate: 'tangential',
            fontSize: 12,
            color: mogaoColors.lapis,
            fontFamily: 'Noto Serif SC, serif',
          },
          itemStyle: { borderWidth: 2, borderColor: '#fff' },
        },
        {
          // 中圈：民族
          r0: '28%',
          r: '50%',
          label: {
            rotate: 'tangential',
            fontSize: 10,
            color: mogaoColors.cinnabar,
            fontFamily: 'Noto Serif SC, serif',
          },
          itemStyle: { borderWidth: 1.5, borderColor: '#fff' },
        },
        {
          // 外圈：工艺
          r0: '50%',
          r: '75%',
          label: {
            position: 'outside',
            padding: 4,
            fontSize: 9,
            color: mogaoColors.malachite,
            fontFamily: 'Noto Serif SC, serif',
            silent: true,
          },
          itemStyle: { borderWidth: 1, borderColor: '#fff' },
        },
      ],
    }],
  }
}

// ───────── 图表生命周期 ─────────

function initCharts(): void {
  // 地理分布图（地图或散点）
  if (geoChartRef.value) {
    geoChart = echarts.init(geoChartRef.value)
    geoChart.setOption(buildGeoOption())
  }

  // 工艺饼图
  if (craftChartRef.value) {
    craftChart = echarts.init(craftChartRef.value)
    craftChart.setOption(buildCraftOption())
  }

  // 旭日图（仅在有数据时渲染）
  if (sunburstChartRef.value && sankeyData.value && sankeyData.value.links.length > 0) {
    sunburstChart = echarts.init(sunburstChartRef.value)
    sunburstChart.setOption(buildSunburstOption())
  }

  resizeHandler = () => {
    geoChart?.resize()
    craftChart?.resize()
    sunburstChart?.resize()
  }
  window.addEventListener('resize', resizeHandler)
}

function disposeCharts(): void {
  geoChart?.dispose()
  craftChart?.dispose()
  sunburstChart?.dispose()
  geoChart = null
  craftChart = null
  sunburstChart = null
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }
}

// ───────── 数据获取 ─────────

onMounted(async () => {
  try {
    const [geoRes, flowRes, overviewRes, sankeyRes] = await Promise.all([
      statisticsAPI.getGeoDistribution(),
      statisticsAPI.getCraftFlow(),
      statisticsAPI.getOverview(),
      statisticsAPI.getSankey(),
    ])

    geoData.value = (geoRes as ApiResponse<GeoDataItem[]>).data || (geoRes as unknown as GeoDataItem[])
    craftFlow.value = (flowRes as ApiResponse<CraftFlowItem[]>).data || (flowRes as unknown as CraftFlowItem[])
    overview.value = (overviewRes as ApiResponse<OverviewStats>).data || (overviewRes as unknown as OverviewStats)
    sankeyData.value = (sankeyRes as ApiResponse<SankeyData>).data || (sankeyRes as unknown as SankeyData)
  } catch {
    // fallback 数据 — 与 35 件种子数据保持一致
    geoData.value = [
      { province: '贵州', count: 5 }, { province: '云南', count: 4 },
      { province: '四川', count: 3 }, { province: '西藏', count: 3 },
      { province: '新疆', count: 3 }, { province: '广西', count: 2 },
      { province: '湖南', count: 2 }, { province: '内蒙古', count: 2 },
      { province: '江苏', count: 2 }, { province: '浙江', count: 2 },
      { province: '广东', count: 1 }, { province: '福建', count: 1 },
      { province: '青海', count: 1 }, { province: '甘肃', count: 1 },
      { province: '宁夏', count: 1 }, { province: '陕西', count: 1 },
      { province: '湖北', count: 1 }, { province: '海南', count: 1 },
      { province: '黑龙江', count: 1 }, { province: '吉林', count: 1 },
      { province: '辽宁', count: 1 }, { province: '台湾', count: 1 },
    ]
    craftFlow.value = [
      { craft: '刺绣', count: 8 }, { craft: '织锦', count: 5 },
      { craft: '蜡染', count: 4 }, { craft: '氆氇织造', count: 4 },
      { craft: '挑花', count: 3 }, { craft: '贴花', count: 3 },
      { craft: '缂丝', count: 2 }, { craft: '扎染', count: 2 },
      { craft: '擀毡', count: 2 }, { craft: '珠绣', count: 2 },
    ]
    overview.value = {
      totalGarments: 35, totalEthnics: 26, totalCrafts: 32, totalProvinces: 22,
    }
    sankeyData.value = {
      nodes: [
        { name: '清' }, { name: '明' }, { name: '元' }, { name: '现代' }, { name: '古代' },
        { name: '苗族' }, { name: '汉族' }, { name: '藏族' }, { name: '蒙古族' }, { name: '傣族' },
        { name: '彝族' }, { name: '侗族' }, { name: '布依族' }, { name: '羌族' },
        { name: '壮族' }, { name: '瑶族' }, { name: '土家族' },
        { name: '维吾尔族' }, { name: '哈萨克族' }, { name: '柯尔克孜族' },
        { name: '白族' }, { name: '哈尼族' }, { name: '畲族' },
        { name: '门巴族' }, { name: '裕固族' }, { name: '回族' },
        { name: '黎族' }, { name: '赫哲族' }, { name: '朝鲜族' },
        { name: '满族' }, { name: '高山族' },
        { name: '蜡染' }, { name: '刺绣' }, { name: '织锦' }, { name: '氆氇织造' },
        { name: '挑花' }, { name: '扎染' }, { name: '缂丝' }, { name: '擀毡' },
        { name: '贴花' }, { name: '珠绣' }, { name: '银饰锻造' }, { name: '蜀绣' },
      ],
      links: [
        { source: '清', target: '苗族', value: 4 }, { source: '清', target: '彝族', value: 2 },
        { source: '清', target: '藏族', value: 3 }, { source: '清', target: '傣族', value: 1 },
        { source: '清', target: '侗族', value: 1 }, { source: '清', target: '布依族', value: 1 },
        { source: '清', target: '壮族', value: 1 }, { source: '清', target: '瑶族', value: 1 },
        { source: '清', target: '土家族', value: 2 }, { source: '清', target: '维吾尔族', value: 1 },
        { source: '清', target: '哈萨克族', value: 1 }, { source: '清', target: '柯尔克孜族', value: 1 },
        { source: '清', target: '畲族', value: 1 }, { source: '清', target: '门巴族', value: 1 },
        { source: '清', target: '裕固族', value: 1 }, { source: '清', target: '黎族', value: 1 },
        { source: '清', target: '赫哲族', value: 1 }, { source: '清', target: '满族', value: 1 },
        { source: '清', target: '高山族', value: 1 }, { source: '清', target: '汉族', value: 2 },
        { source: '明', target: '苗族', value: 2 }, { source: '明', target: '汉族', value: 1 },
        { source: '元', target: '蒙古族', value: 1 },
        { source: '现代', target: '汉族', value: 3 }, { source: '现代', target: '藏族', value: 1 },
        { source: '现代', target: '白族', value: 1 }, { source: '现代', target: '羌族', value: 1 },
        { source: '现代', target: '回族', value: 1 }, { source: '现代', target: '朝鲜族', value: 1 },
        { source: '古代', target: '哈尼族', value: 1 },
        { source: '苗族', target: '蜡染', value: 3 }, { source: '苗族', target: '刺绣', value: 5 },
        { source: '苗族', target: '织锦', value: 1 }, { source: '苗族', target: '挑花', value: 2 },
        { source: '苗族', target: '银饰锻造', value: 1 },
        { source: '汉族', target: '刺绣', value: 5 }, { source: '汉族', target: '缂丝', value: 1 },
        { source: '汉族', target: '蜀绣', value: 1 }, { source: '汉族', target: '贴花', value: 1 },
        { source: '藏族', target: '氆氇织造', value: 4 }, { source: '蒙古族', target: '刺绣', value: 2 },
        { source: '蒙古族', target: '贴花', value: 2 }, { source: '傣族', target: '织锦', value: 1 },
        { source: '彝族', target: '蜡染', value: 1 }, { source: '彝族', target: '挑花', value: 1 },
        { source: '彝族', target: '贴花', value: 1 }, { source: '彝族', target: '擀毡', value: 1 },
        { source: '侗族', target: '刺绣', value: 1 }, { source: '布依族', target: '蜡染', value: 1 },
        { source: '羌族', target: '刺绣', value: 1 }, { source: '壮族', target: '织锦', value: 1 },
        { source: '瑶族', target: '刺绣', value: 1 }, { source: '瑶族', target: '挑花', value: 1 },
        { source: '土家族', target: '织锦', value: 3 }, { source: '土家族', target: '刺绣', value: 1 },
        { source: '维吾尔族', target: '织锦', value: 1 },
        { source: '哈萨克族', target: '刺绣', value: 1 }, { source: '哈萨克族', target: '贴花', value: 1 },
        { source: '柯尔克孜族', target: '擀毡', value: 1 }, { source: '柯尔克孜族', target: '刺绣', value: 1 },
        { source: '白族', target: '扎染', value: 1 }, { source: '哈尼族', target: '织锦', value: 1 },
        { source: '畲族', target: '刺绣', value: 1 }, { source: '畲族', target: '织锦', value: 1 },
        { source: '门巴族', target: '氆氇织造', value: 1 }, { source: '裕固族', target: '刺绣', value: 1 },
        { source: '裕固族', target: '珠绣', value: 1 }, { source: '回族', target: '刺绣', value: 1 },
        { source: '黎族', target: '织锦', value: 1 }, { source: '赫哲族', target: '贴花', value: 1 },
        { source: '朝鲜族', target: '刺绣', value: 1 }, { source: '满族', target: '刺绣', value: 1 },
        { source: '高山族', target: '珠绣', value: 1 },
      ],
    }
  } finally {
    // 先加载 GeoJSON（在 loading 状态下并行），再初始化图表
    await loadChinaGeoJSON()
    loading.value = false
    setTimeout(initCharts, 0)
  }
})

onUnmounted(disposeCharts)
</script>

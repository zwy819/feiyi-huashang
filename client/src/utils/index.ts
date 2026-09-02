/**
 * 工具函数模块
 */

/** 格式化日期为中文格式 */
export function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}年${month}月${day}日`
}

/** 格式化文件大小 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

/** 省份名称映射到经纬度坐标 */
export const provinceCoordinates: Record<string, [number, number]> = {
  '北京': [116.40, 39.90],
  '天津': [117.20, 39.13],
  '上海': [121.47, 31.23],
  '重庆': [106.55, 29.57],
  '河北': [114.52, 38.05],
  '山西': [112.55, 37.87],
  '内蒙古': [111.75, 40.83],
  '辽宁': [123.43, 41.30],
  '吉林': [125.33, 43.90],
  '黑龙江': [126.53, 45.80],
  '江苏': [118.78, 32.06],
  '浙江': [120.15, 30.28],
  '安徽': [117.28, 31.86],
  '福建': [119.30, 26.08],
  '江西': [115.90, 28.68],
  '河南': [113.67, 34.76],
  '湖北': [114.30, 30.60],
  '湖南': [112.98, 28.20],
  '广东': [113.27, 23.13],
  '广西': [108.37, 22.82],
  '海南': [110.35, 20.02],
  '四川': [104.07, 30.67],
  '贵州': [106.71, 26.57],
  '云南': [102.71, 25.05],
  '西藏': [91.14, 29.65],
  '陕西': [108.95, 34.27],
  '甘肃': [103.83, 36.06],
  '青海': [101.78, 36.62],
  '宁夏': [106.27, 38.47],
  '新疆': [87.62, 43.79],
  '台湾': [121.52, 25.03],
  '香港': [114.17, 22.28],
  '澳门': [113.55, 22.19],
}

/** 莫高窟色系常量（与 Tailwind 配置保持一致，供 JS/图表使用） */
export const mogaoColors: Record<string, string> = {
  cinnabar: '#C43A31',
  malachite: '#2E7D5A',
  ochre: '#C89B5E',
  lapis: '#2E4A7A',
  leadwhite: '#F5F0E8',
  gold: '#B8862D',
}

/** 工艺流向配色 */
export const craftColors: string[] = [
  '#C43A31', '#2E7D5A', '#C89B5E', '#2E4A7A', '#B8862D',
  '#8B5E3C', '#5C4033', '#7B3F4A', '#3D5A4B', '#6B4E31',
]

// Delete old JS file

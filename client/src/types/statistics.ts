export interface GeoDataItem {
  province: string
  count: number
}

export interface CraftFlowItem {
  craft: string
  count: number
}

export interface OverviewStats {
  totalGarments: number
  totalEthnics: number
  totalCrafts: number
  totalProvinces: number
}

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

export interface DashboardStats {
  total_garments: number
  total_models: number
  total_users: number
  total_favorites: number
  recent_uploads: Array<{
    id: number
    original_name: string
    file_size: number
    created_at: string
  }>
}

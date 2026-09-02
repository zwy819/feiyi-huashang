export interface FavoriteItem {
  id: number
  user_id: number
  garment_id: number
  created_at: string
  garment_name?: string
  garment_dynasty?: string
  garment_ethnic?: string
  garment_craft?: string
  garment_province?: string
  garment_image_url?: string
}

export interface FavoriteCheckResult {
  isFavorited: boolean
}

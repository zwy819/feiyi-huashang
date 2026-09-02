export interface Garment {
  id: number
  name: string
  dynasty: string
  ethnic: string
  craft: string
  material: string
  pattern: string
  cultural_meaning: string
  province: string
  description: string
  image_url: string
  model_url: string
  model_id: number | null
  created_at: string
  updated_at: string
}

export interface GarmentListItem {
  id: number
  name: string
  dynasty: string
  ethnic: string
  craft: string
  province: string
  model_id: number | null
}

export interface GarmentFormData {
  name: string
  dynasty: string
  ethnic: string
  craft: string
  material: string
  pattern: string
  cultural_meaning: string
  province: string
  description: string
  image_url: string
  model_id: number | null
}

export interface RecommendItem {
  id: number
  name: string
  similarity: number
}

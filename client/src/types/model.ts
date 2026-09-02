export interface ModelInfo {
  id: number
  filename: string
  original_name: string
  file_size: number
  mime_type: string
  file_path: string
  garment_id: number | null
  garment_name: string | null
  created_at: string
  updated_at: string
}

export interface ModelListItem {
  id: number
  filename: string
  original_name: string
  file_size: number
  mime_type: string
  garment_id: number | null
  garment_name: string | null
  created_at: string
  updated_at: string
}

export interface ModelUploadResult {
  id: number
  filename: string
  original_name: string
  file_size: number
  mime_type: string
  file_path: string
  garment_id: number | null
  created_at: string
  updated_at: string
}

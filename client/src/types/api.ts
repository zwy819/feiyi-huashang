export interface ApiResponse<T = unknown> {
  code: number
  message: string | null
  data: T
}

export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

export interface PaginatedParams {
  page?: number
  limit?: number
  search?: string
}

export interface GarmentQueryParams extends PaginatedParams {
  dynasty?: string
  ethnic?: string
  craft?: string
}

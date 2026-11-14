export interface TourStop {
  order: number
  garment_id: number
  name: string
  ethnic: string
  craft: string
  province: string
  annotation_id: number
  annotation_title: string
  duration_min: number
  reason: string
}

export interface TourPlan {
  title: string
  minutes: number
  total_minutes: number
  stops: TourStop[]
}

export const tourAPI = {
  async plan(query: string): Promise<TourPlan> {
    const response = await fetch('/api/tour/plan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query }),
    })
    const payload = await response.json()
    if (!response.ok) throw new Error(payload?.message || '路线生成失败')
    return payload.data as TourPlan
  },
}

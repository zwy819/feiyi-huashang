/**
 * 热点标注类型定义
 *
 * 标注是挂在 3D 模型上的纹样/工艺讲解点，
 * position_* 为前端展示场景的世界坐标（模型平移居中后不旋转，坐标稳定）
 */
export interface Annotation {
  id: number
  garment_id: number
  /** 服饰名称（仅管理员列表接口返回） */
  garment_name?: string
  title: string
  content: string
  position_x: number
  position_y: number
  position_z: number
  sort_order: number
  created_at: string
  updated_at: string
}

/** 新建/编辑热点标注请求体 */
export interface AnnotationPayload {
  garment_id?: number
  title: string
  content: string
  position_x: number
  position_y: number
  position_z: number
  sort_order?: number
}

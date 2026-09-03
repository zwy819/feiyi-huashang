import axios, { type AxiosInstance } from 'axios'
import type {
  ApiResponse,
  GarmentQueryParams,
  PaginatedData,
} from '@/types/api'
import type { Garment } from '@/types/garment'
import type { ModelInfo, ModelListItem } from '@/types/model'
import type { AuthResponse } from '@/types/user'
import type {
  GeoDataItem,
  CraftFlowItem,
  OverviewStats,
  SankeyData,
  DashboardStats,
} from '@/types/statistics'
import type { FavoriteItem, FavoriteCheckResult } from '@/types/favorite'
import type { Submission } from '@/types/submission'
import type { Comment } from '@/types/comment'

/**
 * API 客户端配置
 *
 * 基础 URL 由 Vite 代理转发到后端 3001 端口
 * 自动附加 Authorization 请求头
 */

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 附加 JWT Token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器 - 统一错误处理
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message =
      error.response?.data?.message || error.message || '请求失败'

    // 401 未登录 → 清除本地 token 并跳转登录页
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (typeof window !== 'undefined') {
        window.__showToast?.('登录已过期，请重新登录', 'info')
        setTimeout(() => {
          window.location.href = '/login'
        }, 800)
      }
    } else {
      if (typeof window !== 'undefined') {
        window.__showToast?.(message, 'error')
      }
    }

    return Promise.reject(new Error(message))
  },
)

// ===== 服饰相关 API =====
export const garmentAPI = {
  /** 获取服饰列表 */
  list: (params?: GarmentQueryParams) =>
    api.get<unknown, ApiResponse<PaginatedData<Garment>>>('/garments', { params }),

  /** 获取单件服饰详情 */
  getById: (id: number) =>
    api.get<unknown, ApiResponse<Garment>>(`/garments/${id}`),

  /** 获取推荐链 */
  getRecommendChain: (id: number) =>
    api.get<unknown, ApiResponse<{ items: Array<{ id: number; name: string; similarity: number }> }>>(
      `/garments/${id}/recommend`,
    ),

  /** 获取每日一服 */
  getDaily: () => api.get<unknown, ApiResponse<Garment>>('/daily'),
}

// ===== 收藏相关 API =====
export const favoriteAPI = {
  /** 获取收藏列表 */
  list: () =>
    api.get<unknown, ApiResponse<{ favorites: FavoriteItem[] }>>('/favorites'),

  /** 添加收藏 */
  add: (garmentId: number) =>
    api.post<unknown, ApiResponse<null>>('/favorites', { garment_id: garmentId }),

  /** 移除收藏 */
  remove: (garmentId: number) =>
    api.delete<unknown, ApiResponse<null>>(`/favorites/${garmentId}`),

  /** 检查是否已收藏 */
  check: (garmentId: number) =>
    api.get<unknown, ApiResponse<FavoriteCheckResult>>(
      `/favorites/check/${garmentId}`,
    ),
}

// ===== 统计相关 API =====
export const statisticsAPI = {
  /** 获取地理分布数据 */
  getGeoDistribution: () =>
    api.get<unknown, ApiResponse<GeoDataItem[]>>('/statistics/geo'),

  /** 获取工艺流向数据 */
  getCraftFlow: () =>
    api.get<unknown, ApiResponse<CraftFlowItem[]>>('/statistics/craft-flow'),

  /** 获取概览统计 */
  getOverview: () =>
    api.get<unknown, ApiResponse<OverviewStats>>('/statistics/overview'),

  /** 获取桑基图流向数据 */
  getSankey: () =>
    api.get<unknown, ApiResponse<SankeyData>>('/statistics/sankey'),
}

// ===== 报告相关 API =====
export const reportAPI = {
  /** 生成并下载 PDF 报告 */
  generate: (params: { ids: string }) =>
    api.get('/report/generate', {
      params,
      responseType: 'blob',
    }),
}

// ===== 认证相关 API =====
export const authAPI = {
  /** 用户注册 */
  register: (data: { username: string; email: string; password: string }) =>
    api.post<unknown, ApiResponse<AuthResponse>>('/auth/register', data),

  /** 用户登录 */
  login: (data: { username: string; password: string }) =>
    api.post<unknown, ApiResponse<AuthResponse>>('/auth/login', data),

  /** 获取当前用户信息 */
  getProfile: () =>
    api.get<unknown, ApiResponse<{ id: number; username: string; email: string; role: string }>>(
      '/auth/profile',
    ),
}

// ===== 模型相关 API =====
export const modelAPI = {
  /** 获取公开模型列表 */
  list: (params?: { page?: number; limit?: number; garment_id?: number }) =>
    api.get<unknown, ApiResponse<PaginatedData<ModelListItem>>>('/models', { params }),

  /** 获取模型文件 URL（用于 Three.js 加载） */
  getFileUrl: (modelId: number): string => `/api/models/${modelId}/file`,
}

// ===== 管理员 API =====
export const adminModelAPI = {
  list: (params?: { page?: number; limit?: number; search?: string }) =>
    api.get<unknown, ApiResponse<PaginatedData<ModelInfo>>>('/admin/models', { params }),

  getById: (id: number) =>
    api.get<unknown, ApiResponse<ModelInfo>>(`/admin/models/${id}`),

  upload: (formData: FormData) =>
    api.post<unknown, ApiResponse<ModelInfo>>('/admin/models', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  update: (id: number, data: Partial<Pick<ModelInfo, 'original_name' | 'garment_id'>>) =>
    api.put<unknown, ApiResponse<ModelInfo>>(`/admin/models/${id}`, data),

  delete: (id: number) =>
    api.delete<unknown, ApiResponse<null>>(`/admin/models/${id}`),
}

export const adminGarmentAPI = {
  create: (data: Record<string, unknown>) =>
    api.post<unknown, ApiResponse<Garment>>('/admin/garments', data),

  update: (id: number, data: Record<string, unknown>) =>
    api.put<unknown, ApiResponse<Garment>>(`/admin/garments/${id}`, data),

  delete: (id: number) =>
    api.delete<unknown, ApiResponse<null>>(`/admin/garments/${id}`),
}

export const adminDashboardAPI = {
  getStats: () =>
    api.get<unknown, ApiResponse<DashboardStats>>('/admin/dashboard'),
}

// ===== 投稿相关 API =====
export const submissionAPI = {
  /** 用户提交投稿 */
  submit: (formData: FormData) =>
    api.post<unknown, ApiResponse<Submission>>('/submissions', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  /** 用户查看自己的投稿列表 */
  listMy: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get<unknown, ApiResponse<PaginatedData<Submission>>>('/submissions', { params }),

  /** 用户查看自己的单条投稿 */
  getMy: (id: number) =>
    api.get<unknown, ApiResponse<Submission>>(`/submissions/${id}`),

  /** 获取投稿图片 URL */
  getImageUrl: (submissionId: number): string =>
    `/api/submissions/${submissionId}/image`,

  /** 用户撤回自己的投稿 */
  retract: (id: number) =>
    api.put<unknown, ApiResponse<null>>(`/submissions/${id}/retract`),

  /** 获取公开的通过投稿列表（用户贡献区域） */
  listPublic: (params?: { page?: number; limit?: number }) =>
    api.get<unknown, ApiResponse<PaginatedData<Submission>>>('/public-submissions', { params }),
}

export const adminSubmissionAPI = {
  /** 管理员获取投稿列表 */
  list: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get<unknown, ApiResponse<PaginatedData<Submission>>>('/admin/submissions', { params }),

  /** 管理员获取单条投稿详情 */
  getById: (id: number) =>
    api.get<unknown, ApiResponse<Submission>>(`/admin/submissions/${id}`),

  /** 管理员审核通过 */
  approve: (id: number) =>
    api.put<unknown, ApiResponse<null>>(`/admin/submissions/${id}/approve`),

  /** 管理员驳回 */
  reject: (id: number, reject_reason: string) =>
    api.put<unknown, ApiResponse<null>>(`/admin/submissions/${id}/reject`, { reject_reason }),

  /** 管理员标记完成（关联 garment 和 model） */
  complete: (id: number, data: { garment_id?: number; model_id?: number }) =>
    api.put<unknown, ApiResponse<null>>(`/admin/submissions/${id}/complete`, data),

  /** 管理员删除投稿（含图片文件） */
  remove: (id: number) =>
    api.delete<unknown, ApiResponse<null>>(`/admin/submissions/${id}`),
}

// ===== 评论相关 API =====
export const commentAPI = {
  /** 获取服饰评论列表（公开，登录后可见自己的待审核/已驳回评论） */
  list: (params: { garment_id: number; page?: number; limit?: number }) =>
    api.get<unknown, ApiResponse<PaginatedData<Comment>>>('/comments', { params }),

  /** 发表评论（审核通过后公开） */
  create: (data: { garment_id: number; content: string }) =>
    api.post<unknown, ApiResponse<Comment>>('/comments', data),

  /** 编辑自己的评论（重新进入审核） */
  update: (id: number, content: string) =>
    api.put<unknown, ApiResponse<Comment>>(`/comments/${id}`, { content }),

  /** 删除自己的评论 */
  remove: (id: number) =>
    api.delete<unknown, ApiResponse<null>>(`/comments/${id}`),

  /** 点赞评论 */
  like: (id: number) =>
    api.post<unknown, ApiResponse<Comment>>(`/comments/${id}/like`),

  /** 取消点赞 */
  unlike: (id: number) =>
    api.delete<unknown, ApiResponse<Comment>>(`/comments/${id}/like`),
}

export const adminCommentAPI = {
  /** 管理员获取评论列表（可按状态筛选） */
  list: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get<unknown, ApiResponse<PaginatedData<Comment>>>('/admin/comments', { params }),

  /** 管理员审核通过评论 */
  approve: (id: number) =>
    api.put<unknown, ApiResponse<null>>(`/admin/comments/${id}/approve`),

  /** 管理员驳回评论 */
  reject: (id: number, reject_reason: string) =>
    api.put<unknown, ApiResponse<null>>(`/admin/comments/${id}/reject`, { reject_reason }),

  /** 管理员删除评论 */
  remove: (id: number) =>
    api.delete<unknown, ApiResponse<null>>(`/admin/comments/${id}`),
}

export default api

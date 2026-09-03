export interface Comment {
  id: number
  garment_id: number
  garment_name?: string
  user_id: number
  username: string
  content: string
  status: CommentStatus
  reject_reason: string
  like_count: number
  is_liked: boolean
  created_at: string
  updated_at: string
}

export type CommentStatus = 'pending' | 'approved' | 'rejected'

export const CommentStatusLabels: Record<CommentStatus, string> = {
  pending: '审核中',
  approved: '已通过',
  rejected: '已驳回',
}

export const CommentStatusColors: Record<CommentStatus, string> = {
  pending: 'bg-amber-100 text-amber-700',
  approved: 'bg-green-100 text-green-700',
  rejected: 'bg-red-100 text-red-700',
}

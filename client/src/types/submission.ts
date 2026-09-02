export interface Submission {
  id: number
  user_id: number
  username?: string
  image_filename: string
  image_original_name: string
  image_file_size: number
  image_mime_type: string
  image_file_path: string
  name: string
  dynasty: string
  ethnic: string
  craft: string
  material: string
  pattern: string
  cultural_meaning: string
  province: string
  description: string
  status: SubmissionStatus
  reject_reason: string
  admin_note: string
  created_garment_id: number | null
  created_model_id: number | null
  created_at: string
  updated_at: string
}

export type SubmissionStatus = 'pending' | 'approved' | 'rejected' | 'completed' | 'retracted'

export const SubmissionStatusLabels: Record<SubmissionStatus, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已驳回',
  completed: '已完成',
  retracted: '已撤回',
}

export const SubmissionStatusColors: Record<SubmissionStatus, string> = {
  pending: 'bg-amber-100 text-amber-700',
  approved: 'bg-green-100 text-green-700',
  rejected: 'bg-red-100 text-red-700',
  completed: 'bg-blue-100 text-blue-700',
  retracted: 'bg-gray-100 text-gray-500',
}

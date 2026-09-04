/**
 * 华裳问答小游戏类型定义
 */

/** 单道题目（不含答案，判分由服务端完成） */
export interface QuizQuestion {
  id: string
  garment_id: number
  /** 判分字段：name / ethnic / dynasty / province / craft / pattern */
  field: string
  type: string
  stem: string
  options: string[]
  explanation: string
}

/** 提交的单题答案 */
export interface QuizAnswerItem {
  garment_id: number
  field: string
  choice: string
}

/** 服务端返回的单题判定 */
export interface QuizResultItem {
  garment_id: number
  field: string
  choice: string
  answer: string
  correct: boolean
}

/** 提交判分结果 */
export interface QuizSubmitResult {
  score: number
  total: number
  correct_count: number
  best_score: number
  duration: number
  results: QuizResultItem[]
}

/** 排行榜条目（每用户最高分） */
export interface QuizLeaderboardItem {
  user_id: number
  username: string
  score: number
  total: number
  correct_count: number
  duration_seconds: number
  created_at: string
}

/** 我的答题记录 */
export interface QuizRecordItem {
  score: number
  total: number
  correct_count: number
  duration_seconds: number
  created_at: string
}

export interface QuizMyRecords {
  best: number
  records: QuizRecordItem[]
}

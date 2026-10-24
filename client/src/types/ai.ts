export interface AiSource {
  id: string
  type: string
  title: string
  publisher?: string
  url?: string
  snippet: string
  garment_id?: number
  annotation_id?: number
}

export interface AiAction {
  type: 'highlight_annotation' | string
  garment_id?: number
  annotation_id?: number
  label: string
}

export interface AiToolEvent {
  name: string
  status: string
  summary: string
  elapsed_ms: number
}

export interface AiUsage {
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export interface AiStatus {
  enabled: boolean
  configured: boolean
  mock_mode: boolean
  model: string
  embedding_model?: string
  chunks: number
  message: string
}

export interface AiChatContext {
  page?: string
  garment_id?: number
  annotation_id?: number
}

export interface AiChatTurn {
  role: 'user' | 'assistant'
  content: string
}

export interface AiChatRequest {
  message: string
  history: AiChatTurn[]
  context: AiChatContext
}

export type AiMessageRole = 'user' | 'assistant'

export interface AiMessage {
  id: string
  serverMessageId?: string
  role: AiMessageRole
  content: string
  sources: AiSource[]
  tools: AiToolEvent[]
  action?: AiAction
  streaming?: boolean
  stopped?: boolean
  error?: string
  feedback?: 1 | -1
  usage?: AiUsage
}

export interface AiDoneEvent {
  conversation_id: string
  message_id: string
  content: string
  usage: AiUsage
  latency_ms: number
  sources_used: string[]
}

export interface AiErrorEvent {
  code: number
  message: string
  retryable: boolean
}

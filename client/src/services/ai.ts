import type {
  AiChatRequest,
  AiDoneEvent,
  AiErrorEvent,
  AiSource,
  AiStatus,
  AiToolEvent,
} from '@/types/ai'

export class AiRequestError extends Error {
  code: number
  retryable: boolean

  constructor(message: string, code = 0, retryable = true) {
    super(message)
    this.name = 'AiRequestError'
    this.code = code
    this.retryable = retryable
  }
}

export interface AiStreamHandlers {
  onMeta?: (data: Record<string, unknown>) => void
  onSources?: (sources: AiSource[]) => void
  onTool?: (tool: AiToolEvent) => void
  onDelta?: (text: string) => void
  onAction?: (action: Record<string, unknown>) => void
  onDone?: (data: AiDoneEvent) => void
  onError?: (data: AiErrorEvent) => void
}

async function parseError(response: Response): Promise<AiRequestError> {
  try {
    const payload = await response.json()
    return new AiRequestError(
      payload?.message || `AI 请求失败（${response.status}）`,
      Number(payload?.code) || response.status,
      response.status >= 500,
    )
  } catch {
    return new AiRequestError(`AI 请求失败（${response.status}）`, response.status, response.status >= 500)
  }
}

export const aiAPI = {
  async getStatus(): Promise<AiStatus> {
    const response = await fetch('/api/ai/status', { headers: { Accept: 'application/json' } })
    if (!response.ok) throw await parseError(response)
    const payload = await response.json()
    return payload.data as AiStatus
  },

  async streamChat(
    request: AiChatRequest,
    handlers: AiStreamHandlers,
    signal?: AbortSignal,
  ): Promise<void> {
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'text/event-stream' },
      body: JSON.stringify(request),
      signal,
    })
    if (!response.ok) throw await parseError(response)
    if (!response.body) throw new AiRequestError('浏览器未收到流式响应', 50201, true)

    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true }).replace(/\r\n/g, '\n')
      let separator = buffer.indexOf('\n\n')
      while (separator >= 0) {
        const block = buffer.slice(0, separator)
        buffer = buffer.slice(separator + 2)
        dispatchEventBlock(block, handlers)
        separator = buffer.indexOf('\n\n')
      }
    }
    buffer += decoder.decode()
    if (buffer.trim()) dispatchEventBlock(buffer, handlers)
  },

  async feedback(messageID: string, rating: 1 | -1, reason = ''): Promise<void> {
    const response = await fetch('/api/ai/feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message_id: messageID, rating, reason }),
    })
    if (!response.ok) throw await parseError(response)
  },
}

function dispatchEventBlock(block: string, handlers: AiStreamHandlers): void {
  let event = 'message'
  const dataLines: string[] = []
  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) event = line.slice(6).trim()
    else if (line.startsWith('data:')) dataLines.push(line.slice(5).trimStart())
  }
  if (dataLines.length === 0) return
  const raw = dataLines.join('\n')
  let data: unknown
  try {
    data = JSON.parse(raw)
  } catch {
    return
  }

  switch (event) {
    case 'meta':
      handlers.onMeta?.(data as Record<string, unknown>)
      break
    case 'sources':
      handlers.onSources?.(data as AiSource[])
      break
    case 'tool':
      handlers.onTool?.(data as AiToolEvent)
      break
    case 'delta':
      handlers.onDelta?.((data as { text?: string }).text || '')
      break
    case 'action':
      handlers.onAction?.(data as Record<string, unknown>)
      break
    case 'done':
      handlers.onDone?.(data as AiDoneEvent)
      break
    case 'error':
      handlers.onError?.(data as AiErrorEvent)
      break
  }
}

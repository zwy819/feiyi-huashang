import { computed, ref } from 'vue'
import { aiAPI, AiRequestError } from '@/services/ai'
import type {
  AiAction,
  AiChatContext,
  AiChatTurn,
  AiMessage,
  AiSource,
  AiStatus,
  AiToolEvent,
} from '@/types/ai'

const messages = ref<AiMessage[]>([])
const isOpen = ref(false)
const isStreaming = ref(false)
const error = ref('')
const status = ref<AiStatus | null>(null)
const context = ref<AiChatContext>({ page: 'home' })
const activeAction = ref<AiAction | null>(null)
const actionTick = ref(0)

let controller: AbortController | null = null
let statusPromise: Promise<void> | null = null

function ensureStatus(): Promise<void> {
  if (status.value) return Promise.resolve()
  if (statusPromise) return statusPromise
  statusPromise = aiAPI
    .getStatus()
    .then((value) => {
      status.value = value
    })
    .catch(() => {
      status.value = {
        enabled: false,
        configured: false,
        mock_mode: false,
        model: '',
        chunks: 0,
        message: '无法连接 AI 导览服务',
      }
    })
    .finally(() => {
      statusPromise = null
    })
  return statusPromise
}

function open(nextContext?: AiChatContext): void {
  if (nextContext) context.value = { ...nextContext }
  isOpen.value = true
  void ensureStatus()
}

function close(): void {
  isOpen.value = false
}

function setContext(nextContext: AiChatContext): void {
  context.value = { ...nextContext }
}

function makeMessage(role: AiMessage['role'], content: string): AiMessage {
  return {
    id: `${role}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    role,
    content,
    sources: [],
    tools: [],
    streaming: role === 'assistant',
  }
}

function historyForRequest(): AiChatTurn[] {
  return messages.value
    .filter((message) => message.content.trim() !== '' && !message.error)
    .slice(-8)
    .map((message) => ({ role: message.role, content: message.content }))
}

async function ask(text: string): Promise<void> {
  const question = text.trim()
  if (!question || isStreaming.value) return
  await ensureStatus()
  if (!status.value?.configured) {
    error.value = status.value?.message || 'AI 导览暂未配置'
    isOpen.value = true
    return
  }

  const history = historyForRequest()
  const userMessage = makeMessage('user', question)
  const assistant = makeMessage('assistant', '')
  messages.value.push(userMessage, assistant)
  error.value = ''
  activeAction.value = null
  isStreaming.value = true
  isOpen.value = true

  controller = new AbortController()
  // Read the item back through the reactive array; mutating the raw object
  // passed to push() would not trigger Vue updates.
  const current = messages.value[messages.value.length - 1]

  try {
    await aiAPI.streamChat(
      { message: question, history, context: context.value },
      {
        onMeta: (data) => {
          if (typeof data.message_id === 'string') current.serverMessageId = data.message_id
        },
        onSources: (sources: AiSource[]) => {
          current.sources = sources
        },
        onTool: (tool: AiToolEvent) => {
          current.tools = [...current.tools, tool]
        },
        onDelta: (delta: string) => {
          current.content += delta
        },
        onAction: (raw) => {
          const action = raw as unknown as AiAction
          if (action && typeof action.label === 'string') {
            current.action = action
            activeAction.value = action
            actionTick.value += 1
          }
        },
        onDone: (done) => {
          current.content = done.content || current.content
          current.serverMessageId = done.message_id
          current.usage = done.usage
          current.streaming = false
        },
        onError: (failure) => {
          current.error = failure.message
          current.streaming = false
          error.value = failure.message
        },
      },
      controller.signal,
    )
  } catch (reason) {
    if (reason instanceof DOMException && reason.name === 'AbortError') {
      current.stopped = true
    } else if (reason instanceof AiRequestError) {
      current.error = reason.message
      error.value = reason.message
    } else {
      current.error = 'AI 导览连接中断，请重试'
      error.value = current.error
    }
  } finally {
    current.streaming = false
    isStreaming.value = false
    controller = null
  }
}

function stop(): void {
  controller?.abort()
  controller = null
  isStreaming.value = false
}

function clear(): void {
  if (isStreaming.value) stop()
  messages.value = []
  activeAction.value = null
  error.value = ''
}

async function retryLast(): Promise<void> {
  if (isStreaming.value) return
  const index = [...messages.value].map((m) => m.role).lastIndexOf('user')
  if (index < 0) return
  const question = messages.value[index].content
  messages.value = messages.value.slice(0, index)
  error.value = ''
  await ask(question)
}

async function feedback(message: AiMessage, rating: 1 | -1): Promise<void> {
  if (!message.serverMessageId || message.feedback) return
  const previous = message.feedback
  message.feedback = rating
  try {
    await aiAPI.feedback(message.serverMessageId, rating)
  } catch {
    message.feedback = previous
  }
}

function focusAction(action: AiAction): void {
  activeAction.value = action
  actionTick.value += 1
}

const suggestions = computed<string[]>(() => {
  if (context.value.page === 'detail') {
    return [
      '这件服饰最特别的工艺是什么？',
      '它在 3D 模型中有哪些讲解点？',
      '推荐几件相似馆藏',
    ]
  }
  if (context.value.page === 'statistics') {
    return ['哪个省的馆藏最多？', '哪个民族的服饰最多？', '这些分布数据说明了什么？']
  }
  return [
    '想找一件适合秋天婚礼、有刺绣元素的传统服饰',
    '苗族百褶裙的裙褶为什么要压这么多道？',
    '哪个省的馆藏最多？',
  ]
})

export function useAiChat() {
  return {
    messages,
    isOpen,
    isStreaming,
    error,
    status,
    context,
    activeAction,
    actionTick,
    suggestions,
    open,
    close,
    setContext,
    ask,
    stop,
    clear,
    retryLast,
    feedback,
    focusAction,
  }
}

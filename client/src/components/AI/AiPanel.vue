<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[60]">
      <div class="absolute inset-0 bg-mogao-lapis-900/30 backdrop-blur-[2px]" @click="close" />

      <section
        class="absolute inset-0 md:inset-y-0 md:left-auto md:right-0 md:w-[460px] bg-mogao-leadwhite-50 shadow-2xl flex flex-col md:border-l md:border-mogao-ochre-200"
        role="dialog"
        aria-modal="true"
        aria-label="织语 AI 导览"
        :data-streaming="isStreaming ? 'true' : 'false'"
      >
        <header class="relative overflow-hidden bg-mogao-lapis-900 text-white px-5 pt-5 pb-4">
          <div class="absolute -right-10 -top-12 w-40 h-40 rounded-full bg-mogao-gold/20 blur-2xl" />
          <div class="relative flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <span class="w-8 h-8 rounded-full bg-mogao-cinnabar flex items-center justify-center font-serif text-sm">织</span>
                <div>
                  <h2 class="font-serif text-lg font-bold tracking-wide">织语 · AI 导览</h2>
                  <p class="text-[11px] text-mogao-leadwhite-300 mt-0.5">会讲解 · 能指认 · 可溯源</p>
                </div>
              </div>
              <div class="mt-3 flex flex-wrap items-center gap-2 text-[11px] text-mogao-leadwhite-300">
                <span class="inline-flex items-center gap-1.5 rounded-full bg-white/10 px-2.5 py-1">
                  <span class="w-1.5 h-1.5 rounded-full" :class="statusDotClass" />
                  {{ statusText }}
                </span>
                <span class="rounded-full bg-white/10 px-2.5 py-1">{{ contextLabel }}</span>
              </div>
            </div>
            <button
              class="p-2 rounded-full text-mogao-leadwhite-300 hover:text-white hover:bg-white/10 transition-colors"
              aria-label="关闭 AI 导览"
              @click="close"
            >✕</button>
          </div>
        </header>

        <div
          ref="scrollEl"
          class="flex-1 overflow-y-auto px-4 py-5 space-y-5"
          aria-live="polite"
          @scroll="onScroll"
        >
          <section v-if="messages.length === 0" class="pt-4">
            <p class="text-xs tracking-[0.25em] text-mogao-gold-600 uppercase">Ask the collection</p>
            <h3 class="font-serif text-2xl font-bold text-mogao-lapis-800 mt-2 leading-snug">
              想了解哪件<br />非遗服饰？
            </h3>
            <p class="text-sm text-mogao-lapis-600 leading-relaxed mt-3">
              我会先查馆藏和资料，再回答你。每个关键结论都带来源；如果资料里没有，我会直接说没有。
            </p>

            <div class="mt-5 space-y-2">
              <button
                v-for="suggestion in suggestions"
                :key="suggestion"
                class="w-full text-left rounded-xl border border-mogao-ochre-200 bg-white px-3.5 py-3 text-xs leading-relaxed text-mogao-lapis-700 hover:border-mogao-gold-400 hover:bg-mogao-gold-50/50 transition-colors"
                @click="submitSuggestion(suggestion)"
              >
                <span class="text-mogao-cinnabar mr-1.5">◇</span>{{ suggestion }}
              </button>
            </div>

            <p class="mt-5 text-[11px] leading-relaxed text-mogao-ochre-500">
              匿名可用，不保存完整对话；3D 模型为艺术化复原，AI 回答不作为鉴定依据。
            </p>
          </section>

          <MessageBubble
            v-for="message in messages"
            :key="message.id"
            :message="message"
            @cite="openCitation"
            @sources="openSources"
            @retry="retryLast"
            @feedback="feedback"
            @action="handleAction"
          />
          <div ref="endRef" />
        </div>

        <div v-if="error && !isStreaming" class="mx-4 mb-2 rounded-lg border border-mogao-cinnabar-200 bg-mogao-cinnabar-50 px-3 py-2 text-xs text-mogao-cinnabar-700">
          {{ error }}
        </div>

        <footer class="border-t border-mogao-ochre-200 bg-white/80 px-4 pt-3 pb-4">
          <form class="relative" @submit.prevent="submit">
            <textarea
              ref="inputRef"
              v-model="draft"
              rows="2"
              maxlength="500"
              class="w-full resize-none rounded-xl border border-mogao-ochre-300 bg-white px-3.5 py-3 pr-24 text-sm leading-relaxed text-mogao-lapis-800 placeholder:text-mogao-ochre-400 focus:border-mogao-gold-500 focus:outline-none focus:ring-2 focus:ring-mogao-gold-200"
              placeholder="问一件服饰、一种工艺，或一个地域…"
              aria-label="输入你的问题"
              @keydown.enter.exact.prevent="submit"
            />
            <div class="absolute right-2.5 bottom-2.5 flex items-center gap-1.5">
              <button
                v-if="isStreaming"
                type="button"
                class="rounded-lg border border-mogao-ochre-300 px-2.5 py-1.5 text-[11px] text-mogao-lapis-600 hover:bg-mogao-ochre-50"
                @click="stop"
              >停止</button>
              <button
                v-else
                type="submit"
                class="rounded-lg bg-mogao-cinnabar px-3 py-1.5 text-[11px] font-medium text-white shadow-sm hover:bg-mogao-cinnabar-600 disabled:opacity-40"
                :disabled="!draft.trim()"
              >发送</button>
            </div>
          </form>
          <div class="mt-2 flex items-center justify-between text-[10px] text-mogao-ochre-400">
            <span>AI 生成内容，请以馆藏说明与来源为准</span>
            <button class="hover:text-mogao-cinnabar" @click="clear">清空对话</button>
          </div>
        </footer>

        <SourceDrawer
          :open="drawerOpen"
          :sources="drawerSources"
          :highlight="drawerHighlight"
          @close="drawerOpen = false"
        />
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MessageBubble from './MessageBubble.vue'
import SourceDrawer from './SourceDrawer.vue'
import { useAiChat } from '@/composables/useAiChat'
import type { AiAction, AiMessage, AiSource } from '@/types/ai'

const {
  messages,
  isStreaming,
  status,
  error,
  context,
  suggestions,
  close,
  ask,
  stop,
  clear,
  retryLast,
  feedback,
  focusAction,
} = useAiChat()

const route = useRoute()
const router = useRouter()
const draft = ref('')
const inputRef = ref<HTMLTextAreaElement | null>(null)
const scrollEl = ref<HTMLElement | null>(null)
const endRef = ref<HTMLElement | null>(null)
const follow = ref(true)

const drawerOpen = ref(false)
const drawerSources = ref<AiSource[]>([])
const drawerHighlight = ref('')

const statusText = computed(() => {
  if (!status.value) return '连接中…'
  if (status.value.mock_mode) return '演示模式'
  if (status.value.configured) return status.value.model || '已连接'
  return '未配置模型'
})
const statusDotClass = computed(() => {
  if (!status.value) return 'bg-mogao-ochre-300'
  if (status.value.mock_mode) return 'bg-mogao-gold-400'
  return status.value.configured ? 'bg-mogao-malachite-400' : 'bg-mogao-cinnabar-400'
})
const contextLabel = computed(() => {
  if (context.value.page === 'detail') return '当前服饰上下文'
  if (context.value.page === 'statistics') return '馆藏统计上下文'
  return '全馆导览'
})

function submit(): void {
  const value = draft.value.trim()
  if (!value || isStreaming.value) return
  draft.value = ''
  void ask(value).finally(() => nextTick(() => inputRef.value?.focus()))
}

function submitSuggestion(value: string): void {
  if (isStreaming.value) return
  void ask(value)
}

function onScroll(): void {
  const el = scrollEl.value
  if (!el) return
  follow.value = el.scrollHeight - el.scrollTop - el.clientHeight < 90
}

function scrollToBottom(): void {
  if (!follow.value) return
  void nextTick(() => endRef.value?.scrollIntoView({ behavior: 'smooth', block: 'end' }))
}

watch(
  () => messages.value.map((message) => message.content.length + message.tools.length).join('|'),
  scrollToBottom,
)
watch(
  () => messages.value.length,
  () => {
    follow.value = true
    scrollToBottom()
  },
)

function openSources(message: AiMessage): void {
  drawerSources.value = message.sources
  drawerHighlight.value = ''
  drawerOpen.value = true
}

function openCitation(id: string): void {
  const message = [...messages.value].reverse().find((item) => item.sources.some((source) => source.id === id))
  if (!message) return
  drawerSources.value = message.sources
  drawerHighlight.value = id
  drawerOpen.value = true
}

function handleAction(action: AiAction): void {
  if (!action.garment_id) return
  if (action.type === 'highlight_annotation' && action.annotation_id) {
    if (route.path === `/detail/${action.garment_id}`) {
      focusAction(action)
      close()
      return
    }
    void router.push({
      path: `/detail/${action.garment_id}`,
      query: { annotation: String(action.annotation_id) },
    })
    close()
    return
  }
  void router.push(`/detail/${action.garment_id}`)
  close()
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    if (drawerOpen.value) drawerOpen.value = false
    else close()
  }
}

onMounted(() => {
  document.body.style.overflow = 'hidden'
  window.addEventListener('keydown', onKeydown)
  void nextTick(() => inputRef.value?.focus())
})
onUnmounted(() => {
  document.body.style.overflow = ''
  window.removeEventListener('keydown', onKeydown)
})
</script>

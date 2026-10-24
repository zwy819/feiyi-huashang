<template>
  <div :class="message.role === 'user' ? 'flex justify-end' : 'flex justify-start'">
    <div
      v-if="message.role === 'user'"
      class="max-w-[88%] rounded-2xl rounded-br-sm bg-mogao-lapis-800 text-mogao-leadwhite-50 px-4 py-3 text-sm leading-relaxed shadow-sm"
    >
      {{ message.content }}
    </div>

    <article v-else class="w-full max-w-[94%]">
      <div class="flex items-center gap-2 mb-2">
        <span class="w-6 h-6 rounded-full bg-mogao-cinnabar text-white text-[11px] flex items-center justify-center font-serif">织</span>
        <span class="text-xs font-semibold text-mogao-lapis-700">织语</span>
        <span v-if="message.streaming" class="text-[11px] text-mogao-gold-600 animate-pulse">正在讲解…</span>
        <span v-else-if="message.stopped" class="text-[11px] text-mogao-ochre-500">已停止</span>
      </div>

      <div class="rounded-2xl rounded-tl-sm border border-mogao-ochre-200 bg-white px-4 py-3 shadow-sm">
        <div v-if="!message.content && message.streaming" class="flex items-center gap-2 py-1 text-xs text-mogao-ochre-500">
          <span class="flex gap-1">
            <i class="w-1.5 h-1.5 rounded-full bg-mogao-gold-500 animate-bounce [animation-delay:-0.2s]" />
            <i class="w-1.5 h-1.5 rounded-full bg-mogao-gold-500 animate-bounce [animation-delay:-0.1s]" />
            <i class="w-1.5 h-1.5 rounded-full bg-mogao-gold-500 animate-bounce" />
          </span>
          正在查馆藏、翻资料…
        </div>

        <div
          v-if="message.content"
          class="ai-markdown text-sm leading-relaxed text-mogao-lapis-700"
          @click="onContentClick"
          v-html="rendered"
        />
        <span v-if="message.streaming && message.content" class="inline-block w-1.5 h-4 ml-1 align-text-bottom bg-mogao-cinnabar animate-pulse" />

        <ToolTrace :tools="message.tools" />

        <div v-if="message.error" class="mt-3 rounded-md border border-mogao-cinnabar-200 bg-mogao-cinnabar-50 px-3 py-2 text-xs text-mogao-cinnabar-700">
          {{ message.error }}
          <button class="ml-2 underline font-medium" @click="$emit('retry')">重试</button>
        </div>

        <div v-if="message.action" class="mt-3">
          <button
            class="inline-flex items-center gap-2 rounded-lg border border-mogao-gold-400 bg-mogao-gold-50 px-3 py-2 text-xs font-medium text-mogao-lapis-800 hover:bg-mogao-gold-100 transition-colors"
            @click="$emit('action', message.action)"
          >
            <span class="text-mogao-cinnabar">◎</span>
            {{ message.action.label }}
            <span aria-hidden="true">→</span>
          </button>
        </div>

        <div class="mt-3 flex flex-wrap items-center gap-2 border-t border-mogao-ochre-100 pt-2.5">
          <button
            v-if="message.sources.length"
            class="inline-flex items-center gap-1 rounded-full bg-mogao-ochre-50 px-2.5 py-1 text-[11px] text-mogao-lapis-600 hover:bg-mogao-ochre-100"
            @click="$emit('sources', message)"
          >
            来源 {{ message.sources.length }}
            <span aria-hidden="true">↗</span>
          </button>
          <span v-if="message.usage" class="text-[10px] text-mogao-ochre-400">
            {{ message.usage.total_tokens }} tokens
          </span>
          <span class="flex-1" />
          <button
            class="p-1.5 rounded-full transition-colors"
            :class="message.feedback === 1 ? 'text-mogao-malachite-600 bg-mogao-malachite-50' : 'text-mogao-ochre-400 hover:text-mogao-malachite-600'"
            aria-label="回答有帮助"
            @click="$emit('feedback', message, 1)"
          >👍</button>
          <button
            class="p-1.5 rounded-full transition-colors"
            :class="message.feedback === -1 ? 'text-mogao-cinnabar bg-mogao-cinnabar-50' : 'text-mogao-ochre-400 hover:text-mogao-cinnabar'"
            aria-label="回答没帮助"
            @click="$emit('feedback', message, -1)"
          >👎</button>
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ToolTrace from './ToolTrace.vue'
import { renderAiMarkdown } from '@/utils/markdown'
import type { AiMessage } from '@/types/ai'

const props = defineProps<{ message: AiMessage }>()

const emit = defineEmits<{
  (e: 'cite', id: string): void
  (e: 'sources', message: AiMessage): void
  (e: 'retry'): void
  (e: 'feedback', message: AiMessage, rating: 1 | -1): void
  (e: 'action', action: NonNullable<AiMessage['action']>): void
}>()

const rendered = computed(() => renderAiMarkdown(props.message.content))

function onContentClick(event: MouseEvent): void {
  const target = event.target as HTMLElement | null
  const cite = target?.closest<HTMLElement>('[data-cite]')
  if (cite?.dataset.cite) {
    emit('cite', cite.dataset.cite)
  }
}
</script>

<style scoped>
.ai-markdown :deep(p) {
  margin: 0.35rem 0;
}
.ai-markdown :deep(p:first-child) {
  margin-top: 0;
}
.ai-markdown :deep(p:last-child) {
  margin-bottom: 0;
}
.ai-markdown :deep(h3) {
  margin: 0.8rem 0 0.35rem;
  font-family: var(--font-serif, serif);
  font-size: 0.95rem;
  font-weight: 700;
  color: #2e4a7a;
}
.ai-markdown :deep(ul),
.ai-markdown :deep(ol) {
  margin: 0.4rem 0;
  padding-left: 1.15rem;
}
.ai-markdown :deep(ul) {
  list-style: disc;
}
.ai-markdown :deep(ol) {
  list-style: decimal;
}
.ai-markdown :deep(li) {
  margin: 0.2rem 0;
}
.ai-markdown :deep(strong) {
  color: #8a2f28;
  font-weight: 600;
}
.ai-markdown :deep(code) {
  border-radius: 3px;
  background: #f5efe6;
  padding: 0.05rem 0.25rem;
  font-size: 0.85em;
}
.ai-markdown :deep(.ai-cite) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.35rem;
  height: 1.15rem;
  margin: 0 0.12rem;
  padding: 0 0.22rem;
  border-radius: 999px;
  background: #f5e8d8;
  border: 1px solid #d9b98b;
  color: #8a5a1f;
  font-size: 0.65rem;
  font-weight: 700;
  line-height: 1;
  vertical-align: 0.12rem;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}
.ai-markdown :deep(.ai-cite:hover) {
  background: #c43a31;
  border-color: #c43a31;
  color: #fff;
}
</style>

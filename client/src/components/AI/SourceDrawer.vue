<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="fixed inset-0 z-[70]">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-[2px]" @click="$emit('close')" />
        <aside class="absolute right-0 top-0 h-full w-full max-w-md bg-mogao-leadwhite-50 shadow-2xl flex flex-col">
          <header class="flex items-start justify-between gap-3 px-5 py-4 border-b border-mogao-ochre-200">
            <div>
              <p class="text-xs tracking-[0.2em] text-mogao-gold-600 uppercase">Sources</p>
              <h2 class="font-serif text-lg font-bold text-mogao-lapis-800 mt-1">回答来源</h2>
              <p class="text-xs text-mogao-ochre-500 mt-1">共 {{ sources.length }} 条 · 点击可查看原文或馆藏</p>
            </div>
            <button
              class="p-2 rounded-full text-mogao-ochre-500 hover:text-mogao-cinnabar hover:bg-mogao-cinnabar-50 transition-colors"
              aria-label="关闭来源"
              @click="$emit('close')"
            >✕</button>
          </header>

          <div class="flex-1 overflow-y-auto px-4 py-4 space-y-3">
            <component
              :is="source.url ? (isExternal(source.url) ? 'a' : 'router-link') : 'div'"
              v-for="source in sources"
              :key="source.id"
              :href="source.url && isExternal(source.url) ? source.url : undefined"
              :to="source.url && !isExternal(source.url) ? source.url : undefined"
              :target="source.url && isExternal(source.url) ? '_blank' : undefined"
              :rel="source.url && isExternal(source.url) ? 'noopener noreferrer' : undefined"
              class="block rounded-lg border p-3.5 transition-colors"
              :class="highlight === source.id
                ? 'border-mogao-cinnabar bg-mogao-cinnabar-50/60'
                : 'border-mogao-ochre-200 bg-white hover:border-mogao-gold-400'"
            >
              <div class="flex items-center gap-2">
                <span class="inline-flex items-center justify-center min-w-8 h-6 px-1.5 rounded bg-mogao-lapis-800 text-white text-[11px] font-semibold">
                  {{ source.id }}
                </span>
                <span class="text-[11px] px-2 py-0.5 rounded-full bg-mogao-ochre-100 text-mogao-ochre-700">{{ source.type }}</span>
                <span v-if="source.annotation_id" class="text-[11px] text-mogao-malachite-600">3D 可定位</span>
              </div>
              <h3 class="font-serif text-sm font-semibold text-mogao-lapis-800 mt-2">{{ source.title }}</h3>
              <p v-if="source.publisher" class="text-[11px] text-mogao-ochre-500 mt-0.5">{{ source.publisher }}</p>
              <p class="text-xs text-mogao-lapis-600 leading-relaxed mt-2 line-clamp-4">{{ source.snippet }}</p>
              <p v-if="source.url" class="text-[11px] text-mogao-cinnabar mt-2">
                {{ isExternal(source.url) ? '打开原文 ↗' : '查看馆藏 →' }}
              </p>
            </component>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import type { AiSource } from '@/types/ai'

defineProps<{
  open: boolean
  sources: AiSource[]
  highlight?: string
}>()

defineEmits<{ (e: 'close'): void }>()

function isExternal(url: string): boolean {
  return /^https?:\/\//i.test(url)
}
</script>

<style scoped>
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.22s ease;
}
.drawer-enter-active aside,
.drawer-leave-active aside {
  transition: transform 0.25s ease;
}
.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}
.drawer-enter-from aside,
.drawer-leave-to aside {
  transform: translateX(24px);
}
</style>

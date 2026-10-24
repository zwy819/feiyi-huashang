<template>
  <div v-if="tools.length" class="mt-3 rounded-lg border border-mogao-ochre-200 bg-mogao-ochre-50/60 overflow-hidden">
    <button
      class="w-full flex items-center justify-between gap-3 px-3 py-2 text-left"
      :aria-expanded="expanded"
      @click="expanded = !expanded"
    >
      <span class="flex items-center gap-2 text-xs text-mogao-lapis-600">
        <span class="w-1.5 h-1.5 rounded-full bg-mogao-malachite-500" />
        已完成 {{ tools.length }} 次馆藏检索
      </span>
      <span class="text-[11px] text-mogao-ochre-500">{{ expanded ? '收起' : '查看过程' }}</span>
    </button>
    <Transition name="trace">
      <ol v-if="expanded" class="px-3 pb-3 space-y-2 border-t border-mogao-ochre-200 pt-2">
        <li v-for="(tool, index) in tools" :key="`${tool.name}-${index}`" class="flex items-start gap-2 text-[11px]">
          <span class="mt-0.5 text-mogao-gold-600">{{ toolIcon(tool.name) }}</span>
          <span class="flex-1">
            <span class="text-mogao-lapis-700">{{ toolLabel(tool.name) }}</span>
            <span class="text-mogao-ochre-500"> · {{ tool.summary }}</span>
          </span>
          <span class="text-mogao-ochre-400 tabular-nums">{{ tool.elapsed_ms }}ms</span>
        </li>
      </ol>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { AiToolEvent } from '@/types/ai'

defineProps<{ tools: AiToolEvent[] }>()
const expanded = ref(false)

function toolLabel(name: string): string {
  switch (name) {
    case 'search_garments': return '检索馆藏'
    case 'get_annotations': return '查找 3D 热点'
    case 'find_similar': return '寻找相似馆藏'
    case 'get_statistics': return '查询馆藏统计'
    default: return name
  }
}

function toolIcon(name: string): string {
  switch (name) {
    case 'get_statistics': return '▦'
    case 'get_annotations': return '◎'
    case 'find_similar': return '≈'
    default: return '⌕'
  }
}
</script>

<style scoped>
.trace-enter-active,
.trace-leave-active {
  transition: opacity 0.18s ease;
}
.trace-enter-from,
.trace-leave-to {
  opacity: 0;
}
</style>

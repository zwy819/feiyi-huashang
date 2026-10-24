<template>
  <div v-if="visible">
    <Transition name="dock">
      <button
        v-if="!isOpen"
        class="fixed right-4 bottom-5 z-50 group flex items-center gap-2 rounded-full bg-mogao-cinnabar px-4 py-3 text-white shadow-mogao-lg hover:bg-mogao-cinnabar-600 transition-all"
        aria-label="打开织语 AI 导览"
        @click="open()"
      >
        <span class="relative flex h-6 w-6 items-center justify-center">
          <span class="absolute inline-flex h-full w-full rounded-full bg-white/30 animate-ping opacity-30" />
          <span class="relative font-serif text-sm">织</span>
        </span>
        <span class="text-xs font-medium tracking-wide">问织语</span>
      </button>
    </Transition>

    <component :is="AiPanel" v-if="isOpen" />
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAiChat } from '@/composables/useAiChat'

const AiPanel = defineAsyncComponent(() => import('./AiPanel.vue'))
const { isOpen, open, setContext } = useAiChat()
const route = useRoute()

const visible = computed(() => !route.path.startsWith('/admin') && route.path !== '/login')

function syncContext(): void {
  if (route.path.startsWith('/detail/')) {
    const garmentID = Number(route.params.id)
    const annotationID = Number(route.query.annotation)
    setContext({
      page: 'detail',
      ...(Number.isFinite(garmentID) && garmentID > 0 ? { garment_id: garmentID } : {}),
      ...(Number.isFinite(annotationID) && annotationID > 0 ? { annotation_id: annotationID } : {}),
    })
    return
  }
  if (route.path === '/statistics') {
    setContext({ page: 'statistics' })
    return
  }
  if (route.path === '/') {
    setContext({ page: 'home' })
    return
  }
  setContext({ page: route.path.replace('/', '') || 'home' })
}

watch(() => route.fullPath, syncContext, { immediate: true })
</script>

<style scoped>
.dock-enter-active,
.dock-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.dock-enter-from,
.dock-leave-to {
  opacity: 0;
  transform: translateY(8px) scale(0.96);
}
</style>

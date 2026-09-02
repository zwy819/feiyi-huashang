<template>
  <Teleport to="body">
    <div class="fixed top-20 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup name="toast">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="pointer-events-auto"
        >
          <div
            :class="[
              'flex items-center gap-2.5 px-4 py-3 rounded-lg shadow-lg text-sm font-medium min-w-[260px] max-w-[380px]',
              typeStyles[t.type] || typeStyles.error,
            ]"
          >
            <span class="flex-shrink-0 w-5 h-5 rounded-full bg-white/20 flex items-center justify-center text-xs">
              {{ typeIcons[t.type] || typeIcons.error }}
            </span>
            <span class="flex-1 leading-snug">{{ t.message }}</span>
            <button
              @click="onClose(t.id)"
              class="flex-shrink-0 w-5 h-5 rounded-full bg-white/20 hover:bg-white/30 flex items-center justify-center text-xs transition-colors"
            >
              ✕
            </button>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { inject, type Ref } from 'vue'

interface ToastItem {
  id: number
  message: string
  type: string
}

const toasts = inject<Ref<ToastItem[]>>('toasts', { value: [] } as unknown as Ref<ToastItem[]>)
const onClose = inject<(id: number) => void>('removeToast', () => {})

const typeStyles: Record<string, string> = {
  error: 'bg-mogao-cinnabar text-white',
  success: 'bg-mogao-malachite text-white',
  info: 'bg-mogao-lapis text-white',
}

const typeIcons: Record<string, string> = {
  error: '✕',
  success: '✓',
  info: 'ℹ',
}
</script>

<style scoped>
.toast-enter-active {
  transition: all 0.25s ease;
}
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(80px) scale(0.95);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(80px) scale(0.95);
}
</style>

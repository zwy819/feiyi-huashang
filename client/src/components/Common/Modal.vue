<template>
  <Teleport to="body">
    <Transition name="overlay">
      <div
        v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @keydown.esc="onClose"
      >
        <!-- 遮罩 -->
        <div
          class="absolute inset-0 bg-black/50 backdrop-blur-sm"
          @click="onClose"
        />

        <!-- 弹窗内容 -->
        <div
          :class="[
            'relative w-full mx-4 bg-white rounded-lg shadow-mogao-lg border border-mogao-ochre-200',
            sizeClass,
          ]"
        >
          <!-- 标题栏 -->
          <div
            v-if="title"
            class="flex items-center justify-between px-6 py-4 border-b border-mogao-ochre-100"
          >
            <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800">{{ title }}</h3>
            <button
              @click="onClose"
              class="p-1 rounded-mogao text-mogao-ochre-500 hover:text-mogao-cinnabar hover:bg-mogao-ochre-50 transition-colors"
              aria-label="关闭"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- 内容 -->
          <div class="px-6 py-4">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  isOpen?: boolean
  title?: string
  size?: 'sm' | 'md' | 'lg'
}>()

const emit = defineEmits<{
  close: []
}>()

function onClose() {
  emit('close')
}

const sizeClasses: Record<string, string> = {
  sm: 'max-w-sm',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
}

const sizeClass = computed(() => sizeClasses[props.size || 'md'])
</script>

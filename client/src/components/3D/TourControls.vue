<template>
  <Transition name="tour-fade">
    <div
      v-if="active"
      class="absolute inset-x-3 bottom-3 z-40 rounded-xl px-3 py-3 text-white shadow-2xl"
      style="background: rgba(36, 31, 27, 0.9); backdrop-filter: blur(16px); border: 1px solid rgba(255,255,255,0.12);"
      aria-live="polite"
    >
      <div class="flex items-center justify-between gap-3">
        <div class="flex min-w-0 items-center gap-2.5">
          <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-mogao-cinnabar text-xs font-semibold">
            {{ index + 1 }}
          </span>
          <div class="min-w-0">
            <p class="whitespace-nowrap text-[9px] tracking-[0.12em] text-white/45">{{ speechSupported ? '3D 语音导览' : '3D 导览' }} · {{ index + 1 }}/{{ stepCount }}</p>
            <p class="truncate text-sm font-medium">
              {{ title }}
              <span v-if="tourSpeaking" class="ml-1 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-mogao-gold align-middle" aria-label="正在讲解" />
              <span v-else-if="tourLoading" class="ml-1 align-middle text-[10px] text-white/45">语音加载中</span>
            </p>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1">
          <button
            @click="emit('previous')"
            :disabled="index === 0"
            class="rounded-full px-2.5 py-1.5 text-[11px] text-white/75 transition-colors hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-30"
          >
            上一站
          </button>
          <button
            @click="emit('next')"
            class="rounded-full bg-mogao-gold px-2.5 py-1.5 text-[11px] font-medium text-mogao-lapis-900 transition-opacity hover:opacity-90"
          >
            {{ isLastStep ? '完成' : '下一站' }}
          </button>
          <button
            v-if="speechSupported"
            @click="emit('toggleVoice')"
            :aria-label="tourVoiceEnabled ? '关闭语音讲解' : '开启语音讲解'"
            :aria-pressed="tourVoiceEnabled"
            :title="tourVoiceEnabled ? '关闭语音讲解' : '开启语音讲解'"
            class="flex h-7 w-7 items-center justify-center rounded-full text-sm text-white/75 transition-colors hover:bg-white/10"
          >
            {{ tourVoiceEnabled ? '🔊' : '🔇' }}
          </button>
          <button
            @click="emit('ask')"
            class="hidden rounded-full px-2.5 py-1.5 text-[11px] text-white/75 transition-colors hover:bg-white/10 sm:inline-flex"
          >
            问织语
          </button>
          <button
            @click="emit('exit')"
            class="flex h-7 w-7 items-center justify-center rounded-full text-white/60 transition-colors hover:bg-white/10 hover:text-white"
            aria-label="退出 3D 语音导览"
          >
            ✕
          </button>
        </div>
      </div>
      <div class="mt-2 h-0.5 overflow-hidden rounded-full bg-white/15">
        <div
          class="h-full rounded-full bg-mogao-gold transition-[width] duration-200"
          :style="{ width: `${((index + 1) / stepCount) * 100}%` }"
        />
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
defineProps<{
  active: boolean;
  title: string;
  index: number;
  stepCount: number;
  isLastStep: boolean;
  speechSupported: boolean;
  tourSpeaking: boolean;
  tourLoading: boolean;
  tourVoiceEnabled: boolean;
}>();

const emit = defineEmits<{
  previous: [];
  next: [];
  toggleVoice: [];
  ask: [];
  exit: [];
}>();
</script>

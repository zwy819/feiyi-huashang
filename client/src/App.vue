<template>
  <router-view v-slot="{ Component }">
    <transition name="page" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup lang="ts">
import { onErrorCaptured, ref, type Ref } from 'vue'

interface AppError {
  error: unknown
  info: string
}

const appError: Ref<AppError | null> = ref(null)

onErrorCaptured((err: unknown, _instance: unknown, info: string) => {
  console.error('[App ErrorBoundary]', err, info)
  appError.value = { error: err, info }
  return false
})
</script>

<style>
.page-enter-active,
.page-leave-active {
  transition: opacity 0.2s ease;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>

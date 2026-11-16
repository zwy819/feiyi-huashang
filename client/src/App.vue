<template>
  <router-view v-slot="{ Component }">
    <transition name="page">
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

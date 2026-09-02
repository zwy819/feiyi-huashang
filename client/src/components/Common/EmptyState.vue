<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
    <span class="text-5xl text-mogao-ochre-300 mb-4">{{ icon }}</span>
    <h3 class="font-serif text-xl font-semibold text-mogao-lapis-700 mb-2">
      {{ title }}
    </h3>
    <p v-if="description" class="text-sm text-mogao-ochre-500 max-w-sm mb-6">{{ description }}</p>
    <router-link v-if="action?.to" :to="action.to" class="btn-mogao-primary">
      {{ action.label }}
    </router-link>
    <button v-else-if="action" @click="action.onClick" class="btn-mogao-primary">
      {{ action.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
interface EmptyAction {
  label: string
  to?: string
  onClick?: () => void
}

defineProps<{
  icon?: string
  title?: string
  description?: string
  action?: EmptyAction | null
}>()
</script>

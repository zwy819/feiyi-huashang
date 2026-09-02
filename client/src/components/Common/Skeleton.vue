<template>
  <template v-for="i in count" :key="i">
    <!-- Card 骨架 -->
    <div v-if="type === 'card'" :class="['card-mogao overflow-hidden', className]">
      <div class="aspect-[4/3] bg-mogao-ochre-100 animate-pulse" />
      <div class="p-4 space-y-2">
        <div class="h-4 w-3/4 bg-mogao-ochre-100 rounded animate-pulse" />
        <div class="h-3 w-1/2 bg-mogao-ochre-100 rounded animate-pulse" />
      </div>
    </div>

    <!-- Text 骨架 -->
    <div v-else-if="type === 'text'" :class="['space-y-2', className]">
      <div class="h-4 w-full bg-mogao-ochre-100 rounded animate-pulse" />
      <div class="h-4 w-5/6 bg-mogao-ochre-100 rounded animate-pulse" />
      <div class="h-4 w-2/3 bg-mogao-ochre-100 rounded animate-pulse" />
    </div>

    <!-- Detail 骨架 -->
    <div v-else-if="type === 'detail'" :class="['space-y-4', className]">
      <div class="aspect-video bg-mogao-ochre-100 rounded-lg animate-pulse" />
      <div class="h-6 w-1/3 bg-mogao-ochre-100 rounded animate-pulse" />
      <div class="space-y-2">
        <div v-for="j in 4" :key="j" class="h-4 w-full bg-mogao-ochre-100 rounded animate-pulse" />
      </div>
    </div>

    <!-- 默认 -->
    <div v-else :class="['h-20 bg-mogao-ochre-100 rounded animate-pulse', className]" />
  </template>
</template>

<script setup lang="ts">
defineProps<{
  type?: 'card' | 'text' | 'detail'
  count?: number
  className?: string
}>()
</script>

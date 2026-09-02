<template>
  <div>
    <h1 class="text-2xl font-bold text-mogao-lapis-900 mb-6">仪表盘</h1>

    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div
        v-for="i in 4"
        :key="i"
        class="bg-white rounded-xl p-6 shadow-sm animate-pulse"
      >
        <div class="h-4 bg-gray-200 rounded w-20 mb-4"></div>
        <div class="h-8 bg-gray-200 rounded w-16"></div>
      </div>
    </div>

    <!-- Stats -->
    <div v-else-if="stats" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div
        v-for="card in statCards"
        :key="card.label"
        class="bg-white rounded-xl p-6 shadow-sm border border-mogao-ochre-100"
      >
        <p class="text-sm text-mogao-lapis-700/60 mb-2">{{ card.label }}</p>
        <p class="text-3xl font-bold" :style="{ color: card.color }">
          {{ card.value }}
        </p>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="bg-white rounded-xl p-8 text-center text-red-500">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { adminDashboardAPI } from '@/services/api'
import type { DashboardStats } from '@/types/statistics'

const loading = ref(true)
const error = ref<string | null>(null)
const stats = ref<DashboardStats | null>(null)

const statCards = computed(() => {
  if (!stats.value) return []
  return [
    {
      label: '服饰总数',
      value: stats.value.total_garments,
      color: '#C43A31',
    },
    {
      label: '模型总数',
      value: stats.value.total_models,
      color: '#2E7D5A',
    },
    {
      label: '用户总数',
      value: stats.value.total_users,
      color: '#2E4A7A',
    },
    {
      label: '收藏总数',
      value: stats.value.total_favorites,
      color: '#B8862D',
    },
  ]
})

onMounted(async () => {
  try {
    const res = await adminDashboardAPI.getStats()
    stats.value = (res as unknown as { data: DashboardStats }).data
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取统计数据失败'
  } finally {
    loading.value = false
  }
})
</script>

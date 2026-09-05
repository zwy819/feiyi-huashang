<template>
  <div class="page-container">
    <div class="mb-8">
      <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">导出报告</h1>
      <p class="mt-2 text-sm text-mogao-ochre-500">将收藏的非遗服饰生成 PDF 报告</p>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="animate-pulse space-y-3">
      <div v-for="i in 3" :key="i" class="h-16 bg-mogao-ochre-100 rounded-lg" />
    </div>

    <!-- 无收藏 -->
    <EmptyState
      v-else-if="favorites.length === 0"
      icon="☷"
      title="暂无收藏"
      description="请先收藏非遗服饰后再生成报告"
      :action="{ label: '去浏览', to: '/explore' }"
    />

    <!-- 报告预览 -->
    <div v-else class="space-y-6">
      <!-- 收藏清单 -->
      <div class="card-mogao p-6">
        <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-4">
          报告内容预览（共 {{ favorites.length }} 件）
        </h3>
        <div class="space-y-2">
          <div
            v-for="(fav, i) in favorites"
            :key="fav.id || fav.garment_id"
            class="flex items-center gap-4 px-4 py-3 bg-mogao-ochre-50 rounded-lg"
          >
            <span class="font-serif text-mogao-cinnabar text-lg w-8 text-center">
              {{ i + 1 }}
            </span>
            <div class="flex-1">
              <p class="font-medium text-mogao-lapis-700">{{ fav.garment_name }}</p>
              <p class="text-xs text-mogao-ochre-500">
                {{ dynastyLabel(fav.garment_dynasty) }} · {{ fav.garment_ethnic || '' }} · {{ fav.garment_craft || '' }} · {{ fav.garment_province || '' }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 报告说明 -->
      <div class="card-mogao p-6">
        <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-3">报告说明</h3>
        <ul class="text-sm text-mogao-lapis-600 space-y-2 list-disc list-inside">
          <li>报告格式：PDF</li>
          <li>包含内容：封面、每件服饰的详细信息（名称、朝代、民族、工艺、材质、纹样、文化寓意）</li>
          <li>生成日期：{{ formatDate(new Date().toISOString()) }}</li>
        </ul>
      </div>

      <!-- 生成按钮 -->
      <div class="text-center">
        <button
          @click="handleGenerate"
          :disabled="generating"
          class="btn-mogao-primary text-lg px-10 py-3 disabled:opacity-60"
        >
          <span v-if="generating" class="flex items-center gap-2">
            <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            生成中...
          </span>
          <span v-else>生成并下载 PDF 报告</span>
        </button>
        <p v-if="generated" class="mt-3 text-sm text-mogao-malachite-600">
          报告已生成，若未自动下载请检查浏览器下载设置
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { reportAPI, favoriteAPI } from '@/services/api'
import EmptyState from '@/components/Common/EmptyState.vue'
import { formatDate } from '@/utils'
import type { ApiResponse } from '@/types/api'
import type { FavoriteItem } from '@/types/favorite'

const favorites: Ref<FavoriteItem[]> = ref([])
const loading: Ref<boolean> = ref(true)
const generating: Ref<boolean> = ref(false)
const generated: Ref<boolean> = ref(false)

// "清"→"清代"；"现代/古代"这类已含"代"的原样保留
function dynastyLabel(dynasty?: string): string {
  if (!dynasty) return ''
  return dynasty.endsWith('代') ? dynasty : `${dynasty}代`
}

onMounted(async () => {
  try {
    const res = await favoriteAPI.list()
    const data = res as ApiResponse<{ favorites: FavoriteItem[] }> | { favorites: FavoriteItem[] }
    favorites.value = (data as ApiResponse<{ favorites: FavoriteItem[] }>).data?.favorites
      || (data as { favorites: FavoriteItem[] }).favorites
      || []
  } catch {
    favorites.value = []
  } finally {
    loading.value = false
  }
})

async function handleGenerate(): Promise<void> {
  generating.value = true
  try {
    const response = (await reportAPI.generate({
      ids: favorites.value.map((f) => f.garment_id).join(','),
    })) as unknown as Blob

    const blob = new Blob([response], { type: 'application/pdf' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `非遗服饰收藏报告_${new Date().toISOString().slice(0, 10)}.pdf`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    generated.value = true
  } catch (err: unknown) {
    console.error('报告生成失败:', err)
  } finally {
    generating.value = false
  }
}
</script>

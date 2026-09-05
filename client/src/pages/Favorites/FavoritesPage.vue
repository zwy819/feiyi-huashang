<template>
  <div class="page-container">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">我的收藏</h1>
        <p class="mt-2 text-sm text-mogao-ochre-500">
          共收藏 {{ favorites.length }} 件非遗服饰
        </p>
      </div>
      <router-link v-if="favorites.length > 0" to="/report" class="btn-mogao-primary text-sm">
        导出报告
      </router-link>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      <Skeleton type="card" :count="4" />
    </div>

    <!-- 空状态 -->
    <EmptyState
      v-else-if="favorites.length === 0"
      icon="♡"
      title="收藏夹为空"
      description="浏览服饰库，收藏感兴趣的非遗服饰"
      :action="{ label: '去浏览', to: '/explore' }"
    />

    <!-- 收藏列表 -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      <TransitionGroup name="fade-list">
        <div
          v-for="fav in favorites"
          :key="fav.id || fav.garment_id"
          class="relative group"
        >
          <Card
            :to="`/detail/${fav.garment_id}`"
            :title="fav.garment_name"
            :subtitle="`${dynastyLabel(fav.garment_dynasty)} · ${fav.garment_ethnic || ''} · ${fav.garment_craft || ''}`"
          />
          <!-- 移除按钮 -->
          <button
            @click.prevent.stop="handleRemove(fav.garment_id)"
            :disabled="removing === fav.garment_id"
            class="absolute top-2 right-2 p-1.5 rounded-full bg-white/90 text-mogao-cinnabar
                   opacity-0 group-hover:opacity-100 transition-opacity
                   hover:bg-mogao-cinnabar hover:text-white disabled:opacity-50"
            aria-label="移除收藏"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { favoriteAPI } from '@/services/api'
import Card from '@/components/Common/Card.vue'
import EmptyState from '@/components/Common/EmptyState.vue'
import Skeleton from '@/components/Common/Skeleton.vue'
import type { FavoriteItem } from '@/types/favorite'
import type { ApiResponse } from '@/types/api'

const favorites: Ref<FavoriteItem[]> = ref([])
const loading: Ref<boolean> = ref(true)
const removing: Ref<number | null> = ref(null)

// "清"→"清代"；"现代/古代"这类已含"代"的原样保留
function dynastyLabel(dynasty?: string): string {
  if (!dynasty) return ''
  return dynasty.endsWith('代') ? dynasty : `${dynasty}代`
}

async function fetchFavorites(): Promise<void> {
  loading.value = true
  try {
    const result = await favoriteAPI.list()
    const data = result as ApiResponse<{ favorites: FavoriteItem[] }> | { favorites: FavoriteItem[] }
    favorites.value = (data as ApiResponse<{ favorites: FavoriteItem[] }>).data?.favorites
      || (data as { favorites: FavoriteItem[] }).favorites
      || []
  } catch {
    favorites.value = []
  } finally {
    loading.value = false
  }
}

async function handleRemove(garmentId: number | undefined): Promise<void> {
  if (!garmentId) return
  removing.value = garmentId
  try {
    await favoriteAPI.remove(garmentId)
    favorites.value = favorites.value.filter((f) => Number(f.garment_id) !== Number(garmentId))
  } catch {
    console.error('移除收藏失败')
  } finally {
    removing.value = null
  }
}

onMounted(fetchFavorites)
</script>

<style scoped>
.fade-list-enter-active,
.fade-list-leave-active {
  transition: all 0.3s ease;
}
.fade-list-enter-from {
  opacity: 0;
  transform: scale(0.9);
}
.fade-list-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>

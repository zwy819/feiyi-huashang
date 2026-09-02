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
            :to="`/detail/${fav.garment?.id || fav.garment_id}`"
            :title="fav.garment?.name || fav.name"
            :subtitle="`${fav.garment?.dynasty || fav.dynasty || ''}代 · ${fav.garment?.ethnic || fav.ethnic || ''} · ${fav.garment?.craft || fav.craft || ''}`"
          />
          <!-- 移除按钮 -->
          <button
            @click.prevent.stop="handleRemove(fav.garment_id || fav.garment?.id)"
            :disabled="removing === (fav.garment_id || fav.garment?.id)"
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

interface FavoriteDisplay {
  id?: number
  garment_id?: number
  garment?: {
    id?: number
    name?: string
    dynasty?: string
    ethnic?: string
    craft?: string
    province?: string
  }
  name?: string
  dynasty?: string
  ethnic?: string
  craft?: string
  province?: string
}

const favorites: Ref<FavoriteDisplay[]> = ref([])
const loading: Ref<boolean> = ref(true)
const removing: Ref<number | null> = ref(null)

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
    favorites.value = favorites.value.filter(
      (f) => Number(f.garment_id || f.garment?.id) !== Number(garmentId),
    )
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

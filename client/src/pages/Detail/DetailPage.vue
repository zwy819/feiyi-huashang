<template>
  <div class="page-container">
    <!-- 面包屑导航 -->
    <nav class="flex items-center gap-2 text-xs text-mogao-ochre-500 mb-6">
      <router-link to="/" class="hover:text-mogao-cinnabar">首页</router-link>
      <span>/</span>
      <router-link to="/explore" class="hover:text-mogao-cinnabar">服饰浏览</router-link>
      <span>/</span>
      <span class="text-mogao-lapis-700">{{ garment?.name }}</span>
    </nav>

    <!-- 加载中 -->
    <Skeleton v-if="loading" type="detail" />

    <!-- 未找到 -->
    <div v-else-if="!garment" class="text-center py-20">
      <p class="text-mogao-cinnabar">服饰信息未找到</p>
      <router-link to="/explore" class="btn-mogao-primary mt-4 inline-block">返回浏览</router-link>
    </div>

    <!-- 详情内容 -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-5 gap-8">
      <!-- 左侧：3D 预览 -->
      <div class="lg:col-span-3">
        <div class="card-mogao overflow-hidden">
          <ModelViewer
            v-if="modelPath"
            :model-path="modelPath"
            :auto-rotate="true"
            :height="480"
          />
          <div v-else class="aspect-video bg-gradient-to-br from-mogao-lapis-50 to-mogao-ochre-50 flex items-center justify-center">
            <div class="text-center">
              <span class="text-6xl block mb-3">&#x1f457;</span>
              <p class="text-sm text-mogao-ochre-500">3D 服饰展示</p>
              <p class="text-xs text-mogao-ochre-400 mt-1">暂无 3D 模型</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：信息卡片 -->
      <div class="lg:col-span-2 space-y-5">
        <!-- 标题 + 收藏 -->
        <div class="card-mogao p-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h1 class="font-serif text-2xl font-bold text-mogao-lapis-800">{{ garment.name }}</h1>
              <p class="text-sm text-mogao-ochre-500 mt-1">
                {{ garment.dynasty }}代 · {{ garment.ethnic }} · {{ garment.province }}
              </p>
            </div>
            <button
              @click="toggleFavorite"
              :class="[
                'p-2 rounded-full transition-colors',
                isFavored
                  ? 'text-mogao-cinnabar bg-mogao-cinnabar-50'
                  : 'text-mogao-ochre-400 hover:text-mogao-cinnabar hover:bg-mogao-cinnabar-50',
              ]"
              :aria-label="isFavored ? '取消收藏' : '添加收藏'"
            >
              <svg class="w-6 h-6" :fill="isFavored ? 'currentColor' : 'none'" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
            </button>
          </div>

          <!-- 属性列表 -->
          <dl class="grid grid-cols-2 gap-3 text-sm">
            <div v-for="[label, value] in attributes" :key="label">
              <dt class="text-mogao-ochre-500 text-xs">{{ label }}</dt>
              <dd class="text-mogao-lapis-700 font-medium mt-0.5">{{ value || '-' }}</dd>
            </div>
          </dl>
        </div>

        <!-- 文化寓意 -->
        <div class="card-mogao p-6">
          <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-3">文化寓意</h3>
          <p class="text-sm text-mogao-lapis-600 leading-relaxed">
            {{ garment.cultural_meaning || '暂无文化寓意信息' }}
          </p>
        </div>
      </div>
    </div>

    <!-- 智能推荐链 -->
    <div v-if="recommends.length > 0" class="mt-12">
      <h2 class="title-mogao font-serif text-xl font-bold text-mogao-lapis-800 mb-6">
        智能推荐链
      </h2>
      <p class="text-xs text-mogao-ochre-500 mb-4">
        基于工艺、民族、纹样的相似性，推荐以下相关服饰
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <router-link
          v-for="rec in recommends"
          :key="rec.id"
          :to="`/detail/${rec.id}`"
          class="card-mogao p-4 block hover:shadow-mogao-lg transition-shadow"
        >
          <h4 class="font-serif font-semibold text-mogao-lapis-700">{{ rec.name }}</h4>
          <p class="text-xs text-mogao-ochre-500 mt-1">{{ rec.similarity }}</p>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, type Ref, type ComputedRef } from 'vue'
import { useRoute, useRouter, type RouteLocationNormalizedLoaded, type Router } from 'vue-router'
import { garmentAPI, favoriteAPI } from '@/services/api'
import ModelViewer from '@/components/3D/ModelViewer.vue'
import Skeleton from '@/components/Common/Skeleton.vue'
import type { Garment, RecommendItem } from '@/types/garment'
import type { ApiResponse } from '@/types/api'
import type { FavoriteCheckResult } from '@/types/favorite'

const route: RouteLocationNormalizedLoaded = useRoute()
const router: Router = useRouter()
const garment: Ref<Garment | null> = ref(null)
const recommends: Ref<RecommendItem[]> = ref([])
const loading: Ref<boolean> = ref(true)
const isFavored: Ref<boolean> = ref(false)

const attributes: ComputedRef<[string, string][]> = computed(() => {
  if (!garment.value) return []
  return [
    ['工艺', garment.value.craft],
    ['材质', garment.value.material],
    ['纹样', garment.value.pattern],
    ['省份', garment.value.province],
    ['朝代', garment.value.dynasty + '代'],
    ['民族', garment.value.ethnic],
  ]
})

const modelPath = computed(() => {
  if (!garment.value) return ''
  if (garment.value.model_id) return `/api/models/${garment.value.model_id}/file`
  if (garment.value.model_url) return garment.value.model_url
  return ''
})

async function loadDetail(id: string | number): Promise<void> {
  loading.value = true
  try {
    const requests: Promise<unknown>[] = [
      garmentAPI.getById(Number(id)),
      garmentAPI.getRecommendChain(Number(id)),
    ]
    if (localStorage.getItem('token')) {
      requests.push(
        favoriteAPI.check(Number(id)).catch(() => ({ isFavorited: false })),
      )
    }
    const [detailRes, recRes, favRes] = await Promise.all(requests)
    const detailData = detailRes as ApiResponse<Garment> | Garment
    garment.value = (detailData as ApiResponse<Garment>).data || (detailData as Garment)

    const recData = recRes as ApiResponse<{ items: RecommendItem[] }> | { items: RecommendItem[] }
    recommends.value = (recData as ApiResponse<{ items: RecommendItem[] }>).data?.items || (recData as { items: RecommendItem[] }).items || []

    const favData = favRes as FavoriteCheckResult | ApiResponse<FavoriteCheckResult> | undefined
    isFavored.value = favData
      ? (favData as ApiResponse<FavoriteCheckResult>).data?.isFavorited || (favData as FavoriteCheckResult).isFavorited
      : false
  } catch {
    garment.value = {
      id: Number(id),
      name: '苗族百褶裙',
      dynasty: '清',
      ethnic: '苗族',
      craft: '蜡染 + 刺绣',
      material: '棉麻混纺',
      pattern: '几何纹 + 花卉纹',
      cultural_meaning: '百褶裙是苗族女性传统服饰，裙褶象征苗族祖先迁徙途中翻越的山川；几何纹代表星空与大地，花卉纹寓意生命繁衍，整体承载着苗族的迁徙历史与自然崇拜。',
      province: '贵州',
      description: '苗族百褶裙以蜡染为底、刺绣为饰，制作需经纺纱、织布、蜡染、刺绣、压褶等多道工序。',
      image_url: '',
      model_url: '',
      model_id: null,
      created_at: '',
      updated_at: '',
    }
    recommends.value = [
      { id: 2, name: '苏绣旗袍', similarity: 0 },
      { id: 3, name: '藏族氆氇袍', similarity: 0 },
      { id: 4, name: '蒙古族长袍', similarity: 0 },
    ]
  } finally {
    loading.value = false
  }
}

onMounted(() => loadDetail(route.params.id as string))
watch(
  () => route.params.id,
  (id) => loadDetail(id as string),
)

async function toggleFavorite(): Promise<void> {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }
  const id = Number(route.params.id)
  try {
    if (isFavored.value) {
      await favoriteAPI.remove(id)
    } else {
      await favoriteAPI.add(id)
    }
    isFavored.value = !isFavored.value
  } catch (err: unknown) {
    console.error('收藏操作失败:', err)
  }
}
</script>

<template>
  <div>
    <!-- 首屏 Hero -->
    <section class="relative overflow-hidden bg-gradient-to-b from-mogao-lapis-900 via-mogao-lapis-800 to-mogao-lapis-700 text-white">
      <!-- 装饰底纹 -->
      <div class="absolute inset-0 opacity-5">
        <div class="absolute top-10 left-10 w-64 h-64 rounded-full bg-mogao-gold blur-3xl" />
        <div class="absolute bottom-10 right-10 w-96 h-96 rounded-full bg-mogao-cinnabar blur-3xl" />
      </div>

      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 md:py-28">
        <div class="text-center">
          <h1 class="font-serif text-4xl md:text-6xl font-bold tracking-wide mb-4">
            非遗华裳
          </h1>
          <p class="text-lg md:text-xl text-mogao-leadwhite-300 mb-8 max-w-2xl mx-auto leading-relaxed">
            以数字技术重现千年织造智慧，探索中国非物质文化遗产服饰的纹样之美、工艺之精、文化之魂
          </p>
          <div class="flex flex-wrap justify-center gap-4">
            <router-link to="/explore" class="btn-mogao-primary !bg-mogao-gold !text-mogao-lapis-900 hover:!bg-mogao-gold-400">
              浏览服饰库
            </router-link>
            <router-link to="/statistics" class="btn-mogao-secondary !border-mogao-leadwhite-300 !text-mogao-leadwhite-100 hover:!bg-white/10">
              查看地理分布
            </router-link>
          </div>
        </div>
      </div>
    </section>

    <!-- 每日一服 -->
    <section class="page-container">
      <div class="text-center mb-10">
        <h2 class="title-mogao font-serif text-2xl md:text-3xl font-bold text-mogao-lapis-800">
          每日一服
        </h2>
        <p class="mt-3 text-sm text-mogao-ochre-600">
          每天一件非遗服饰，走近中华衣冠之美
        </p>
        <p class="text-xs text-mogao-ochre-400 mt-1">
          {{ formatDate(new Date().toISOString()) }}
        </p>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="max-w-2xl mx-auto animate-pulse">
        <div class="aspect-video bg-mogao-ochre-100 rounded-lg" />
        <div class="mt-4 h-6 w-1/3 bg-mogao-ochre-100 rounded" />
      </div>

      <!-- 每日一服卡片 -->
      <div v-else-if="dailyGarment" class="max-w-2xl mx-auto card-mogao overflow-hidden">
        <!-- 3D 视口 -->
        <ModelViewer
          v-if="modelPath"
          :model-path="modelPath"
          :auto-rotate="true"
          :height="400"
        />
        <div v-else class="aspect-video bg-gradient-to-br from-mogao-lapis-50 to-mogao-ochre-50 flex items-center justify-center">
          <div class="text-center">
            <span class="text-6xl block mb-3">&#x1f457;</span>
            <p class="text-sm text-mogao-ochre-500">3D 服饰模型预览区</p>
            <p class="text-xs text-mogao-ochre-400 mt-1">暂无 3D 模型</p>
          </div>
        </div>

        <div class="p-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <h3 class="font-serif text-2xl font-bold text-mogao-lapis-800">
                {{ dailyGarment.name }}
              </h3>
              <p class="text-sm text-mogao-ochre-500 mt-1">
                {{ dailyGarment.dynasty }}代 · {{ dailyGarment.ethnic }} · {{ dailyGarment.province }}
              </p>
            </div>
            <router-link
              :to="`/detail/${dailyGarment.id}`"
              class="btn-mogao-secondary text-sm"
            >
              查看详情
            </router-link>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
            <div class="bg-mogao-ochre-50 p-3 rounded">
              <span class="text-mogao-ochre-500 text-xs">工艺</span>
              <p class="text-mogao-lapis-700 font-medium mt-0.5">{{ dailyGarment.craft }}</p>
            </div>
            <div class="bg-mogao-ochre-50 p-3 rounded">
              <span class="text-mogao-ochre-500 text-xs">材质</span>
              <p class="text-mogao-lapis-700 font-medium mt-0.5">{{ dailyGarment.material }}</p>
            </div>
            <div class="bg-mogao-ochre-50 p-3 rounded">
              <span class="text-mogao-ochre-500 text-xs">纹样</span>
              <p class="text-mogao-lapis-700 font-medium mt-0.5">{{ dailyGarment.pattern }}</p>
            </div>
            <div class="bg-mogao-ochre-50 p-3 rounded">
              <span class="text-mogao-ochre-500 text-xs">寓意</span>
              <p class="text-mogao-lapis-700 font-medium mt-0.5 line-clamp-2">{{ dailyGarment.cultural_meaning }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 用户贡献 -->
    <UserContributions />

    <!-- 模块入口 -->
    <section class="bg-mogao-leadwhite-100 py-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h2 class="text-center title-mogao font-serif text-2xl font-bold text-mogao-lapis-800 mb-10">
          探索非遗服饰
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <router-link
            v-for="item in exploreModules"
            :key="item.to"
            :to="item.to"
            class="block group"
          >
            <div class="card-mogao p-6 text-center h-full hover:shadow-mogao-lg transition-shadow">
              <span :class="['text-4xl block mb-3', `text-mogao-${item.color}`]">{{ item.icon }}</span>
              <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-2 group-hover:text-mogao-cinnabar transition-colors">
                {{ item.title }}
              </h3>
              <p class="text-sm text-mogao-ochre-500">{{ item.desc }}</p>
            </div>
          </router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, type Ref } from 'vue'
import { garmentAPI } from '@/services/api'
import ModelViewer from '@/components/3D/ModelViewer.vue'
import UserContributions from '@/components/UserContributions/UserContributions.vue'
import { formatDate } from '@/utils'
import type { Garment } from '@/types/garment'
import type { ApiResponse } from '@/types/api'

interface ExploreModule {
  icon: string
  title: string
  desc: string
  to: string
  color: string
}

const dailyGarment: Ref<Garment | null> = ref(null)
const loading: Ref<boolean> = ref(true)

const modelPath = computed(() => {
  if (!dailyGarment.value) return ''
  if (dailyGarment.value.model_id) return `/api/models/${dailyGarment.value.model_id}/file`
  if (dailyGarment.value.model_url) return dailyGarment.value.model_url
  return ''
})

const exploreModules: ExploreModule[] = [
  { icon: '☷', title: '服饰浏览', desc: '按朝代、民族、工艺筛选', to: '/explore', color: 'cinnabar' },
  { icon: '◎', title: '地理分布', desc: '查看全国非遗服饰分布', to: '/statistics', color: 'malachite' },
  { icon: '🎯', title: '华裳问答', desc: '答题闯关涨知识赢积分', to: '/quiz', color: 'gold' },
  { icon: '♡', title: '我的收藏', desc: '收藏喜欢的服饰', to: '/favorites', color: 'lapis' },
]

onMounted(async () => {
  try {
    const res = await garmentAPI.getDaily()
    const data = (res as ApiResponse<Garment>).data || (res as unknown as Garment)
    dailyGarment.value = data
  } catch {
    dailyGarment.value = {
      id: 1,
      name: '苗族百褶裙',
      dynasty: '清',
      ethnic: '苗族',
      craft: '蜡染 + 刺绣',
      material: '棉麻混纺',
      pattern: '几何纹 + 花卉纹',
      cultural_meaning: '象征苗族的迁徙历史与自然崇拜',
      province: '贵州',
      description: '',
      image_url: '',
      model_url: '',
      model_id: null,
      created_at: '',
      updated_at: '',
    }
  } finally {
    loading.value = false
  }
})
</script>

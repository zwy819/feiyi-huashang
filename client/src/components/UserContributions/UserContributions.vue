<template>
  <section v-if="items.length > 0" class="bg-white py-16">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="text-center mb-10">
        <h2 class="font-serif text-2xl md:text-3xl font-bold text-mogao-lapis-800">
          用户贡献
        </h2>
        <p class="mt-3 text-sm text-mogao-ochre-600">
          来自社区用户的非遗服饰投稿，每一件都是珍贵的文化记忆
        </p>
        <router-link to="/submit" class="inline-block mt-3 text-sm text-mogao-cinnabar hover:underline">
          我也要投稿 →
        </router-link>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="flex justify-center">
        <div class="flex gap-4 overflow-hidden">
          <div
            v-for="n in 4"
            :key="n"
            class="w-56 h-48 bg-mogao-ochre-100 rounded-lg animate-pulse shrink-0"
          />
        </div>
      </div>

      <!-- 卡片轮播网格 -->
      <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        <div
          v-for="sub in items"
          :key="sub.id"
          class="group bg-white rounded-xl border border-mogao-ochre-100 overflow-hidden hover:shadow-md transition-shadow"
        >
          <!-- 图片 -->
          <div class="aspect-[4/3] bg-mogao-ochre-100 overflow-hidden">
            <img
              :src="`/api/submissions/${sub.id}/image`"
              :alt="sub.name"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              loading="lazy"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
          </div>

          <!-- 信息 -->
          <div class="p-3 space-y-1">
            <h3 class="text-sm font-medium text-mogao-lapis-800 truncate">{{ sub.name }}</h3>
            <div class="text-xs text-mogao-lapis-700/50 flex items-center justify-between">
              <span v-if="sub.dynasty || sub.ethnic">
                <span v-if="sub.dynasty">{{ sub.dynasty }}代</span>
                <span v-if="sub.dynasty && sub.ethnic"> · </span>
                <span v-if="sub.ethnic">{{ sub.ethnic }}</span>
              </span>
              <span v-else>{{ sub.username || '匿名用户' }}</span>

              <router-link
                v-if="sub.created_garment_id"
                :to="`/detail/${sub.created_garment_id}`"
                class="text-mogao-cinnabar hover:underline shrink-0"
              >
                查看详情
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { submissionAPI } from '@/services/api'
import type { Submission } from '@/types/submission'

const items = ref<Submission[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await submissionAPI.listPublic({ page: 1, limit: 8 })
    const data = (res as unknown as { data: { items: Submission[] } }).data
    items.value = data.items
  } catch {
    // silently fail — user contributions are optional
  } finally {
    loading.value = false
  }
})
</script>

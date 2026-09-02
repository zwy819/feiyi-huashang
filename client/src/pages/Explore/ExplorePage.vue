<template>
  <div class="page-container">
    <!-- 页面标题 -->
    <div class="mb-8">
      <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">服饰浏览</h1>
      <p class="mt-2 text-sm text-mogao-ochre-500">浏览非遗服饰数据库，探索千年衣冠之美</p>
    </div>

    <!-- 搜索 + 筛选栏 -->
    <div class="card-mogao p-4 mb-8">
      <!-- 搜索框 -->
      <div class="mb-4">
        <input
          v-model="search"
          type="text"
          placeholder="搜索服饰名称..."
          class="w-full px-4 py-2.5 border border-mogao-ochre-200 rounded-mogao text-sm
                 focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                 placeholder:text-mogao-ochre-400"
        />
      </div>

      <!-- 筛选行 -->
      <div class="flex flex-wrap gap-3">
        <select v-model="dynasty" class="px-3 py-2 text-sm border border-mogao-ochre-200 rounded-mogao bg-white focus:outline-none focus:border-mogao-cinnabar text-mogao-lapis-700">
          <option v-for="opt in DYNASTY_OPTIONS" :key="opt" :value="opt">
            {{ opt || '全部朝代' }}
          </option>
        </select>
        <select v-model="ethnic" class="px-3 py-2 text-sm border border-mogao-ochre-200 rounded-mogao bg-white focus:outline-none focus:border-mogao-cinnabar text-mogao-lapis-700">
          <option v-for="opt in ETHNIC_OPTIONS" :key="opt" :value="opt">
            {{ opt || '全部民族' }}
          </option>
        </select>
        <select v-model="craft" class="px-3 py-2 text-sm border border-mogao-ochre-200 rounded-mogao bg-white focus:outline-none focus:border-mogao-cinnabar text-mogao-lapis-700">
          <option v-for="opt in CRAFT_OPTIONS" :key="opt" :value="opt">
            {{ opt || '全部工艺' }}
          </option>
        </select>
      </div>
    </div>

    <!-- 服饰网格 -->
    <template v-if="loading">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        <Skeleton type="card" :count="8" />
      </div>
    </template>

    <template v-else-if="garments.length === 0">
      <EmptyState
        icon="☷"
        title="未找到匹配的服饰"
        description="尝试调整搜索关键词或筛选条件"
        :action="{ label: '清除筛选', onClick: clearFilters }"
      />
    </template>

    <template v-else>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        <Card
          v-for="g in garments"
          :key="g.id"
          :to="`/detail/${g.id}`"
          :title="g.name"
          :subtitle="`${g.dynasty}代 · ${g.ethnic} · ${g.craft}`"
        >
          <div class="flex flex-wrap gap-1.5">
            <span class="px-2 py-0.5 text-xs rounded-full bg-mogao-cinnabar-50 text-mogao-cinnabar-700">
              {{ g.ethnic }}
            </span>
            <span class="px-2 py-0.5 text-xs rounded-full bg-mogao-malachite-50 text-mogao-malachite-700">
              {{ g.craft }}
            </span>
            <span class="px-2 py-0.5 text-xs rounded-full bg-mogao-lapis-50 text-mogao-lapis-700">
              {{ g.province }}
            </span>
          </div>
        </Card>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 mt-10">
        <button
          @click="page = Math.max(1, page - 1)"
          :disabled="page === 1"
          class="btn-mogao-secondary text-sm disabled:opacity-40"
        >
          上一页
        </button>
        <span class="px-4 text-sm text-mogao-ochre-600">
          {{ page }} / {{ totalPages }}
        </span>
        <button
          @click="page = Math.min(totalPages, page + 1)"
          :disabled="page === totalPages"
          class="btn-mogao-secondary text-sm disabled:opacity-40"
        >
          下一页
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, type Ref, type ComputedRef } from 'vue'
import { garmentAPI } from '@/services/api'
import { useDebounce } from '@/composables'
import Card from '@/components/Common/Card.vue'
import Skeleton from '@/components/Common/Skeleton.vue'
import EmptyState from '@/components/Common/EmptyState.vue'
import type { GarmentListItem } from '@/types/garment'
import type { ApiResponse, PaginatedData } from '@/types/api'

const DYNASTY_OPTIONS: string[] = ['', '唐', '宋', '元', '明', '清', '现代']
const ETHNIC_OPTIONS: string[] = ['', '汉族', '苗族', '藏族', '蒙古族', '傣族', '维吾尔族', '壮族', '彝族', '满族', '回族']
const CRAFT_OPTIONS: string[] = ['', '刺绣', '蜡染', '织锦', '缂丝', '扎染', '挑花', '印花']

const garments: Ref<GarmentListItem[]> = ref([])
const loading: Ref<boolean> = ref(true)
const page: Ref<number> = ref(1)
const total: Ref<number> = ref(0)

const search: Ref<string> = ref('')
const dynasty: Ref<string> = ref('')
const ethnic: Ref<string> = ref('')
const craft: Ref<string> = ref('')

const debouncedSearch: Ref<string> = useDebounce(search, 400)

const totalPages: ComputedRef<number> = computed(() => Math.ceil(total.value / 12))

function clearFilters(): void {
  search.value = ''
  dynasty.value = ''
  ethnic.value = ''
  craft.value = ''
}

async function fetchGarments(): Promise<void> {
  loading.value = true
  try {
    const result = await garmentAPI.list({
      page: page.value,
      limit: 12,
      search: debouncedSearch.value || undefined,
      dynasty: dynasty.value || undefined,
      ethnic: ethnic.value || undefined,
      craft: craft.value || undefined,
    })
    const data = (result as ApiResponse<PaginatedData<GarmentListItem>>).data || (result as unknown as PaginatedData<GarmentListItem>)
    garments.value = data.items || []
    total.value = data.total || 0
  } catch {
    garments.value = [
      { id: 1, name: '苗族百褶裙', dynasty: '清', ethnic: '苗族', craft: '蜡染 + 刺绣', province: '贵州', model_id: null },
      { id: 2, name: '苏绣旗袍', dynasty: '现代', ethnic: '汉族', craft: '苏绣', province: '江苏', model_id: null },
      { id: 3, name: '藏族氆氇袍', dynasty: '清', ethnic: '藏族', craft: '氆氇织造', province: '西藏', model_id: null },
      { id: 4, name: '蒙古族长袍', dynasty: '元', ethnic: '蒙古族', craft: '镶边刺绣', province: '内蒙古', model_id: null },
      { id: 5, name: '傣族筒裙', dynasty: '清', ethnic: '傣族', craft: '织锦', province: '云南', model_id: null },
    ]
    total.value = 5
  } finally {
    loading.value = false
  }
}

// 筛选条件变化时重置分页
watch([debouncedSearch, dynasty, ethnic, craft], () => {
  page.value = 1
})

// 页码或筛选条件变化时重新获取
watch([page, debouncedSearch, dynasty, ethnic, craft], () => {
  fetchGarments()
})

onMounted(() => {
  fetchGarments()
})
</script>

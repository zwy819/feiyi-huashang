<template>
  <div class="page-container">
    <!-- 页面标题 -->
    <div class="mb-8 flex items-start justify-between gap-4">
      <div>
        <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">服饰浏览</h1>
        <p class="mt-2 text-sm text-mogao-ochre-500">浏览非遗服饰数据库，探索千年衣冠之美</p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button
          type="button"
          @click="pickPhoto"
          class="btn-mogao-secondary flex items-center gap-1.5 text-sm"
        >
          <span aria-hidden="true">📷</span>
          拍照识衣
        </button>
        <button
          type="button"
          @click="openRoute"
          class="btn-mogao-primary flex items-center gap-1.5 text-sm"
        >
          <span aria-hidden="true">◈</span>
          一句话路线
        </button>
      </div>
    </div>
    <input ref="photoInput" type="file" accept="image/*" class="hidden" @change="onPhotoChange" />

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

    <Teleport to="body">
      <div
        v-if="photoOpen"
        class="fixed inset-0 z-[70] flex items-center justify-center bg-mogao-lapis-900/40 px-4 py-6 backdrop-blur-[2px]"
        @click.self="closePhoto"
      >
        <section class="max-h-[88vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-mogao-leadwhite-50 p-5 shadow-2xl">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] tracking-[0.2em] text-mogao-gold-600">PHOTO TO COLLECTION</p>
              <h2 class="mt-1 font-serif text-2xl font-bold text-mogao-lapis-800">拍照识衣</h2>
              <p class="mt-1 text-xs text-mogao-ochre-500">上传服饰照片，先识别民族/工艺/纹样，再从馆藏中匹配 Top 5。</p>
            </div>
            <button class="rounded-full p-2 text-mogao-ochre-500 hover:bg-mogao-ochre-100" aria-label="关闭拍照识衣" @click="closePhoto">✕</button>
          </div>

          <div class="mt-4 grid gap-4 md:grid-cols-[240px_1fr]">
            <figure class="overflow-hidden rounded-xl border border-mogao-ochre-200 bg-white">
              <img v-if="photoPreview" :src="photoPreview" alt="待识别服饰照片" class="h-64 w-full object-contain" />
              <figcaption class="border-t border-mogao-ochre-100 px-3 py-2 text-[11px] text-mogao-ochre-500">
                {{ photoLoading ? '正在识别，请稍候…' : '已上传图片' }}
              </figcaption>
            </figure>

            <div class="min-h-48 rounded-xl border border-mogao-ochre-200 bg-white p-4">
              <div v-if="photoLoading" class="flex h-full items-center justify-center text-sm text-mogao-ochre-500">
                正在调用视觉模型识别…
              </div>
              <div v-else-if="photoError" class="text-sm leading-relaxed text-mogao-cinnabar-700">{{ photoError }}</div>
              <template v-else-if="visionResult">
                <div class="flex flex-wrap gap-1.5">
                  <span v-for="item in visionAttributes" :key="item.label" class="rounded-full bg-mogao-cinnabar-50 px-2.5 py-1 text-xs text-mogao-cinnabar-700">
                    {{ item.label }}：{{ item.value }}
                  </span>
                  <span class="rounded-full bg-mogao-gold-50 px-2.5 py-1 text-xs text-mogao-gold-700">
                    置信度 {{ Math.round((visionResult.attributes?.confidence || 0) * 100) }}%
                  </span>
                </div>
                <p class="mt-3 text-xs leading-relaxed text-mogao-lapis-600">
                  {{ visionResult.attributes?.description || '视觉模型未给出描述。' }}
                </p>
                <p class="mt-2 text-[11px] text-mogao-ochre-500">
                  识别用时 {{ visionResult.latency_ms }}ms · 候选来自馆藏库，不是直接由模型生成
                </p>
              </template>
              <p v-else class="text-sm text-mogao-ochre-500">选择图片后开始识别。</p>
            </div>
          </div>

          <div v-if="visionResult && visionResult.candidates?.length" class="mt-5">
            <h3 class="text-sm font-semibold text-mogao-lapis-800">馆藏候选 Top {{ visionResult.candidates.length }}</h3>
            <div class="mt-3 grid gap-2 sm:grid-cols-2">
              <button
                v-for="item in visionResult.candidates"
                :key="item.garment_id"
                type="button"
                class="rounded-xl border border-mogao-ochre-200 bg-white p-3 text-left transition-colors hover:border-mogao-gold-400 hover:bg-mogao-gold-50/40"
                @click="openCandidate(item.garment_id)"
              >
                <div class="flex items-start justify-between gap-2">
                  <span class="font-serif text-sm font-semibold text-mogao-lapis-800">{{ item.name }}</span>
                  <span class="text-[11px] text-mogao-gold-700">{{ Math.round(item.score * 10) / 10 }} 分</span>
                </div>
                <p class="mt-1 text-[11px] text-mogao-ochre-500">{{ item.dynasty }}代 · {{ item.ethnic }} · {{ item.craft }} · {{ item.province }}</p>
                <p v-if="item.reason" class="mt-1 line-clamp-2 text-[11px] text-mogao-lapis-500">{{ item.reason }}</p>
              </button>
            </div>
          </div>
          <p v-else-if="visionResult" class="mt-4 rounded-lg bg-mogao-ochre-50 px-3 py-2 text-xs text-mogao-ochre-600">
            没有找到足够接近的馆藏，请换一张更清晰的正面照片，或尝试“服饰浏览”里的关键词搜索。
          </p>
        </section>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="routeOpen"
        class="fixed inset-0 z-[70] flex items-center justify-center bg-mogao-lapis-900/40 px-4 py-6 backdrop-blur-[2px]"
        @click.self="routeOpen = false"
      >
        <section class="max-h-[88vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-mogao-leadwhite-50 p-5 shadow-2xl">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] tracking-[0.2em] text-mogao-gold-600">ONE-SENTENCE ROUTE</p>
              <h2 class="mt-1 font-serif text-2xl font-bold text-mogao-lapis-800">一句话生成参观路线</h2>
              <p class="mt-1 text-xs text-mogao-ochre-500">告诉织语你有多少时间、喜欢什么工艺，生成可逐站执行的 3D 路线。</p>
            </div>
            <button class="rounded-full p-2 text-mogao-ochre-500 hover:bg-mogao-ochre-100" aria-label="关闭路线规划" @click="routeOpen = false">✕</button>
          </div>

          <div class="mt-4 flex flex-col gap-2 sm:flex-row">
            <input
              v-model="routeQuery"
              type="text"
              maxlength="200"
              class="min-w-0 flex-1 rounded-xl border border-mogao-ochre-300 bg-white px-3.5 py-2.5 text-sm text-mogao-lapis-800 placeholder:text-mogao-ochre-400 focus:border-mogao-gold-500 focus:outline-none focus:ring-2 focus:ring-mogao-gold-200"
              placeholder="例如：我有40分钟，喜欢刺绣和银饰，想看最有代表性的"
              @keydown.enter.prevent="generateRoute"
            />
            <button
              type="button"
              class="btn-mogao-primary shrink-0 text-sm disabled:opacity-40"
              :disabled="routeLoading || !routeQuery.trim()"
              @click="generateRoute"
            >
              {{ routeLoading ? '规划中…' : '生成路线' }}
            </button>
          </div>

          <p v-if="routeError" class="mt-3 rounded-lg bg-mogao-cinnabar-50 px-3 py-2 text-xs text-mogao-cinnabar-700">
            {{ routeError }}
          </p>

          <template v-if="routePlan">
            <div class="mt-4 flex items-center justify-between gap-3 border-b border-mogao-ochre-200 pb-2">
              <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800">{{ routePlan.title }}</h3>
              <span class="text-xs text-mogao-ochre-500">共 {{ routePlan.stops.length }} 站 · 约 {{ routePlan.total_minutes }} 分钟</span>
            </div>
            <div class="mt-3 space-y-2">
              <div
                v-for="stop in routePlan.stops"
                :key="stop.order"
                class="flex items-start gap-3 rounded-xl border border-mogao-ochre-200 bg-white p-3"
              >
                <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-mogao-cinnabar text-xs font-semibold text-white">{{ stop.order }}</span>
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                    <span class="font-serif text-sm font-semibold text-mogao-lapis-800">{{ stop.name }}</span>
                    <span class="text-[11px] text-mogao-ochre-500">{{ stop.ethnic }} · {{ stop.craft }} · {{ stop.duration_min }} 分钟</span>
                  </div>
                  <p class="mt-1 text-xs text-mogao-lapis-600">3D 热点：{{ stop.annotation_title }}</p>
                  <p class="mt-1 text-[11px] text-mogao-gold-700">{{ stop.reason }}</p>
                </div>
              </div>
            </div>
            <div class="mt-4 flex items-center justify-between gap-3">
              <p class="text-[11px] leading-relaxed text-mogao-ochre-500">路线按馆藏与热点生成；进入详情后可逐站镜头飞行、语音讲解，随时退出。</p>
              <button type="button" class="btn-mogao-primary shrink-0 text-sm" @click="startRoute">开始路线</button>
            </div>
          </template>
          <p v-else class="mt-4 text-xs leading-relaxed text-mogao-ochre-500">
            试试：“我有 30 分钟，想看蜡染和银饰”“第一次来，想看最有代表性的 5 件”。
          </p>
        </section>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, type Ref, type ComputedRef } from 'vue'
import { useRouter } from 'vue-router'
import { garmentAPI } from '@/services/api'
import { aiAPI } from '@/services/ai'
import { tourAPI, type TourPlan } from '@/services/tour'
import { useDebounce } from '@/composables'
import Card from '@/components/Common/Card.vue'
import Skeleton from '@/components/Common/Skeleton.vue'
import EmptyState from '@/components/Common/EmptyState.vue'
import type { GarmentListItem } from '@/types/garment'
import type { ApiResponse, PaginatedData } from '@/types/api'
import type { AiVisionResult } from '@/types/ai'

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

const router = useRouter()
const photoInput = ref<HTMLInputElement | null>(null)
const photoOpen = ref(false)
const photoLoading = ref(false)
const photoError = ref('')
const photoPreview = ref('')
const visionResult = ref<AiVisionResult | null>(null)
const routeOpen = ref(false)
const routeLoading = ref(false)
const routeError = ref('')
const routeQuery = ref('我有40分钟，喜欢刺绣和银饰，想看最有代表性的')
const routePlan = ref<TourPlan | null>(null)

const visionAttributes = computed(() => {
  if (!visionResult.value) return []
  const attrs = visionResult.value.attributes
  return [
    { label: '民族', value: attrs.ethnic },
    { label: '工艺', value: attrs.craft },
    { label: '纹样', value: attrs.pattern },
    { label: '朝代', value: attrs.dynasty },
  ].filter((item) => item.value)
})

function pickPhoto(): void {
  photoInput.value?.click()
}

function readAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('图片读取失败'))
    reader.readAsDataURL(file)
  })
}

async function onPhotoChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  if (!file.type.startsWith('image/')) {
    photoOpen.value = true
    photoError.value = '请选择图片文件（JPG/PNG/WebP）。'
    return
  }
  if (file.size > 8 * 1024 * 1024) {
    photoOpen.value = true
    photoError.value = '图片不能超过 8MB，请压缩后再试。'
    return
  }

  photoOpen.value = true
  photoLoading.value = true
  photoError.value = ''
  visionResult.value = null
  try {
    photoPreview.value = await readAsDataURL(file)
    visionResult.value = await aiAPI.visionSearch(photoPreview.value)
  } catch (error) {
    photoError.value = error instanceof Error ? error.message : '图片识别失败，请换一张清晰照片。'
  } finally {
    photoLoading.value = false
  }
}

function closePhoto(): void {
  photoOpen.value = false
}

function openCandidate(garmentID: number): void {
  closePhoto()
  void router.push(`/detail/${garmentID}`)
}

function openRoute(): void {
  routeOpen.value = true
}

async function generateRoute(): Promise<void> {
  if (!routeQuery.value.trim() || routeLoading.value) return
  routeLoading.value = true
  routeError.value = ''
  routePlan.value = null
  try {
    routePlan.value = await tourAPI.plan(routeQuery.value.trim())
  } catch (error) {
    routeError.value = error instanceof Error ? error.message : '路线生成失败，请稍后重试。'
  } finally {
    routeLoading.value = false
  }
}

function startRoute(): void {
  const plan = routePlan.value
  if (!plan || !plan.stops.length) return
  const first = plan.stops[0]
  sessionStorage.setItem('feiyi-route', JSON.stringify({ plan, index: 0 }))
  routeOpen.value = false
  void router.push({
    path: `/detail/${first.garment_id}`,
    query: { annotation: String(first.annotation_id), route: '1' },
  })
}

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

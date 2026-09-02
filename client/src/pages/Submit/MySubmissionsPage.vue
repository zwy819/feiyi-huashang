<template>
  <div class="page-container">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">我的投稿</h1>
        <p class="mt-2 text-sm text-mogao-ochre-500">
          共 {{ total }} 条投稿记录
        </p>
      </div>
      <router-link to="/submit" class="btn-mogao-primary text-sm">
        + 新投稿
      </router-link>
    </div>

    <!-- 状态筛选 -->
    <div class="flex gap-2 mb-6">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="activeStatus = tab.key; page = 1; loadSubmissions()"
        class="px-4 py-2 text-sm rounded-lg transition-colors"
        :class="
          activeStatus === tab.key
            ? 'bg-mogao-lapis-800 text-white'
            : 'bg-white text-mogao-lapis-700 hover:bg-mogao-ochre-100 border border-mogao-ochre-200'
        "
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <Skeleton type="card" :count="3" />
    </div>

    <!-- 空状态 -->
    <EmptyState
      v-else-if="submissions.length === 0"
      icon="📨"
      title="暂无投稿"
      description="分享你身边的非遗服饰图片，为非遗传承贡献力量"
      :action="{ label: '去投稿', to: '/submit' }"
    />

    <!-- 投稿卡片列表 -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="sub in submissions"
        :key="sub.id"
        class="bg-white rounded-xl border border-mogao-ochre-100 overflow-hidden hover:shadow-md transition-shadow"
      >
        <!-- 图片 -->
        <div class="aspect-[4/3] bg-mogao-ochre-100 overflow-hidden">
          <img
            :src="submissionAPI.getImageUrl(sub.id)"
            :alt="sub.name"
            class="w-full h-full object-cover"
            @error="($event.target as HTMLImageElement).style.display = 'none'"
          />
        </div>

        <!-- 信息 -->
        <div class="p-4 space-y-2">
          <div class="flex items-start justify-between gap-2">
            <h3 class="font-medium text-mogao-lapis-800 text-sm truncate">{{ sub.name }}</h3>
            <span
              class="shrink-0 text-xs px-2 py-0.5 rounded-full font-medium"
              :class="statusColors[sub.status]"
            >
              {{ statusLabels[sub.status] }}
            </span>
          </div>

          <div v-if="sub.dynasty || sub.ethnic" class="text-xs text-mogao-lapis-700/60">
            <span v-if="sub.dynasty">{{ sub.dynasty }}代</span>
            <span v-if="sub.dynasty && sub.ethnic"> · </span>
            <span v-if="sub.ethnic">{{ sub.ethnic }}</span>
          </div>

          <!-- 驳回原因 -->
          <div
            v-if="sub.status === 'rejected' && sub.reject_reason"
            class="text-xs text-red-600 bg-red-50 px-2.5 py-1.5 rounded"
          >
            驳回原因：{{ sub.reject_reason }}
          </div>

          <!-- 撤回按钮（已撤回的不可再撤回） -->
          <button
            v-if="sub.status !== 'retracted'"
            @click="handleRetract(sub)"
            class="text-xs px-2.5 py-1 rounded border border-mogao-ochre-300 text-mogao-ochre-600 hover:bg-red-50 hover:text-red-600 hover:border-red-300 transition-colors"
          >
            撤回投稿
          </button>

          <!-- 关联信息 -->
          <div v-if="sub.created_garment_id" class="text-xs text-mogao-malachite-600">
            <router-link :to="`/detail/${sub.created_garment_id}`" class="hover:underline">
              查看服饰详情 →
            </router-link>
          </div>

          <div class="text-xs text-mogao-lapis-700/40 pt-1">
            {{ formatDate(sub.created_at) }}
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div
      v-if="total > limit"
      class="flex items-center justify-center gap-2 mt-8"
    >
      <button
        @click="page--; loadSubmissions()"
        :disabled="page <= 1"
        class="px-3 py-1.5 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        上一页
      </button>
      <span class="text-sm text-mogao-lapis-700 px-3">
        {{ page }} / {{ Math.ceil(total / limit) }}
      </span>
      <button
        @click="page++; loadSubmissions()"
        :disabled="page >= Math.ceil(total / limit)"
        class="px-3 py-1.5 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        下一页
      </button>
    </div>

    <!-- 撤回确认弹窗 -->
    <div
      v-if="retractTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="retractTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-sm shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-3">确认撤回</h2>
        <p class="text-sm text-mogao-lapis-700/70 mb-2">
          撤回后该投稿将从"用户贡献"区域移除。
        </p>
        <p
          v-if="retractTarget.created_garment_id"
          class="text-sm text-red-600 bg-red-50 px-3 py-2 rounded mb-4"
        >
          ⚠️ 该投稿已关联服饰记录，撤回后将同时删除服饰，此操作不可撤销。
        </p>
        <p v-else class="text-sm text-mogao-ochre-500 mb-4">
          撤回后仍可在"已撤回"标签页中查看。
        </p>
        <div class="flex justify-end gap-3">
          <button
            @click="retractTarget = null"
            class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            @click="confirmRetract"
            :disabled="retracting"
            class="px-5 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {{ retracting ? '撤回中...' : '确认撤回' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { submissionAPI } from '@/services/api'
import { formatDate } from '@/utils'
import { SubmissionStatusLabels, SubmissionStatusColors, type Submission, type SubmissionStatus } from '@/types/submission'
import Skeleton from '@/components/Common/Skeleton.vue'
import EmptyState from '@/components/Common/EmptyState.vue'

const submissions: Ref<Submission[]> = ref([])
const loading = ref(true)
const page = ref(1)
const limit = 12
const total = ref(0)
const activeStatus: Ref<string> = ref('')

const statusLabels = SubmissionStatusLabels
const statusColors = SubmissionStatusColors

const tabs = [
  { key: '', label: '全部' },
  { key: 'pending', label: '待审核' },
  { key: 'approved', label: '已通过' },
  { key: 'rejected', label: '已驳回' },
  { key: 'completed', label: '已完成' },
  { key: 'retracted', label: '已撤回' },
]

// 撤回功能
const retractTarget: Ref<Submission | null> = ref(null)
const retracting = ref(false)

async function handleRetract(sub: Submission) {
  retractTarget.value = sub
}

async function confirmRetract() {
  if (!retractTarget.value) return
  retracting.value = true
  try {
    await submissionAPI.retract(retractTarget.value.id)
    retractTarget.value = null
    loadSubmissions()
  } catch (err) {
    alert(err instanceof Error ? err.message : '撤回失败')
  } finally {
    retracting.value = false
  }
}

async function loadSubmissions() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, limit }
    if (activeStatus.value) {
      params.status = activeStatus.value
    }
    const res = await submissionAPI.listMy(params as { page?: number; limit?: number; status?: string })
    const data = (res as unknown as { data: { items: Submission[]; total: number } }).data
    submissions.value = data.items
    total.value = data.total
  } catch {
    // fail silently, empty state shows
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSubmissions()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-mogao-lapis-900">评论审核</h1>
    </div>

    <!-- 状态筛选 -->
    <div class="flex gap-2 mb-6">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="activeStatus = tab.key; page = 1; loadComments()"
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

    <!-- Table -->
    <div class="bg-white rounded-xl shadow-sm border border-mogao-ochre-100 overflow-hidden">
      <div v-if="loading" class="p-8 text-center text-mogao-lapis-700/60">加载中...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500">{{ error }}</div>
      <div v-else-if="comments.length === 0" class="p-8 text-center text-mogao-lapis-700/60">
        暂无评论记录
      </div>

      <table v-else class="w-full text-sm">
        <thead>
          <tr class="bg-mogao-ochre-50 text-left text-mogao-lapis-700">
            <th class="px-4 py-3 font-medium">ID</th>
            <th class="px-4 py-3 font-medium">评论内容</th>
            <th class="px-4 py-3 font-medium">服饰</th>
            <th class="px-4 py-3 font-medium">评论人</th>
            <th class="px-4 py-3 font-medium">状态</th>
            <th class="px-4 py-3 font-medium">点赞</th>
            <th class="px-4 py-3 font-medium">发表时间</th>
            <th class="px-4 py-3 font-medium">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="comment in comments"
            :key="comment.id"
            class="border-t border-mogao-ochre-100 hover:bg-mogao-ochre-50/50 transition-colors"
          >
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ comment.id }}</td>
            <td class="px-4 py-3 max-w-[280px]">
              <p class="line-clamp-2 whitespace-pre-wrap break-words" :title="comment.content">
                {{ comment.content }}
              </p>
              <p v-if="comment.status === 'rejected' && comment.reject_reason" class="text-xs text-red-500 mt-1">
                驳回原因：{{ comment.reject_reason }}
              </p>
            </td>
            <td class="px-4 py-3">
              <router-link
                :to="`/detail/${comment.garment_id}`"
                class="text-mogao-lapis-700 hover:text-mogao-cinnabar-500 transition-colors"
                :title="comment.garment_name || `服饰 #${comment.garment_id}`"
              >
                {{ comment.garment_name || `#${comment.garment_id}` }}
              </router-link>
            </td>
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ comment.username || '-' }}</td>
            <td class="px-4 py-3">
              <span
                class="text-xs px-2 py-0.5 rounded-full font-medium"
                :class="statusColors[comment.status]"
              >
                {{ statusLabels[comment.status] }}
              </span>
            </td>
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ comment.like_count }}</td>
            <td class="px-4 py-3 text-mogao-lapis-700/60 text-xs">{{ comment.created_at }}</td>
            <td class="px-4 py-3">
              <div class="flex gap-2 flex-wrap">
                <!-- pending → approve / reject -->
                <template v-if="comment.status === 'pending'">
                  <button
                    @click="handleApprove(comment.id)"
                    class="text-xs px-2 py-1 bg-green-100 text-green-700 rounded hover:bg-green-200 transition-colors"
                  >
                    通过
                  </button>
                  <button
                    @click="openRejectModal(comment)"
                    class="text-xs px-2 py-1 bg-red-100 text-red-700 rounded hover:bg-red-200 transition-colors"
                  >
                    驳回
                  </button>
                </template>

                <!-- rejected → re-review -->
                <template v-if="comment.status === 'rejected'">
                  <button
                    @click="handleApprove(comment.id)"
                    class="text-xs px-2 py-1 bg-amber-100 text-amber-700 rounded hover:bg-amber-200 transition-colors"
                  >
                    重新通过
                  </button>
                </template>

                <!-- delete -->
                <button
                  @click="openDeleteConfirm(comment)"
                  class="text-xs px-2 py-1 text-mogao-ochre-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                  title="删除此评论"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div
        v-if="total > limit"
        class="flex items-center justify-between px-4 py-3 border-t border-mogao-ochre-100"
      >
        <span class="text-sm text-mogao-lapis-700/60">共 {{ total }} 条评论</span>
        <div class="flex gap-2">
          <button
            @click="page--; loadComments()"
            :disabled="page <= 1"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            上一页
          </button>
          <span class="px-3 py-1 text-sm text-mogao-lapis-700">
            {{ page }} / {{ Math.ceil(total / limit) }}
          </span>
          <button
            @click="page++; loadComments()"
            :disabled="page >= Math.ceil(total / limit)"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- ============ 驳回弹窗 ============ -->
    <div
      v-if="rejectTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="rejectTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-4">驳回评论</h2>
        <p class="text-sm text-mogao-lapis-700/70 bg-mogao-ochre-50 rounded-lg p-3 mb-3">
          {{ rejectTarget.content }}
        </p>
        <textarea
          v-model="rejectReason"
          rows="3"
          class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                 focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                 placeholder:text-mogao-ochre-400 resize-none"
          placeholder="输入驳回原因（如：内容不当、与服饰无关等）"
        />
        <div class="flex justify-end gap-3 mt-4">
          <button
            @click="rejectTarget = null"
            class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            @click="handleReject"
            class="px-5 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
          >
            确认驳回
          </button>
        </div>
      </div>
    </div>

    <!-- ============ 删除确认弹窗 ============ -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="deleteTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-sm shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-3">确认删除</h2>
        <p class="text-sm text-mogao-lapis-700/70 mb-2">
          将删除评论 <span class="font-medium text-mogao-lapis-900">#{{ deleteTarget.id }}</span>（{{ deleteTarget.username }}），点赞记录也将一并删除。
        </p>
        <p class="text-xs text-red-500 mb-4">此操作不可恢复，请确认。</p>
        <div class="flex justify-end gap-3">
          <button
            @click="deleteTarget = null"
            class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            @click="handleDelete"
            class="px-5 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
          >
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { adminCommentAPI } from '@/services/api'
import {
  CommentStatusLabels,
  CommentStatusColors,
  type Comment,
} from '@/types/comment'
import type { ApiResponse, PaginatedData } from '@/types/api'

const loading = ref(true)
const error: Ref<string | null> = ref(null)
const comments: Ref<Comment[]> = ref([])
const page = ref(1)
const limit = 20
const total = ref(0)
const activeStatus: Ref<string> = ref('pending')

const statusLabels = CommentStatusLabels
const statusColors = CommentStatusColors

const tabs = [
  { key: 'pending', label: '待审核' },
  { key: 'approved', label: '已通过' },
  { key: 'rejected', label: '已驳回' },
  { key: '', label: '全部' },
]

// List
async function loadComments() {
  loading.value = true
  error.value = null
  try {
    const params: Record<string, unknown> = { page: page.value, limit }
    if (activeStatus.value) params.status = activeStatus.value
    const res = await adminCommentAPI.list(params as { page?: number; limit?: number; status?: string })
    const data = (res as unknown as ApiResponse<PaginatedData<Comment>>).data
    comments.value = data.items
    total.value = data.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取评论列表失败'
  } finally {
    loading.value = false
  }
}

// Approve
async function handleApprove(id: number) {
  try {
    await adminCommentAPI.approve(id)
    loadComments()
  } catch (err) {
    alert(err instanceof Error ? err.message : '操作失败')
  }
}

// Reject
const rejectTarget: Ref<Comment | null> = ref(null)
const rejectReason = ref('')

function openRejectModal(comment: Comment) {
  rejectTarget.value = comment
  rejectReason.value = ''
}

async function handleReject() {
  if (!rejectTarget.value) return
  try {
    await adminCommentAPI.reject(rejectTarget.value.id, rejectReason.value || '内容不符合展示要求')
    rejectTarget.value = null
    loadComments()
  } catch (err) {
    alert(err instanceof Error ? err.message : '操作失败')
  }
}

// Delete
const deleteTarget: Ref<Comment | null> = ref(null)

function openDeleteConfirm(comment: Comment) {
  deleteTarget.value = comment
}

async function handleDelete() {
  if (!deleteTarget.value) return
  try {
    await adminCommentAPI.remove(deleteTarget.value.id)
    deleteTarget.value = null
    loadComments()
  } catch (err) {
    alert(err instanceof Error ? err.message : '删除失败')
  }
}

onMounted(() => {
  loadComments()
})
</script>

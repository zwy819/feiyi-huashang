<template>
  <section class="mt-12">
    <div class="flex items-baseline justify-between mb-6">
      <h2 class="title-mogao font-serif text-xl font-bold text-mogao-lapis-800">
        评论留言
      </h2>
      <span class="text-xs text-mogao-ochre-500">共 {{ total }} 条</span>
    </div>

    <!-- 发表评论 -->
    <div class="card-mogao p-5 mb-6">
      <template v-if="isAuthenticated">
        <textarea
          v-model="newContent"
          rows="3"
          maxlength="500"
          class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                 focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                 placeholder:text-mogao-ochre-400 resize-none"
          placeholder="说点什么吧，分享你对这件服饰的见解…"
        />
        <div class="flex items-center justify-between mt-3">
          <span class="text-xs text-mogao-ochre-400">{{ newContent.length }} / 500 字</span>
          <button
            @click="submitComment"
            :disabled="!newContent.trim() || submitting"
            class="btn-mogao-primary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ submitting ? '提交中…' : '发表评论' }}
          </button>
        </div>
        <p class="text-xs text-mogao-ochre-400 mt-2">
          评论提交后需经管理员审核，通过后将公开展示
        </p>
      </template>
      <template v-else>
        <div class="flex items-center justify-between">
          <p class="text-sm text-mogao-ochre-500">登录后即可发表评论，分享你的见解</p>
          <router-link to="/login" class="btn-mogao-primary text-sm">去登录</router-link>
        </div>
      </template>
    </div>

    <!-- 评论列表 -->
    <div v-if="loading" class="py-10 text-center text-sm text-mogao-ochre-500">
      评论加载中…
    </div>

    <div v-else-if="error" class="card-mogao p-6 text-center text-sm text-red-500">
      {{ error }}
    </div>

    <EmptyState
      v-else-if="comments.length === 0"
      title="暂无评论"
      description="来抢沙发吧，分享你对这件服饰的见解"
      icon="💬"
    />

    <div v-else class="space-y-4">
      <div
        v-for="comment in comments"
        :key="comment.id"
        class="card-mogao p-5"
        :class="comment.status !== 'approved' ? 'opacity-80 border-dashed' : ''"
      >
        <!-- 头部：用户名 + 状态 + 时间 -->
        <div class="flex items-center gap-2 flex-wrap mb-3">
          <span
            class="w-8 h-8 rounded-full bg-mogao-ochre-100 text-mogao-lapis-700
                   flex items-center justify-center text-sm font-semibold flex-shrink-0"
          >
            {{ (comment.username || '匿名').charAt(0).toUpperCase() }}
          </span>
          <span class="text-sm font-medium text-mogao-lapis-800">
            {{ comment.username || '匿名用户' }}
          </span>
          <span
            v-if="comment.status !== 'approved'"
            class="text-xs px-2 py-0.5 rounded-full font-medium"
            :class="statusColors[comment.status]"
          >
            {{ statusLabels[comment.status] }}
          </span>
          <span class="ml-auto text-xs text-mogao-ochre-400">
            {{ formatDate(comment.created_at) }}
          </span>
        </div>

        <!-- 编辑模式 -->
        <template v-if="editingId === comment.id">
          <textarea
            v-model="editingContent"
            rows="3"
            maxlength="500"
            class="w-full px-3.5 py-2.5 text-sm border border-mogao-cinnabar-300 rounded-mogao
                   focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300 resize-none"
          />
          <p class="text-xs text-mogao-ochre-400 mt-1 mb-2">保存后将重新进入审核</p>
          <div class="flex gap-3">
            <button
              @click="saveEdit(comment.id)"
              :disabled="!editingContent.trim() || submitting"
              class="px-4 py-1.5 text-sm btn-mogao-primary disabled:opacity-50"
            >
              保存
            </button>
            <button
              @click="cancelEdit"
              class="px-4 py-1.5 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-mogao transition-colors"
            >
              取消
            </button>
          </div>
        </template>

        <!-- 查看模式 -->
        <template v-else>
          <p class="text-sm text-mogao-lapis-600 leading-relaxed whitespace-pre-wrap break-words">
            {{ comment.content }}
          </p>

          <!-- 驳回原因 -->
          <p v-if="comment.status === 'rejected' && comment.reject_reason" class="text-xs text-red-500 mt-2">
            驳回原因：{{ comment.reject_reason }}
          </p>

          <!-- 操作栏 -->
          <div class="flex items-center gap-4 mt-3 pt-3 border-t border-mogao-ochre-100">
            <!-- 点赞（仅已通过的评论可点赞） -->
            <button
              v-if="comment.status === 'approved'"
              @click="toggleLike(comment)"
              class="flex items-center gap-1.5 text-xs transition-colors"
              :class="comment.is_liked
                ? 'text-mogao-cinnabar'
                : 'text-mogao-ochre-500 hover:text-mogao-cinnabar'"
              :aria-label="comment.is_liked ? '取消点赞' : '点赞'"
            >
              <svg class="w-4 h-4" :fill="comment.is_liked ? 'currentColor' : 'none'" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
              <span>{{ comment.like_count }}</span>
            </button>
            <span v-else class="text-xs text-mogao-ochre-400">💬 {{ comment.like_count }}</span>

            <!-- 本人操作 -->
            <template v-if="isMine(comment)">
              <button
                @click="startEdit(comment)"
                class="text-xs text-mogao-ochre-500 hover:text-mogao-lapis-800 transition-colors"
              >
                编辑
              </button>
              <button
                @click="confirmDelete(comment)"
                class="text-xs text-mogao-ochre-500 hover:text-red-600 transition-colors"
              >
                删除
              </button>
            </template>
          </div>
        </template>
      </div>

      <!-- 分页 -->
      <div
        v-if="totalPages > 1"
        class="flex items-center justify-center gap-3 pt-2"
      >
        <button
          @click="page--; loadComments()"
          :disabled="page <= 1"
          class="px-3 py-1.5 text-sm rounded-mogao border border-mogao-ochre-200 text-mogao-lapis-700
                 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          上一页
        </button>
        <span class="text-sm text-mogao-lapis-700">{{ page }} / {{ totalPages }}</span>
        <button
          @click="page++; loadComments()"
          :disabled="page >= totalPages"
          class="px-3 py-1.5 text-sm rounded-mogao border border-mogao-ochre-200 text-mogao-lapis-700
                 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="deleteTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-sm shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-3">确认删除</h2>
        <p class="text-sm text-mogao-lapis-700/70 mb-4">删除后无法恢复，确定要删除这条评论吗？</p>
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
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, type Ref, type ComputedRef } from 'vue'
import { commentAPI } from '@/services/api'
import { useAuth, useToast } from '@/composables'
import { formatDate } from '@/utils'
import EmptyState from '@/components/Common/EmptyState.vue'
import { CommentStatusLabels, CommentStatusColors, type Comment } from '@/types/comment'
import type { ApiResponse, PaginatedData } from '@/types/api'

const props = defineProps<{
  garmentId: number
}>()

const { user, isAuthenticated } = useAuth()
const { showToast } = useToast()

const comments: Ref<Comment[]> = ref([])
const loading = ref(true)
const error: Ref<string | null> = ref(null)
const page = ref(1)
const limit = 10
const total = ref(0)
const totalPages: ComputedRef<number> = computed(() => Math.ceil(total.value / limit) || 1)

const statusLabels = CommentStatusLabels
const statusColors = CommentStatusColors

// ---- 发表 ----
const newContent = ref('')
const submitting = ref(false)

async function submitComment(): Promise<void> {
  const content = newContent.value.trim()
  if (!content) return
  submitting.value = true
  try {
    await commentAPI.create({ garment_id: props.garmentId, content })
    newContent.value = ''
    page.value = 1
    await loadComments()
    showToast('评论已提交，审核通过后将公开展示', 'success')
  } catch (err) {
    showToast(err instanceof Error ? err.message : '发表评论失败', 'error')
  } finally {
    submitting.value = false
  }
}

// ---- 列表 ----
async function loadComments(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const res = await commentAPI.list({ garment_id: props.garmentId, page: page.value, limit })
    const data = (res as unknown as ApiResponse<PaginatedData<Comment>>).data
    comments.value = data.items
    total.value = data.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取评论列表失败'
  } finally {
    loading.value = false
  }
}

// ---- 判断本人评论 ----
function isMine(comment: Comment): boolean {
  return isAuthenticated.value && user.value?.id === comment.user_id
}

// ---- 编辑 ----
const editingId: Ref<number | null> = ref(null)
const editingContent = ref('')

function startEdit(comment: Comment): void {
  editingId.value = comment.id
  editingContent.value = comment.content
}

function cancelEdit(): void {
  editingId.value = null
  editingContent.value = ''
}

async function saveEdit(id: number): Promise<void> {
  const content = editingContent.value.trim()
  if (!content) return
  submitting.value = true
  try {
    await commentAPI.update(id, content)
    editingId.value = null
    editingContent.value = ''
    await loadComments()
    showToast('评论已更新，将重新进入审核', 'success')
  } catch (err) {
    showToast(err instanceof Error ? err.message : '编辑评论失败', 'error')
  } finally {
    submitting.value = false
  }
}

// ---- 删除 ----
const deleteTarget: Ref<Comment | null> = ref(null)

function confirmDelete(comment: Comment): void {
  deleteTarget.value = comment
}

async function handleDelete(): Promise<void> {
  if (!deleteTarget.value) return
  try {
    await commentAPI.remove(deleteTarget.value.id)
    deleteTarget.value = null
    await loadComments()
    showToast('评论已删除', 'success')
  } catch (err) {
    showToast(err instanceof Error ? err.message : '删除评论失败', 'error')
  }
}

// ---- 点赞 ----
async function toggleLike(comment: Comment): Promise<void> {
  try {
    const res = comment.is_liked
      ? await commentAPI.unlike(comment.id)
      : await commentAPI.like(comment.id)
    const updated = (res as unknown as ApiResponse<Comment>).data
    if (updated) {
      comment.like_count = updated.like_count
      comment.is_liked = updated.is_liked
    }
  } catch (err) {
    showToast(err instanceof Error ? err.message : '点赞操作失败', 'error')
  }
}

onMounted(() => {
  if (props.garmentId) {
    loadComments()
  }
})
</script>

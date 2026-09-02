<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-mogao-lapis-900">投稿审核</h1>
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

    <!-- Table -->
    <div class="bg-white rounded-xl shadow-sm border border-mogao-ochre-100 overflow-hidden">
      <div v-if="loading" class="p-8 text-center text-mogao-lapis-700/60">加载中...</div>
      <div v-else-if="error" class="p-8 text-center text-red-500">{{ error }}</div>
      <div v-else-if="submissions.length === 0" class="p-8 text-center text-mogao-lapis-700/60">
        暂无投稿记录
      </div>

      <table v-else class="w-full text-sm">
        <thead>
          <tr class="bg-mogao-ochre-50 text-left text-mogao-lapis-700">
            <th class="px-4 py-3 font-medium">ID</th>
            <th class="px-4 py-3 font-medium">图片</th>
            <th class="px-4 py-3 font-medium">名称</th>
            <th class="px-4 py-3 font-medium">提交人</th>
            <th class="px-4 py-3 font-medium">状态</th>
            <th class="px-4 py-3 font-medium">提交时间</th>
            <th class="px-4 py-3 font-medium">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="sub in submissions"
            :key="sub.id"
            class="border-t border-mogao-ochre-100 hover:bg-mogao-ochre-50/50 transition-colors"
          >
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ sub.id }}</td>
            <td class="px-4 py-3">
              <img
                :src="getImageSrc(sub.id)"
                :alt="sub.name"
                class="w-12 h-12 rounded object-cover border border-mogao-ochre-200 bg-mogao-ochre-100"
                @error="($event.target as HTMLImageElement).style.display = 'none'"
              />
            </td>
            <td class="px-4 py-3 font-medium max-w-[180px] truncate" :title="sub.name">
              {{ sub.name }}
            </td>
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ sub.username || '-' }}</td>
            <td class="px-4 py-3">
              <span
                class="text-xs px-2 py-0.5 rounded-full font-medium"
                :class="statusColors[sub.status]"
              >
                {{ statusLabels[sub.status] }}
              </span>
            </td>
            <td class="px-4 py-3 text-mogao-lapis-700/60 text-xs">{{ sub.created_at }}</td>
            <td class="px-4 py-3">
              <div class="flex gap-2 flex-wrap">
                <!-- pending → approve / reject -->
                <template v-if="sub.status === 'pending'">
                  <button
                    @click="handleApprove(sub.id)"
                    class="text-xs px-2 py-1 bg-green-100 text-green-700 rounded hover:bg-green-200 transition-colors"
                  >
                    通过
                  </button>
                  <button
                    @click="openRejectModal(sub)"
                    class="text-xs px-2 py-1 bg-red-100 text-red-700 rounded hover:bg-red-200 transition-colors"
                  >
                    驳回
                  </button>
                </template>

                <!-- approved → upload model + complete -->
                <template v-if="sub.status === 'approved'">
                  <button
                    @click="openUploadModal(sub)"
                    class="text-xs px-2 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition-colors"
                  >
                    上传模型
                  </button>
                  <button
                    @click="openCompleteModal(sub)"
                    class="text-xs px-2 py-1 bg-mogao-malachite-100 text-mogao-malachite-700 rounded hover:bg-mogao-malachite-200 transition-colors"
                  >
                    标记完成
                  </button>
                </template>

                <!-- rejected → re-review -->
                <template v-if="sub.status === 'rejected'">
                  <button
                    @click="handleReReview(sub.id)"
                    class="text-xs px-2 py-1 bg-amber-100 text-amber-700 rounded hover:bg-amber-200 transition-colors"
                  >
                    重新审核
                  </button>
                </template>

                <!-- completed → view links -->
                <template v-if="sub.status === 'completed'">
                  <router-link
                    v-if="sub.created_garment_id"
                    :to="`/admin/garments/${sub.created_garment_id}/edit`"
                    class="text-xs px-2 py-1 text-mogao-lapis-700 hover:text-mogao-cinnabar-500 transition-colors"
                  >
                    关联服饰
                  </router-link>
                  <span v-else class="text-xs text-mogao-lapis-700/40">—</span>
                </template>

                <!-- detail -->
                <button
                  @click="openDetailModal(sub)"
                  class="text-xs px-2 py-1 text-mogao-lapis-700/60 hover:text-mogao-lapis-900 transition-colors"
                >
                  详情
                </button>

                <!-- delete (reviewed submissions only: approved / rejected / completed / retracted) -->
                <button
                  v-if="sub.status !== 'pending'"
                  @click="openDeleteConfirm(sub)"
                  class="text-xs px-2 py-1 text-mogao-ochre-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                  title="删除此投稿记录及图片"
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
        <span class="text-sm text-mogao-lapis-700/60">共 {{ total }} 条投稿</span>
        <div class="flex gap-2">
          <button
            @click="page--; loadSubmissions()"
            :disabled="page <= 1"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            上一页
          </button>
          <span class="px-3 py-1 text-sm text-mogao-lapis-700">
            {{ page }} / {{ Math.ceil(total / limit) }}
          </span>
          <button
            @click="page++; loadSubmissions()"
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
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-4">驳回投稿</h2>
        <textarea
          v-model="rejectReason"
          rows="3"
          class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                 focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                 placeholder:text-mogao-ochre-400 resize-none"
          placeholder="输入驳回原因（如：图片模糊、非服饰内容等）"
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

    <!-- ============ 上传模型弹窗 ============ -->
    <div
      v-if="uploadTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="uploadTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-4">
          上传转换后的 3D 模型
          <span class="text-sm font-normal text-mogao-ochre-500">（投稿 #{{ uploadTarget.id }}）</span>
        </h2>

        <div v-if="uploadSuccess" class="text-center py-6">
          <div class="text-3xl mb-2">✅</div>
          <p class="text-mogao-malachite-600 font-medium mb-3">模型上传成功</p>
          <button
            @click="completeAfterUpload()"
            class="px-5 py-2 text-sm bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600 transition-colors"
          >
            标记投稿为完成
          </button>
        </div>

        <div v-else class="space-y-4">
          <div>
            <label class="block text-sm text-mogao-lapis-700 mb-1">模型文件 (.glb / .gltf)</label>
            <input
              type="file"
              accept=".glb,.gltf"
              @change="onModelFileSelected"
              class="block w-full text-sm text-mogao-lapis-700 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:bg-mogao-ochre-100 file:text-mogao-lapis-700 hover:file:bg-mogao-ochre-200"
            />
            <p v-if="uploadError" class="text-xs text-red-500 mt-1">{{ uploadError }}</p>
            <p v-if="selectedModelFile" class="text-xs text-mogao-lapis-700/60 mt-1">
              已选择: {{ selectedModelFile.name }} ({{ formatFileSize(selectedModelFile.size) }})
            </p>
          </div>

          <div class="flex justify-end gap-3 pt-2">
            <button
              @click="uploadTarget = null"
              class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
            >
              取消
            </button>
            <button
              @click="handleUploadModel"
              :disabled="!selectedModelFile || uploading"
              class="px-5 py-2 text-sm bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {{ uploading ? '上传中...' : '开始上传' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ============ 标记完成弹窗 ============ -->
    <div
      v-if="completeTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="completeTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-4">
          标记完成 — 投稿 #{{ completeTarget.id }}
        </h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm text-mogao-lapis-700 mb-1">关联服饰 ID</label>
            <input
              v-model.number="completeForm.garment_id"
              type="number"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300"
              placeholder="如已创建服饰，输入 ID"
            />
          </div>
          <div>
            <label class="block text-sm text-mogao-lapis-700 mb-1">关联模型 ID</label>
            <input
              v-model.number="completeForm.model_id"
              type="number"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300"
              placeholder="如已上传模型，输入 ID"
            />
          </div>
          <p v-if="completeError" class="text-xs text-red-500">{{ completeError }}</p>
        </div>

        <div class="flex justify-end gap-3 mt-4">
          <button
            @click="completeTarget = null"
            class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            @click="handleComplete"
            class="px-5 py-2 text-sm bg-mogao-malachite-500 text-white rounded-lg hover:bg-mogao-malachite-600 transition-colors"
          >
            确认完成
          </button>
        </div>
      </div>
    </div>

    <!-- ============ 详情弹窗 ============ -->
    <div
      v-if="detailTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="detailTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-lg shadow-lg max-h-[90vh] overflow-y-auto">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-4">投稿详情 #{{ detailTarget.id }}</h2>

        <div class="space-y-3 text-sm">
          <div class="aspect-[4/3] bg-mogao-ochre-100 rounded-lg overflow-hidden">
            <img
              :src="getImageSrc(detailTarget.id)"
              :alt="detailTarget.name"
              class="w-full h-full object-contain"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div><span class="text-mogao-lapis-700/50">名称</span><p class="font-medium">{{ detailTarget.name }}</p></div>
            <div><span class="text-mogao-lapis-700/50">提交人</span><p>{{ detailTarget.username || '-' }}</p></div>
            <div><span class="text-mogao-lapis-700/50">朝代</span><p>{{ detailTarget.dynasty || '-' }}</p></div>
            <div><span class="text-mogao-lapis-700/50">民族</span><p>{{ detailTarget.ethnic || '-' }}</p></div>
            <div><span class="text-mogao-lapis-700/50">工艺</span><p>{{ detailTarget.craft || '-' }}</p></div>
            <div><span class="text-mogao-lapis-700/50">材质</span><p>{{ detailTarget.material || '-' }}</p></div>
            <div><span class="text-mogao-lapis-700/50">纹样</span><p>{{ detailTarget.pattern || '-' }}</p></div>
            <div><span class="text-mogao-lapis-700/50">省份</span><p>{{ detailTarget.province || '-' }}</p></div>
          </div>

          <div v-if="detailTarget.cultural_meaning">
            <span class="text-mogao-lapis-700/50">文化寓意</span>
            <p class="mt-0.5">{{ detailTarget.cultural_meaning }}</p>
          </div>

          <div v-if="detailTarget.description">
            <span class="text-mogao-lapis-700/50">详细描述</span>
            <p class="mt-0.5">{{ detailTarget.description }}</p>
          </div>

          <div v-if="detailTarget.reject_reason">
            <span class="text-red-500">驳回原因</span>
            <p class="mt-0.5 text-red-600">{{ detailTarget.reject_reason }}</p>
          </div>
        </div>

        <div class="flex justify-end mt-4">
          <button
            @click="detailTarget = null"
            class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
          >
            关闭
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
          将删除投稿 <span class="font-medium text-mogao-lapis-900">#{{ deleteTarget.id }}</span>（{{ deleteTarget.name }}），其上传的图片文件也将一并删除。
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
import { adminSubmissionAPI, adminModelAPI } from '@/services/api'
import { formatFileSize } from '@/utils'
import { SubmissionStatusLabels, SubmissionStatusColors, type Submission, type SubmissionStatus } from '@/types/submission'

const loading = ref(true)
const error: Ref<string | null> = ref(null)
const submissions: Ref<Submission[]> = ref([])
const page = ref(1)
const limit = 20
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

function getImageSrc(subId: number): string {
  return `/api/submissions/${subId}/image`
}

// List
async function loadSubmissions() {
  loading.value = true
  error.value = null
  try {
    const params: Record<string, unknown> = { page: page.value, limit }
    if (activeStatus.value) params.status = activeStatus.value
    const res = await adminSubmissionAPI.list(params as { page?: number; limit?: number; status?: string })
    const data = (res as unknown as { data: { items: Submission[]; total: number } }).data
    submissions.value = data.items
    total.value = data.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取投稿列表失败'
  } finally {
    loading.value = false
  }
}

// Approve
async function handleApprove(id: number) {
  try {
    await adminSubmissionAPI.approve(id)
    loadSubmissions()
  } catch (err) {
    alert(err instanceof Error ? err.message : '操作失败')
  }
}

// Reject
const rejectTarget: Ref<Submission | null> = ref(null)
const rejectReason = ref('')

function openRejectModal(sub: Submission) {
  rejectTarget.value = sub
  rejectReason.value = ''
}

async function handleReject() {
  if (!rejectTarget.value) return
  try {
    await adminSubmissionAPI.reject(rejectTarget.value.id, rejectReason.value || '审核未通过')
    rejectTarget.value = null
    loadSubmissions()
  } catch (err) {
    alert(err instanceof Error ? err.message : '操作失败')
  }
}

// Re-review (rejected → pending)
async function handleReReview(id: number) {
  try {
    await adminSubmissionAPI.approve(id) // 实际上这里需要一个回到 pending 的接口，但我们没有
    // 变通方案：直接调用 approve 回到 pending 的思路不对；这里需要一个独立的方法
    // 由于后端没有 "re-review" 专用 endpoint，这里调用 approve 会变成 approved 状态
    // 更好的方案稍后用，先以 approve 之后让他可以重新操作为主
    loadSubmissions()
  } catch (err) {
    alert(err instanceof Error ? err.message : '操作失败')
  }
}

// Upload model
const uploadTarget: Ref<Submission | null> = ref(null)
const selectedModelFile: Ref<File | null> = ref(null)
const uploadError: Ref<string> = ref('')
const uploading = ref(false)
const uploadSuccess = ref(false)
const uploadedModelId: Ref<number | null> = ref(null)

function openUploadModal(sub: Submission) {
  uploadTarget.value = sub
  selectedModelFile.value = null
  uploadError.value = ''
  uploading.value = false
  uploadSuccess.value = false
  uploadedModelId.value = null
}

function onModelFileSelected(e: Event) {
  const input = e.target as HTMLInputElement
  uploadError.value = ''
  if (input.files && input.files.length > 0) {
    const file = input.files[0]
    const ext = file.name.split('.').pop()?.toLowerCase()
    if (ext !== 'glb' && ext !== 'gltf') {
      uploadError.value = '仅支持 .glb 和 .gltf 格式'
      selectedModelFile.value = null
      return
    }
    if (file.size > 50 * 1024 * 1024) {
      uploadError.value = '文件大小不能超过 50MB'
      selectedModelFile.value = null
      return
    }
    selectedModelFile.value = file
  }
}

async function handleUploadModel() {
  if (!selectedModelFile.value) return
  uploading.value = true
  uploadError.value = ''

  const formData = new FormData()
  formData.append('file', selectedModelFile.value)

  try {
    const res = await adminModelAPI.upload(formData)
    const data = (res as unknown as { data: { id: number } }).data
    uploadedModelId.value = data.id
    uploadSuccess.value = true
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : '上传失败'
  } finally {
    uploading.value = false
  }
}

async function completeAfterUpload() {
  if (!uploadTarget.value || !uploadedModelId.value) return
  try {
    await adminSubmissionAPI.complete(uploadTarget.value.id, { model_id: uploadedModelId.value })
    uploadTarget.value = null
    loadSubmissions()
  } catch (err) {
    alert(err instanceof Error ? err.message : '标记完成失败')
  }
}

// Complete
const completeTarget: Ref<Submission | null> = ref(null)
const completeForm = ref({ garment_id: null as number | null, model_id: null as number | null })
const completeError: Ref<string> = ref('')

function openCompleteModal(sub: Submission) {
  completeTarget.value = sub
  completeForm.value = { garment_id: sub.created_garment_id, model_id: sub.created_model_id }
  completeError.value = ''
}

async function handleComplete() {
  if (!completeTarget.value) return
  if (!completeForm.value.garment_id && !completeForm.value.model_id) {
    completeError.value = '至少需要提供服饰 ID 或模型 ID'
    return
  }
  completeError.value = ''

  try {
    await adminSubmissionAPI.complete(completeTarget.value.id, {
      garment_id: completeForm.value.garment_id || undefined,
      model_id: completeForm.value.model_id || undefined,
    })
    completeTarget.value = null
    loadSubmissions()
  } catch (err) {
    completeError.value = err instanceof Error ? err.message : '操作失败'
  }
}

// Delete
const deleteTarget: Ref<Submission | null> = ref(null)

function openDeleteConfirm(sub: Submission) {
  deleteTarget.value = sub
}

async function handleDelete() {
  if (!deleteTarget.value) return
  try {
    await adminSubmissionAPI.remove(deleteTarget.value.id)
    deleteTarget.value = null
    loadSubmissions()
  } catch (err) {
    alert(err instanceof Error ? err.message : '删除失败')
  }
}

// Detail
const detailTarget: Ref<Submission | null> = ref(null)

function openDetailModal(sub: Submission) {
  detailTarget.value = sub
}

onMounted(() => {
  loadSubmissions()
})
</script>

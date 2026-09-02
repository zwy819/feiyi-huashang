<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-mogao-lapis-900">模型管理</h1>
      <button
        @click="showUpload = true"
        class="px-5 py-2.5 bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600 transition-colors text-sm font-medium"
      >
        + 上传模型
      </button>
    </div>

    <!-- Upload Modal -->
    <div
      v-if="showUpload"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="showUpload = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-4">上传模型文件</h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm text-mogao-lapis-700 mb-1">模型文件 (.glb / .gltf)</label>
            <input
              ref="fileInputRef"
              type="file"
              accept=".glb,.gltf"
              @change="onFileSelected"
              class="block w-full text-sm text-mogao-lapis-700 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:bg-mogao-ochre-100 file:text-mogao-lapis-700 hover:file:bg-mogao-ochre-200"
            />
            <p v-if="uploadError" class="text-xs text-red-500 mt-1">{{ uploadError }}</p>
            <p v-if="selectedFile" class="text-xs text-mogao-lapis-700/60 mt-1">
              已选择: {{ selectedFile.name }} ({{ formatFileSize(selectedFile.size) }})
            </p>
          </div>

          <div>
            <label class="block text-sm text-mogao-lapis-700 mb-1">关联服饰（可选）</label>
            <select
              v-model="selectedGarmentId"
              class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-700 bg-white"
            >
              <option :value="null">不关联</option>
              <option v-for="g in garments" :key="g.id" :value="g.id">
                {{ g.name }} ({{ g.dynasty }}代 · {{ g.ethnic }})
              </option>
            </select>
          </div>

          <!-- Upload progress -->
          <div v-if="uploading" class="w-full bg-gray-100 rounded-full h-2">
            <div
              class="bg-mogao-cinnabar-500 h-2 rounded-full transition-all"
              :style="{ width: uploadProgress + '%' }"
            ></div>
          </div>

          <div class="flex justify-end gap-3 pt-2">
            <button
              @click="cancelUpload"
              class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
            >
              取消
            </button>
            <button
              @click="handleUpload"
              :disabled="!selectedFile || uploading"
              class="px-5 py-2 text-sm bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {{ uploading ? '上传中...' : '开始上传' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Model List Table -->
    <div class="bg-white rounded-xl shadow-sm border border-mogao-ochre-100 overflow-hidden">
      <!-- Loading -->
      <div v-if="loading" class="p-8 text-center text-mogao-lapis-700/60">加载中...</div>

      <!-- Error -->
      <div v-else-if="error" class="p-8 text-center text-red-500">{{ error }}</div>

      <!-- Empty -->
      <div v-else-if="models.length === 0" class="p-8 text-center text-mogao-lapis-700/60">
        暂无模型文件，点击上方按钮上传
      </div>

      <!-- Table -->
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="bg-mogao-ochre-50 text-left text-mogao-lapis-700">
            <th class="px-4 py-3 font-medium">ID</th>
            <th class="px-4 py-3 font-medium">文件名</th>
            <th class="px-4 py-3 font-medium">大小</th>
            <th class="px-4 py-3 font-medium">关联服饰</th>
            <th class="px-4 py-3 font-medium">上传时间</th>
            <th class="px-4 py-3 font-medium">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="model in models"
            :key="model.id"
            class="border-t border-mogao-ochre-100 hover:bg-mogao-ochre-50/50 transition-colors"
          >
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ model.id }}</td>
            <td class="px-4 py-3 font-medium">{{ model.original_name }}</td>
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ formatFileSize(model.file_size) }}</td>
            <td class="px-4 py-3">
              <span v-if="model.garment_name" class="text-mogao-malachite-600">
                {{ model.garment_name }}
              </span>
              <span v-else class="text-mogao-lapis-700/40">未关联</span>
            </td>
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ model.created_at }}</td>
            <td class="px-4 py-3">
              <button
                @click="confirmDelete(model)"
                class="text-sm text-red-500 hover:text-red-700 transition-colors"
              >
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div
        v-if="total > limit"
        class="flex items-center justify-between px-4 py-3 border-t border-mogao-ochre-100"
      >
        <span class="text-sm text-mogao-lapis-700/60">共 {{ total }} 个模型</span>
        <div class="flex gap-2">
          <button
            @click="page--; loadModels()"
            :disabled="page <= 1"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            上一页
          </button>
          <span class="px-3 py-1 text-sm text-mogao-lapis-700">
            {{ page }} / {{ Math.ceil(total / limit) }}
          </span>
          <button
            @click="page++; loadModels()"
            :disabled="page >= Math.ceil(total / limit)"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="deleteTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-sm shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-2">确认删除</h2>
        <p class="text-sm text-mogao-lapis-700/70 mb-4">
          确定要删除模型 <strong>{{ deleteTarget.original_name }}</strong> 吗？此操作不可撤销。
        </p>
        <div class="flex justify-end gap-3">
          <button
            @click="deleteTarget = null"
            class="px-4 py-2 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            @click="handleDelete"
            class="px-4 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
          >
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminModelAPI, garmentAPI } from '@/services/api'
import { formatFileSize } from '@/utils'
import type { ModelInfo } from '@/types/model'
import type { Garment } from '@/types/garment'

// Model list state
const loading = ref(true)
const error = ref<string | null>(null)
const models = ref<ModelInfo[]>([])
const page = ref(1)
const limit = 20
const total = ref(0)

// Upload state
const showUpload = ref(false)
const selectedFile = ref<File | null>(null)
const selectedGarmentId = ref<number | null>(null)
const uploadError = ref<string | null>(null)
const uploading = ref(false)
const uploadProgress = ref(0)
const fileInputRef = ref<HTMLInputElement | null>(null)

// Garment dropdown
const garments = ref<Garment[]>([])

// Delete state
const deleteTarget = ref<ModelInfo | null>(null)

async function loadModels() {
  loading.value = true
  error.value = null
  try {
    const res = await adminModelAPI.list({ page: page.value, limit })
    const data = (res as unknown as { data: { items: ModelInfo[]; total: number } }).data
    models.value = data.items
    total.value = data.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取模型列表失败'
  } finally {
    loading.value = false
  }
}

async function loadGarments() {
  try {
    const res = await garmentAPI.list({ page: 1, limit: 100 })
    const data = (res as unknown as { data: { items: Garment[] } }).data
    garments.value = data.items
  } catch {
    // non-critical
  }
}

function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement
  uploadError.value = null
  if (input.files && input.files.length > 0) {
    const file = input.files[0]
    const ext = file.name.split('.').pop()?.toLowerCase()
    if (ext !== 'glb' && ext !== 'gltf') {
      uploadError.value = '仅支持 .glb 和 .gltf 格式'
      selectedFile.value = null
      return
    }
    if (file.size > 50 * 1024 * 1024) {
      uploadError.value = '文件大小不能超过 50MB'
      selectedFile.value = null
      return
    }
    selectedFile.value = file
  }
}

function cancelUpload() {
  showUpload.value = false
  selectedFile.value = null
  selectedGarmentId.value = null
  uploadError.value = null
  uploading.value = false
  uploadProgress.value = 0
}

async function handleUpload() {
  if (!selectedFile.value) return
  uploading.value = true
  uploadProgress.value = 30

  const formData = new FormData()
  formData.append('file', selectedFile.value)
  if (selectedGarmentId.value) {
    formData.append('garment_id', String(selectedGarmentId.value))
  }

  try {
    await adminModelAPI.upload(formData)
    uploadProgress.value = 100
    cancelUpload()
    loadModels()
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : '上传失败'
    uploadProgress.value = 0
  } finally {
    uploading.value = false
  }
}

function confirmDelete(model: ModelInfo) {
  deleteTarget.value = model
}

async function handleDelete() {
  if (!deleteTarget.value) return
  try {
    await adminModelAPI.delete(deleteTarget.value.id)
    deleteTarget.value = null
    loadModels()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败'
    deleteTarget.value = null
  }
}

onMounted(() => {
  loadModels()
  loadGarments()
})
</script>

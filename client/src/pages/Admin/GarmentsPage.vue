<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-mogao-lapis-900">服饰管理</h1>
      <router-link
        to="/admin/garments/new"
        class="px-5 py-2.5 bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600 transition-colors text-sm font-medium"
      >
        + 添加服饰
      </router-link>
    </div>

    <!-- Garment Table -->
    <div class="bg-white rounded-xl shadow-sm border border-mogao-ochre-100 overflow-hidden">
      <!-- Loading -->
      <div v-if="loading" class="p-8 text-center text-mogao-lapis-700/60">加载中...</div>

      <!-- Error -->
      <div v-else-if="error" class="p-8 text-center text-red-500">{{ error }}</div>

      <!-- Empty -->
      <div v-else-if="garments.length === 0" class="p-8 text-center text-mogao-lapis-700/60">
        暂无服饰数据
      </div>

      <!-- Table -->
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="bg-mogao-ochre-50 text-left text-mogao-lapis-700">
            <th class="px-4 py-3 font-medium">ID</th>
            <th class="px-4 py-3 font-medium">名称</th>
            <th class="px-4 py-3 font-medium">朝代</th>
            <th class="px-4 py-3 font-medium">民族</th>
            <th class="px-4 py-3 font-medium">工艺</th>
            <th class="px-4 py-3 font-medium">省份</th>
            <th class="px-4 py-3 font-medium">3D 模型</th>
            <th class="px-4 py-3 font-medium">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="garment in garments"
            :key="garment.id"
            class="border-t border-mogao-ochre-100 hover:bg-mogao-ochre-50/50 transition-colors"
          >
            <td class="px-4 py-3 text-mogao-lapis-700/60">{{ garment.id }}</td>
            <td class="px-4 py-3 font-medium">{{ garment.name }}</td>
            <td class="px-4 py-3">{{ garment.dynasty }}代</td>
            <td class="px-4 py-3">{{ garment.ethnic }}</td>
            <td class="px-4 py-3">{{ garment.craft }}</td>
            <td class="px-4 py-3">{{ garment.province }}</td>
            <td class="px-4 py-3">
              <span v-if="garment.model_id" class="text-mogao-malachite-600 text-xs">
                已关联 (ID: {{ garment.model_id }})
              </span>
              <span v-else class="text-mogao-lapis-700/40 text-xs">无</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex gap-3">
                <router-link
                  :to="`/admin/garments/${garment.id}/edit`"
                  class="text-sm text-mogao-lapis-700 hover:text-mogao-cinnabar-500 transition-colors"
                >
                  编辑
                </router-link>
                <button
                  @click="confirmDelete(garment)"
                  class="text-sm text-red-500 hover:text-red-700 transition-colors"
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
        <span class="text-sm text-mogao-lapis-700/60">共 {{ total }} 件服饰</span>
        <div class="flex gap-2">
          <button
            @click="page--; loadGarments()"
            :disabled="page <= 1"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            上一页
          </button>
          <span class="px-3 py-1 text-sm text-mogao-lapis-700">
            {{ page }} / {{ Math.ceil(total / limit) }}
          </span>
          <button
            @click="page++; loadGarments()"
            :disabled="page >= Math.ceil(total / limit)"
            class="px-3 py-1 text-sm rounded border border-mogao-ochre-200 hover:bg-mogao-ochre-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      @click.self="deleteTarget = null"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-sm shadow-lg">
        <h2 class="text-lg font-semibold text-mogao-lapis-900 mb-2">确认删除</h2>
        <p class="text-sm text-mogao-lapis-700/70 mb-4">
          确定要删除服饰 <strong>{{ deleteTarget.name }}</strong> 吗？相关的收藏和每日推荐也会被删除。
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
import { garmentAPI, adminGarmentAPI } from '@/services/api'
import type { Garment } from '@/types/garment'

const loading = ref(true)
const error = ref<string | null>(null)
const garments = ref<Garment[]>([])
const page = ref(1)
const limit = 20
const total = ref(0)
const deleteTarget = ref<Garment | null>(null)

async function loadGarments() {
  loading.value = true
  error.value = null
  try {
    const res = await garmentAPI.list({ page: page.value, limit })
    const data = (res as unknown as { data: { items: Garment[]; total: number } }).data
    garments.value = data.items
    total.value = data.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '获取服饰列表失败'
  } finally {
    loading.value = false
  }
}

function confirmDelete(garment: Garment) {
  deleteTarget.value = garment
}

async function handleDelete() {
  if (!deleteTarget.value) return
  try {
    await adminGarmentAPI.delete(deleteTarget.value.id)
    deleteTarget.value = null
    loadGarments()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败'
    deleteTarget.value = null
  }
}

onMounted(() => {
  loadGarments()
})
</script>

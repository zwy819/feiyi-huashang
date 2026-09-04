<template>
  <div>
    <!-- 页头 -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="font-serif text-2xl font-bold text-mogao-lapis-800">热点标注管理</h1>
        <p class="text-sm text-mogao-ochre-500 mt-1">
          为 3D 模型添加纹样/工艺讲解点，编辑器内点选模型表面即可放置热点
        </p>
      </div>
    </div>

    <!-- 服饰选择 -->
    <div class="card-mogao p-4 mb-6 flex flex-wrap items-center gap-3">
      <label class="text-sm text-mogao-lapis-700 font-medium flex-shrink-0">选择服饰</label>
      <select
        v-model="selectedGarmentId"
        class="flex-1 min-w-56 px-3 py-2 border border-mogao-ochre-200 rounded-mogao bg-white text-sm focus:outline-none focus:border-mogao-cinnabar"
      >
        <option :value="null">— 请选择服饰 —</option>
        <option v-for="g in garments" :key="g.id" :value="g.id">
          {{ g.name }}（{{ g.dynasty }}代 · {{ g.ethnic }}）{{ garmentModelPath(g) ? ' ✔3D' : '' }}
        </option>
      </select>
      <button
        @click="openCreate"
        :disabled="!selectedGarmentId"
        class="btn-mogao-primary text-sm disabled:opacity-40 disabled:cursor-not-allowed"
      >
        + 新增标注
      </button>
    </div>

    <!-- 标注列表 -->
    <div class="card-mogao overflow-hidden">
      <div v-if="loading" class="p-10 text-center text-sm text-mogao-ochre-500">加载中…</div>

      <template v-else-if="selectedGarmentId">
        <table v-if="annotations.length > 0" class="w-full text-sm">
          <thead>
            <tr class="bg-mogao-leadwhite-200 text-left text-mogao-lapis-700">
              <th class="px-4 py-3 font-medium">标题</th>
              <th class="px-4 py-3 font-medium">内容</th>
              <th class="px-4 py-3 font-medium whitespace-nowrap">坐标 (x, y, z)</th>
              <th class="px-4 py-3 font-medium">排序</th>
              <th class="px-4 py-3 font-medium text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="a in annotations"
              :key="a.id"
              class="border-t border-mogao-ochre-100 hover:bg-mogao-leadwhite-100 transition-colors"
            >
              <td class="px-4 py-3 font-medium text-mogao-lapis-800 whitespace-nowrap">{{ a.title }}</td>
              <td class="px-4 py-3 text-mogao-lapis-600 max-w-md">
                <p class="line-clamp-2">{{ a.content || '—' }}</p>
              </td>
              <td class="px-4 py-3 text-mogao-ochre-600 whitespace-nowrap font-mono text-xs">
                {{ a.position_x.toFixed(2) }}, {{ a.position_y.toFixed(2) }}, {{ a.position_z.toFixed(2) }}
              </td>
              <td class="px-4 py-3 text-mogao-lapis-600">{{ a.sort_order }}</td>
              <td class="px-4 py-3 text-right whitespace-nowrap">
                <button @click="openEdit(a)" class="text-mogao-lapis-500 hover:text-mogao-gold text-xs mr-3">编辑</button>
                <button @click="confirmDelete(a)" class="text-mogao-cinnabar-500 hover:text-mogao-cinnabar-700 text-xs">删除</button>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-else class="py-12 text-center">
          <span class="text-4xl block mb-2">&#x1f3f7;</span>
          <p class="text-sm text-mogao-ochre-500">该服饰还没有热点标注</p>
          <p class="text-xs text-mogao-ochre-400 mt-1">点击右上角「新增标注」开始添加</p>
        </div>
      </template>

      <div v-else class="py-12 text-center">
        <span class="text-4xl block mb-2">&#x1f457;</span>
        <p class="text-sm text-mogao-ochre-500">请先选择一件服饰</p>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <Modal :is-open="editorOpen" :title="editingId ? '编辑热点标注' : '新增热点标注'" size="lg" @close="closeEditor">
      <div class="grid md:grid-cols-2 gap-5">
        <!-- 左：3D 点选 -->
        <div>
          <div v-if="editorModelPath" class="rounded-mogao overflow-hidden border border-mogao-ochre-200">
            <ModelViewer
              :key="editorModelPath"
              :model-path="editorModelPath"
              :auto-rotate="false"
              :height="300"
              :annotations="previewAnnotations"
              :placement-mode="true"
              @surface-click="onSurfaceClick"
            />
          </div>
          <div v-else class="aspect-video bg-mogao-leadwhite-300 rounded-mogao flex items-center justify-center">
            <p class="text-xs text-mogao-ochre-500 px-6 text-center">该服饰暂无 3D 模型，请手动输入坐标</p>
          </div>
          <p class="text-xs text-mogao-ochre-500 mt-2">
            {{ editorModelPath ? '在模型上单击表面即可放置热点（拖拽用于旋转视角）' : '' }}
          </p>

          <!-- 坐标输入 -->
          <div class="grid grid-cols-3 gap-2 mt-3">
            <div v-for="axis in (['x', 'y', 'z'] as const)" :key="axis">
              <label class="text-xs text-mogao-ochre-600">{{ axis.toUpperCase() }}</label>
              <input
                v-model.number="form[`position_${axis}`]"
                type="number"
                step="0.01"
                class="w-full px-2 py-1.5 border border-mogao-ochre-200 rounded-mogao text-xs font-mono focus:outline-none focus:border-mogao-cinnabar"
              />
            </div>
          </div>
        </div>

        <!-- 右：表单 -->
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-mogao-lapis-700 font-medium mb-1">标题 <span class="text-mogao-cinnabar">*</span></label>
            <input
              v-model="form.title"
              type="text"
              maxlength="100"
              placeholder="如：蜡染底纹"
              class="w-full px-3 py-2 border border-mogao-ochre-200 rounded-mogao text-sm focus:outline-none focus:border-mogao-cinnabar"
            />
          </div>
          <div>
            <label class="block text-sm text-mogao-lapis-700 font-medium mb-1">讲解内容</label>
            <textarea
              v-model="form.content"
              rows="6"
              maxlength="1000"
              placeholder="介绍该部位的纹样、工艺或文化寓意（最多 1000 字）"
              class="w-full px-3 py-2 border border-mogao-ochre-200 rounded-mogao text-sm leading-relaxed focus:outline-none focus:border-mogao-cinnabar resize-none"
            ></textarea>
            <p class="text-xs text-mogao-ochre-400 mt-1 text-right">{{ form.content.length }} / 1000</p>
          </div>
          <div>
            <label class="block text-sm text-mogao-lapis-700 font-medium mb-1">排序（越小越靠前）</label>
            <input
              v-model.number="form.sort_order"
              type="number"
              class="w-full px-3 py-2 border border-mogao-ochre-200 rounded-mogao text-sm focus:outline-none focus:border-mogao-cinnabar"
            />
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <button @click="closeEditor" class="btn-mogao-secondary text-sm">取消</button>
            <button
              @click="save"
              :disabled="saving || !form.title.trim()"
              class="btn-mogao-primary text-sm disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ saving ? '保存中…' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Modal>

    <!-- 删除确认 -->
    <Modal :is-open="!!deleteTarget" title="确认删除" size="sm" @close="deleteTarget = null">
      <p class="text-sm text-mogao-lapis-600">
        确定要删除标注 <strong>{{ deleteTarget?.title }}</strong> 吗？此操作不可撤销。
      </p>
      <div class="flex justify-end gap-2 mt-6">
        <button @click="deleteTarget = null" class="btn-mogao-secondary text-sm">取消</button>
        <button @click="doDelete" class="px-4 py-2 bg-mogao-cinnabar text-white text-sm rounded-mogao hover:bg-mogao-cinnabar-600 transition-colors">
          确认删除
        </button>
      </div>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, type Ref } from 'vue'
import { garmentAPI, annotationAPI, adminAnnotationAPI } from '@/services/api'
import ModelViewer from '@/components/3D/ModelViewer.vue'
import Modal from '@/components/Common/Modal.vue'
import type { Annotation, AnnotationPayload } from '@/types/annotation'
import type { Garment } from '@/types/garment'
import type { ApiResponse } from '@/types/api'
import { useToast } from '@/composables'

const { showToast } = useToast()

const garments: Ref<Garment[]> = ref([])
const annotations: Ref<Annotation[]> = ref([])
const selectedGarmentId: Ref<number | null> = ref(null)
const loading: Ref<boolean> = ref(true)

// ===== 编辑弹窗状态 =====
const editorOpen: Ref<boolean> = ref(false)
const editingId: Ref<number | null> = ref(null)
const saving: Ref<boolean> = ref(false)
const form = ref<AnnotationPayload>({
  title: '',
  content: '',
  position_x: 0,
  position_y: 0,
  position_z: 0,
  sort_order: 0,
})

const deleteTarget: Ref<Annotation | null> = ref(null)

/** 编辑中服饰的模型路径 */
const editorModelPath = computed<string>(() => {
  const g = garments.value.find((item) => item.id === selectedGarmentId.value)
  return g ? garmentModelPath(g) : ''
})

/** 编辑预览：把当前表单坐标作为热点显示在模型上 */
const previewAnnotations = computed<Annotation[]>(() => {
  if (!editorOpen.value) return []
  return [
    {
      id: -1,
      garment_id: selectedGarmentId.value || 0,
      title: '当前热点位置',
      content: '',
      position_x: Number(form.value.position_x) || 0,
      position_y: Number(form.value.position_y) || 0,
      position_z: Number(form.value.position_z) || 0,
      sort_order: 0,
      created_at: '',
      updated_at: '',
    },
  ]
})

function garmentModelPath(g: Garment): string {
  if (g.model_id) return `/api/models/${g.model_id}/file`
  return g.model_url || ''
}

async function loadGarments(): Promise<void> {
  try {
    const res = await garmentAPI.list({ page: 1, limit: 100 })
    const data = res as ApiResponse<{ items: Garment[] }>
    garments.value = data.data?.items || []
  } catch {
    garments.value = []
  }
}

async function loadAnnotations(): Promise<void> {
  if (!selectedGarmentId.value) {
    annotations.value = []
    return
  }
  loading.value = true
  try {
    const res = await adminAnnotationAPI.list({ garment_id: selectedGarmentId.value })
    annotations.value = (res as ApiResponse<Annotation[]>).data || []
  } catch {
    annotations.value = []
  } finally {
    loading.value = false
  }
}

// 切换服饰时重新加载
watch(selectedGarmentId, loadAnnotations)

function openCreate(): void {
  editingId.value = null
  form.value = { title: '', content: '', position_x: 0, position_y: 0, position_z: 0, sort_order: annotations.value.length }
  editorOpen.value = true
}

function openEdit(a: Annotation): void {
  editingId.value = a.id
  form.value = {
    title: a.title,
    content: a.content,
    position_x: a.position_x,
    position_y: a.position_y,
    position_z: a.position_z,
    sort_order: a.sort_order,
  }
  editorOpen.value = true
}

function closeEditor(): void {
  editorOpen.value = false
}

/** 编辑器内点选模型表面 → 回填坐标 */
function onSurfaceClick(position: { x: number; y: number; z: number }): void {
  form.value.position_x = Number(position.x.toFixed(4))
  form.value.position_y = Number(position.y.toFixed(4))
  form.value.position_z = Number(position.z.toFixed(4))
}

async function save(): Promise<void> {
  if (!selectedGarmentId.value || !form.value.title.trim()) return
  saving.value = true
  try {
    if (editingId.value) {
      await adminAnnotationAPI.update(editingId.value, form.value)
      showToast('热点标注已更新', 'success')
    } else {
      await adminAnnotationAPI.create({ ...form.value, garment_id: selectedGarmentId.value })
      showToast('热点标注已创建', 'success')
    }
    editorOpen.value = false
    await loadAnnotations()
  } catch (err: unknown) {
    showToast(err instanceof Error ? err.message : '保存失败', 'error')
  } finally {
    saving.value = false
  }
}

function confirmDelete(a: Annotation): void {
  deleteTarget.value = a
}

async function doDelete(): Promise<void> {
  if (!deleteTarget.value) return
  try {
    await adminAnnotationAPI.remove(deleteTarget.value.id)
    showToast('热点标注已删除', 'success')
    deleteTarget.value = null
    await loadAnnotations()
  } catch (err: unknown) {
    showToast(err instanceof Error ? err.message : '删除失败', 'error')
  }
}

onMounted(async () => {
  await loadGarments()
  loading.value = false
})
</script>

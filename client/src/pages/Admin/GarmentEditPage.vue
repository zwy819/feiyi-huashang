<template>
  <div>
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-sm text-mogao-lapis-700/60 mb-6">
      <router-link to="/admin/garments" class="hover:text-mogao-cinnabar-500 transition-colors">
        服饰管理
      </router-link>
      <span>/</span>
      <span class="text-mogao-lapis-900">{{ isEdit ? '编辑服饰' : '添加服饰' }}</span>
    </div>

    <h1 class="text-2xl font-bold text-mogao-lapis-900 mb-6">
      {{ isEdit ? '编辑服饰' : '添加服饰' }}
    </h1>

    <!-- Loading (edit mode) -->
    <div v-if="loading" class="bg-white rounded-xl p-8 text-center text-mogao-lapis-700/60">
      加载中...
    </div>

    <!-- Form -->
    <form
      v-else
      @submit.prevent="handleSubmit"
      class="bg-white rounded-xl shadow-sm border border-mogao-ochre-100 p-6"
    >
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Name -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">
            名称 <span class="text-red-500">*</span>
          </label>
          <input
            v-model="form.name"
            type="text"
            required
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors"
            placeholder="例如：苗族百褶裙"
          />
        </div>

        <!-- Dynasty -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">
            朝代 <span class="text-red-500">*</span>
          </label>
          <select
            v-model="form.dynasty"
            required
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 bg-white focus:border-mogao-cinnabar-500 outline-none transition-colors"
          >
            <option value="">请选择</option>
            <option v-for="d in DYNASTY_OPTIONS" :key="d" :value="d">{{ d }}代</option>
          </select>
        </div>

        <!-- Ethnic -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">
            民族 <span class="text-red-500">*</span>
          </label>
          <select
            v-model="form.ethnic"
            required
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 bg-white focus:border-mogao-cinnabar-500 outline-none transition-colors"
          >
            <option value="">请选择</option>
            <option v-for="e in ETHNIC_OPTIONS" :key="e" :value="e">{{ e }}</option>
          </select>
        </div>

        <!-- Craft -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">工艺</label>
          <input
            v-model="form.craft"
            type="text"
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors"
            placeholder="例如：蜡染+刺绣"
          />
        </div>

        <!-- Material -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">材质</label>
          <input
            v-model="form.material"
            type="text"
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors"
            placeholder="例如：丝绸"
          />
        </div>

        <!-- Pattern -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">纹样</label>
          <input
            v-model="form.pattern"
            type="text"
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors"
            placeholder="例如：几何纹+花卉纹"
          />
        </div>

        <!-- Province -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">
            省份 <span class="text-red-500">*</span>
          </label>
          <input
            v-model="form.province"
            type="text"
            required
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors"
            placeholder="例如：贵州"
          />
        </div>

        <!-- Image URL -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">图片 URL</label>
          <input
            v-model="form.image_url"
            type="url"
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors"
            placeholder="https://..."
          />
        </div>

        <!-- Model -->
        <div>
          <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">关联 3D 模型</label>
          <select
            v-model.number="form.model_id"
            class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 bg-white focus:border-mogao-cinnabar-500 outline-none transition-colors"
          >
            <option :value="null">不关联</option>
            <option v-for="m in availableModels" :key="m.id" :value="m.id">
              {{ m.original_name }}
              <template v-if="m.garment_name">(已关联: {{ m.garment_name }})</template>
            </option>
          </select>
        </div>
      </div>

      <!-- Cultural Meaning -->
      <div class="mt-6">
        <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">文化寓意</label>
        <textarea
          v-model="form.cultural_meaning"
          rows="3"
          class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors resize-y"
          placeholder="描述该服饰的文化寓意..."
        ></textarea>
      </div>

      <!-- Description -->
      <div class="mt-6">
        <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">描述</label>
        <textarea
          v-model="form.description"
          rows="4"
          class="w-full border border-mogao-ochre-200 rounded-lg px-3 py-2 text-sm text-mogao-lapis-900 focus:border-mogao-cinnabar-500 focus:ring-1 focus:ring-mogao-cinnabar-500 outline-none transition-colors resize-y"
          placeholder="详细描述该服饰..."
        ></textarea>
      </div>

      <!-- Error message -->
      <p v-if="submitError" class="mt-4 text-sm text-red-500">{{ submitError }}</p>

      <!-- Actions -->
      <div class="mt-8 flex justify-end gap-3">
        <router-link
          to="/admin/garments"
          class="px-5 py-2.5 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
        >
          取消
        </router-link>
        <button
          type="submit"
          :disabled="submitting"
          class="px-6 py-2.5 text-sm bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
        >
          {{ submitting ? '保存中...' : (isEdit ? '更新' : '创建') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { garmentAPI, adminGarmentAPI, modelAPI } from '@/services/api'
import type { Garment, GarmentFormData } from '@/types/garment'
import type { ModelListItem } from '@/types/model'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)

// Constants
const DYNASTY_OPTIONS = ['唐', '宋', '元', '明', '清', '现代']
const ETHNIC_OPTIONS = [
  '汉族', '苗族', '藏族', '蒙古族', '傣族', '维吾尔族', '壮族', '彝族',
  '回族', '满族', '侗族', '瑶族', '白族', '土家族', '哈尼族', '哈萨克族',
  '黎族', '京族', '畲族', '朝鲜族',
]

// Form state
const loading = ref(isEdit.value)
const submitting = ref(false)
const submitError = ref<string | null>(null)

const form = ref<GarmentFormData>({
  name: '',
  dynasty: '',
  ethnic: '',
  craft: '',
  material: '',
  pattern: '',
  cultural_meaning: '',
  province: '',
  description: '',
  image_url: '',
  model_id: null,
})

const availableModels = ref<ModelListItem[]>([])

async function loadGarment() {
  if (!isEdit.value) return
  const id = Number(route.params.id)
  try {
    const res = await garmentAPI.getById(id)
    const garment = (res as unknown as { data: Garment }).data
    form.value = {
      name: garment.name,
      dynasty: garment.dynasty,
      ethnic: garment.ethnic,
      craft: garment.craft,
      material: garment.material,
      pattern: garment.pattern,
      cultural_meaning: garment.cultural_meaning,
      province: garment.province,
      description: garment.description,
      image_url: garment.image_url,
      model_id: garment.model_id,
    }
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : '获取服饰信息失败'
  } finally {
    loading.value = false
  }
}

async function loadModels() {
  try {
    const res = await modelAPI.list({ limit: 100 })
    const data = (res as unknown as { data: { items: ModelListItem[] } }).data
    availableModels.value = data.items
  } catch {
    // non-critical
  }
}

async function handleSubmit() {
  if (!form.value.name || !form.value.dynasty || !form.value.ethnic || !form.value.province) {
    submitError.value = '名称、朝代、民族、省份为必填项'
    return
  }

  submitting.value = true
  submitError.value = null

  try {
    const payload = {
      ...form.value,
      model_id: form.value.model_id || null,
    }

    if (isEdit.value) {
      await adminGarmentAPI.update(Number(route.params.id), payload as Record<string, unknown>)
    } else {
      await adminGarmentAPI.create(payload as Record<string, unknown>)
    }
    router.push('/admin/garments')
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadGarment()
  loadModels()
})
</script>

<template>
  <div class="page-container">
    <div class="mb-8">
      <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">上传投稿</h1>
      <p class="mt-2 text-sm text-mogao-ochre-500">
        上传平面服饰图片，审核通过后将公开展示并转换为 3D 模型
      </p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-5 gap-8">
      <!-- 左侧：图片上传区 -->
      <div class="lg:col-span-2">
        <div
          class="border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors"
          :class="
            dragOver
              ? 'border-mogao-cinnabar-400 bg-mogao-cinnabar-50'
              : 'border-mogao-ochre-300 hover:border-mogao-ochre-400 bg-white'
          "
          @click="triggerFileInput"
          @dragover.prevent="dragOver = true"
          @dragleave.prevent="dragOver = false"
          @drop.prevent="onDrop"
        >
          <!-- 已有图片预览 -->
          <div v-if="imagePreview" class="space-y-4">
            <img
              :src="imagePreview"
              alt="预览"
              class="max-h-64 mx-auto rounded-mogao object-contain"
            />
            <p class="text-sm text-mogao-lapis-700/70">
              {{ selectedFile?.name }} ({{ formatFileSize(selectedFile?.size || 0) }})
            </p>
            <button
              type="button"
              @click.stop="clearImage"
              class="text-xs text-mogao-cinnabar hover:underline"
            >
              移除图片
            </button>
          </div>

          <!-- 未选择图片 -->
          <div v-else class="space-y-3">
            <div class="text-4xl text-mogao-ochre-400">📷</div>
            <p class="text-sm text-mogao-lapis-700/70">
              点击或拖拽上传服饰图片
            </p>
            <p class="text-xs text-mogao-ochre-400">
              支持 JPG、PNG、WebP 格式，最大 10MB
            </p>
          </div>
        </div>

        <input
          ref="fileInputRef"
          type="file"
          accept=".jpg,.jpeg,.png,.webp"
          class="hidden"
          @change="onFileSelected"
        />

        <p v-if="imageError" class="text-sm text-red-500 mt-2">{{ imageError }}</p>
      </div>

      <!-- 右侧：元数据表单 -->
      <div class="lg:col-span-3">
        <form @submit.prevent="handleSubmit" class="bg-white rounded-xl border border-mogao-ochre-100 p-6 space-y-5">
          <div>
            <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">
              服饰名称 <span class="text-mogao-cinnabar">*</span>
            </label>
            <input
              v-model="form.name"
              type="text"
              required
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                     placeholder:text-mogao-ochre-400 transition-colors"
              placeholder="输入服饰名称"
            />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">朝代</label>
              <input
                v-model="form.dynasty"
                type="text"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400 transition-colors"
                placeholder="如：清、明、唐"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">民族</label>
              <input
                v-model="form.ethnic"
                type="text"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400 transition-colors"
                placeholder="如：汉族、苗族"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">省份</label>
              <input
                v-model="form.province"
                type="text"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400 transition-colors"
                placeholder="如：贵州、云南"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">工艺</label>
              <input
                v-model="form.craft"
                type="text"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400 transition-colors"
                placeholder="如：刺绣、蜡染、织锦"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">材质</label>
              <input
                v-model="form.material"
                type="text"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400 transition-colors"
                placeholder="如：丝绸、棉麻、羊毛"
              />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">纹样</label>
            <input
              v-model="form.pattern"
              type="text"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                     placeholder:text-mogao-ochre-400 transition-colors"
              placeholder="如：龙凤纹、几何纹、花卉纹"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">文化寓意</label>
            <textarea
              v-model="form.cultural_meaning"
              rows="2"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                     placeholder:text-mogao-ochre-400 transition-colors resize-none"
              placeholder="描述该服饰的文化内涵与象征意义"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-mogao-lapis-700 mb-1">详细描述</label>
            <textarea
              v-model="form.description"
              rows="3"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                     placeholder:text-mogao-ochre-400 transition-colors resize-none"
              placeholder="详细介绍该服饰的历史背景、制作工艺和文化价值"
            />
          </div>

          <!-- 提交状态 -->
          <Transition name="fade">
            <p v-if="submitError" class="text-sm text-red-500 bg-red-50 px-4 py-2.5 rounded-mogao">
              {{ submitError }}
            </p>
          </Transition>

          <div class="flex items-center gap-4 pt-2">
            <button
              type="submit"
              :disabled="submitting || !selectedFile"
              class="px-6 py-3 bg-mogao-cinnabar-500 text-white rounded-lg hover:bg-mogao-cinnabar-600
                     transition-colors text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <span v-if="submitting" class="flex items-center gap-2">
                <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                提交中...
              </span>
              <span v-else>提交投稿</span>
            </button>
            <button
              type="button"
              @click="resetForm"
              class="px-4 py-3 text-sm text-mogao-lapis-700 hover:bg-mogao-ochre-100 rounded-lg transition-colors"
            >
              重置
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, type Ref } from 'vue'
import { useRouter, type Router } from 'vue-router'
import { submissionAPI } from '@/services/api'
import { formatFileSize } from '@/utils'

const router: Router = useRouter()

interface SubmitForm {
  name: string
  dynasty: string
  ethnic: string
  craft: string
  material: string
  pattern: string
  cultural_meaning: string
  province: string
  description: string
}

const form: Ref<SubmitForm> = ref({
  name: '',
  dynasty: '',
  ethnic: '',
  craft: '',
  material: '',
  pattern: '',
  cultural_meaning: '',
  province: '',
  description: '',
})

const fileInputRef: Ref<HTMLInputElement | null> = ref(null)
const selectedFile: Ref<File | null> = ref(null)
const imagePreview: Ref<string | null> = ref(null)
const imageError: Ref<string> = ref('')
const dragOver: Ref<boolean> = ref(false)
const submitting: Ref<boolean> = ref(false)
const submitError: Ref<string> = ref('')

function triggerFileInput() {
  fileInputRef.value?.click()
}

function validateFile(file: File): boolean {
  const allowedExts = ['jpg', 'jpeg', 'png', 'webp']
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (!ext || !allowedExts.includes(ext)) {
    imageError.value = '仅支持 JPG、PNG、WebP 格式'
    return false
  }
  if (file.size > 10 * 1024 * 1024) {
    imageError.value = '图片大小不能超过 10MB'
    return false
  }
  return true
}

function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement
  imageError.value = ''
  if (input.files && input.files.length > 0) {
    const file = input.files[0]
    if (!validateFile(file)) {
      selectedFile.value = null
      imagePreview.value = null
      return
    }
    selectedFile.value = file
    imagePreview.value = URL.createObjectURL(file)
  }
}

function onDrop(e: DragEvent) {
  dragOver.value = false
  imageError.value = ''
  const files = e.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (!validateFile(file)) {
      selectedFile.value = null
      imagePreview.value = null
      return
    }
    selectedFile.value = file
    imagePreview.value = URL.createObjectURL(file)
  }
}

function clearImage() {
  selectedFile.value = null
  imagePreview.value = null
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

function resetForm() {
  form.value = { name: '', dynasty: '', ethnic: '', craft: '', material: '', pattern: '', cultural_meaning: '', province: '', description: '' }
  clearImage()
  submitError.value = ''
}

async function handleSubmit() {
  if (!selectedFile.value) {
    imageError.value = '请选择图片文件'
    return
  }

  submitError.value = ''
  submitting.value = true

  const formData = new FormData()
  formData.append('image', selectedFile.value)
  formData.append('name', form.value.name)
  formData.append('dynasty', form.value.dynasty)
  formData.append('ethnic', form.value.ethnic)
  formData.append('craft', form.value.craft)
  formData.append('material', form.value.material)
  formData.append('pattern', form.value.pattern)
  formData.append('cultural_meaning', form.value.cultural_meaning)
  formData.append('province', form.value.province)
  formData.append('description', form.value.description)

  try {
    await submissionAPI.submit(formData)
    router.push('/my-submissions')
  } catch (err: unknown) {
    submitError.value = err instanceof Error ? err.message : '投稿提交失败，请重试'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

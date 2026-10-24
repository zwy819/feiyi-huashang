<template>
  <div class="page-container">
    <div class="mb-8">
      <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">导出报告</h1>
      <p class="mt-2 text-sm text-mogao-ochre-500">将收藏的非遗服饰生成 PDF 报告，附带每件服饰的三维预览图</p>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="animate-pulse space-y-3">
      <div v-for="i in 3" :key="i" class="h-16 bg-mogao-ochre-100 rounded-lg" />
    </div>

    <!-- 无收藏 -->
    <EmptyState
      v-else-if="favorites.length === 0"
      icon="☷"
      title="暂无收藏"
      description="请先收藏非遗服饰后再生成报告"
      :action="{ label: '去浏览', to: '/explore' }"
    />

    <!-- 报告预览 -->
    <div v-else class="space-y-6">
      <!-- 收藏清单 -->
      <div class="card-mogao p-6">
        <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-4">
          报告内容预览（共 {{ favorites.length }} 件）
        </h3>
        <div class="space-y-2">
          <div
            v-for="(fav, i) in favorites"
            :key="fav.id || fav.garment_id"
            class="flex items-center gap-4 px-4 py-3 bg-mogao-ochre-50 rounded-lg"
          >
            <span class="font-serif text-mogao-cinnabar text-lg w-8 text-center">
              {{ i + 1 }}
            </span>
            <div class="flex-1">
              <p class="font-medium text-mogao-lapis-700">{{ fav.garment_name }}</p>
              <p class="text-xs text-mogao-ochre-500">
                {{ dynastyLabel(fav.garment_dynasty) }} · {{ fav.garment_ethnic || '' }} · {{ fav.garment_craft || '' }} · {{ fav.garment_province || '' }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 报告说明 -->
      <div class="card-mogao p-6">
        <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800 mb-3">报告说明</h3>
        <ul class="text-sm text-mogao-lapis-600 space-y-2 list-disc list-inside">
          <li>报告格式：PDF（含封面与每件服饰的专属页面）</li>
          <li>包含内容：服饰三维预览图、名称、朝代、民族、工艺、材质、纹样、文化寓意</li>
          <li>生成日期：{{ formatDate(new Date().toISOString()) }}</li>
        </ul>
      </div>

      <!-- 生成按钮 -->
      <div class="text-center">
        <button
          @click="handleGenerate"
          :disabled="generating"
          class="btn-mogao-primary text-lg px-10 py-3 disabled:opacity-60"
        >
          <span v-if="generating" class="flex items-center gap-2">
            <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            生成中...
          </span>
          <span v-else>生成并下载 PDF 报告</span>
        </button>
        <p v-if="renderProgress" class="mt-3 text-sm text-mogao-ochre-600">{{ renderProgress }}</p>
        <p v-if="generated" class="mt-3 text-sm text-mogao-malachite-600">
          报告已生成，若未自动下载请检查浏览器下载设置
        </p>
      </div>

      <!-- 在线预览 -->
      <div v-if="previewUrl" class="card-mogao p-4">
        <div class="flex items-center justify-between mb-3 px-2">
          <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800">报告预览</h3>
          <a :href="previewUrl" download="非遗服饰收藏报告.pdf" class="btn-mogao-primary text-sm">
            下载 PDF
          </a>
        </div>
        <iframe :src="previewUrl" class="w-full h-[720px] border border-mogao-ochre-200 rounded" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, type Ref } from 'vue'
import * as THREE from 'three'
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js'
import { reportAPI, favoriteAPI, garmentAPI } from '@/services/api'
import EmptyState from '@/components/Common/EmptyState.vue'
import { formatDate } from '@/utils'
import type { ApiResponse } from '@/types/api'
import type { FavoriteItem } from '@/types/favorite'

const favorites: Ref<FavoriteItem[]> = ref([])
const loading: Ref<boolean> = ref(true)
const generating: Ref<boolean> = ref(false)
const generated: Ref<boolean> = ref(false)
const renderProgress: Ref<string> = ref('')
const previewUrl: Ref<string> = ref('')

// "清"→"清代"；"现代/古代"这类已含"代"的原样保留
function dynastyLabel(dynasty?: string): string {
  if (!dynasty) return ''
  return dynasty.endsWith('代') ? dynasty : `${dynasty}代`
}

onMounted(async () => {
  try {
    const res = await favoriteAPI.list()
    const data = res as ApiResponse<{ favorites: FavoriteItem[] }> | { favorites: FavoriteItem[] }
    favorites.value = (data as ApiResponse<{ favorites: FavoriteItem[] }>).data?.favorites
      || (data as { favorites: FavoriteItem[] }).favorites
      || []
  } catch {
    favorites.value = []
  } finally {
    loading.value = false
  }
})

// ============================================
// 离屏渲染：将每件服饰的 3D 模型绘制为平面预览图
// ============================================

interface CaptureRig {
  renderer: THREE.WebGLRenderer
  scene: THREE.Scene
  camera: THREE.PerspectiveCamera
}

/** 创建与详情页观察器同风格的离屏渲染器（灯光、色调映射保持一致） */
function createCaptureRig(): CaptureRig {
  const renderer = new THREE.WebGLRenderer({ antialias: true, preserveDrawingBuffer: true })
  renderer.setSize(900, 900)
  renderer.setPixelRatio(1)
  renderer.toneMapping = THREE.ACESFilmicToneMapping
  renderer.toneMappingExposure = 1.2
  renderer.outputColorSpace = THREE.SRGBColorSpace

  const scene = new THREE.Scene()
  // 暖纸色背景（与报告纸色呼应），JPEG 压缩友好
  scene.background = new THREE.Color('#F7F0E5')

  scene.add(new THREE.AmbientLight(0xffffff, 1.5))
  scene.add(new THREE.HemisphereLight(0xddeeff, 0x889966, 0.8))
  const keyLight = new THREE.DirectionalLight(0xffffff, 1.8)
  keyLight.position.set(5, 5, 5)
  scene.add(keyLight)
  const fillLight = new THREE.DirectionalLight(0xffffff, 0.6)
  fillLight.position.set(-3, 2, -3)
  scene.add(fillLight)

  const camera = new THREE.PerspectiveCamera(35, 1, 0.1, 2000)
  return { renderer, scene, camera }
}

/** 释放模型占用的几何体 / 材质 / 纹理显存 */
function disposeModel(model: THREE.Object3D): void {
  model.traverse((child) => {
    const mesh = child as THREE.Mesh
    if (!mesh.isMesh) return
    mesh.geometry?.dispose()
    const materials = Array.isArray(mesh.material) ? mesh.material : [mesh.material]
    for (const mat of materials) {
      if (!mat) continue
      for (const value of Object.values(mat)) {
        if (value && (value as THREE.Texture).isTexture) (value as THREE.Texture).dispose()
      }
      mat.dispose()
    }
  })
}

/** 加载单件服饰模型并渲染为 900×900 JPEG dataURL；失败返回 null（报告页自动降级为占位面板） */
function captureModel(rig: CaptureRig, modelId: number): Promise<string | null> {
  return new Promise((resolve) => {
    new GLTFLoader().load(
      `/api/models/${modelId}/file`,
      (gltf) => {
        try {
          const model = gltf.scene
          const box = new THREE.Box3().setFromObject(model)
          const center = box.getCenter(new THREE.Vector3())
          model.position.sub(center)
          const size = box.getSize(new THREE.Vector3())
          const maxDim = Math.max(size.x, size.y, size.z) || 1
          rig.scene.add(model)

          // 右前上方的 3/4 视角，按包围盒自动取距，留 28% 边距
          const dist = (maxDim / 2 / Math.tan((35 * Math.PI) / 360)) * 1.28
          rig.camera.position.set(dist * 0.62, dist * 0.42, dist * 0.78)
          rig.camera.lookAt(0, 0, 0)
          rig.renderer.render(rig.scene, rig.camera)
          const dataURL = rig.renderer.domElement.toDataURL('image/jpeg', 0.9)

          rig.scene.remove(model)
          disposeModel(model)
          resolve(dataURL)
        } catch {
          resolve(null)
        }
      },
      undefined,
      () => resolve(null),
    )
  })
}

async function handleGenerate(): Promise<void> {
  generating.value = true
  generated.value = false
  let rig: CaptureRig | null = null
  try {
    // 1. 逐件离屏渲染预览图
    const ids: number[] = []
    const images: Record<string, string> = {}
    for (let i = 0; i < favorites.value.length; i++) {
      const fav = favorites.value[i]
      const gid = fav.garment_id
      if (!gid) continue
      ids.push(gid)
      renderProgress.value = `正在生成预览图 (${i + 1}/${favorites.value.length})：${fav.garment_name || ''}`
      await new Promise((r) => setTimeout(r, 0)) // 让进度文本有机会渲染
      try {
        const res = (await garmentAPI.getById(gid)) as unknown as ApiResponse<{ model_id: number | null }>
        const modelId = res.data?.model_id
        if (!modelId) continue
        rig = rig ?? createCaptureRig()
        const dataURL = await captureModel(rig, modelId)
        if (dataURL) images[String(gid)] = dataURL
      } catch {
        // 单件失败不影响整体报告，该页使用占位面板
      }
    }

    // 2. 提交生成 PDF
    renderProgress.value = '正在生成 PDF 报告...'
    const response = (await reportAPI.generate({ ids, images })) as unknown as Blob

    // 3. 触发下载 + 在线预览
    const blob = new Blob([response], { type: 'application/pdf' })
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    const url = window.URL.createObjectURL(blob)
    previewUrl.value = url
    const link = document.createElement('a')
    link.href = url
    link.download = `非遗服饰收藏报告_${new Date().toISOString().slice(0, 10)}.pdf`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    generated.value = true
  } catch (err: unknown) {
    console.error('报告生成失败:', err)
  } finally {
    rig?.renderer.dispose()
    renderProgress.value = ''
    generating.value = false
  }
}

onBeforeUnmount(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
})
</script>

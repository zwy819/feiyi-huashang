<template>
  <div ref="containerRef" class="viewport-3d relative" :style="{ height: `${height}px` }">
    <!-- 无模型路径占位 -->
    <div
      v-if="!modelPath"
      class="absolute inset-0 flex items-center justify-center"
    >
      <div class="text-center">
        <span class="text-6xl block mb-3">&#x1f457;</span>
        <p class="text-sm text-mogao-ochre-500">3D 服饰展示</p>
        <p class="text-xs text-mogao-ochre-400 mt-1">暂无 3D 模型</p>
      </div>
    </div>

    <!-- WebGL 上下文丢失 -->
    <div
      v-else-if="webglLost"
      class="absolute inset-0 flex items-center justify-center"
    >
      <div class="text-center p-8">
        <span class="text-4xl text-mogao-gold block mb-3">⚠</span>
        <p class="text-sm text-mogao-ochre-600">3D 渲染引擎暂时中断</p>
        <p class="text-xs text-mogao-ochre-500 mt-1">正在尝试恢复…</p>
        <button
          @click="retryLoad"
          class="btn-mogao-secondary text-xs mt-4"
        >
          手动重试
        </button>
      </div>
    </div>

    <!-- 加载错误 -->
    <div
      v-else-if="loadError"
      class="absolute inset-0 flex items-center justify-center"
    >
      <div class="text-center p-8">
        <span class="text-4xl text-mogao-cinnabar-400 block mb-3">⚠</span>
        <p class="text-sm text-mogao-cinnabar-600">3D 模型加载失败</p>
        <p class="text-xs text-mogao-ochre-500 mt-1 max-w-xs mx-auto break-all">{{ loadError }}</p>
        <button
          @click="retryLoad"
          class="btn-mogao-secondary text-xs mt-4"
        >
          重新加载
        </button>
      </div>
    </div>

    <!-- 加载中指示器 -->
    <div
      v-if="loading && modelPath && !loadError && !webglLost"
      class="absolute inset-0 flex items-center justify-center pointer-events-none z-10"
    >
      <div class="flex flex-col items-center gap-2">
        <div class="w-8 h-8 border-2 border-mogao-cinnabar border-t-transparent rounded-full animate-spin" />
        <span class="text-xs text-mogao-ochre-600 whitespace-nowrap">加载模型中…</span>
      </div>
    </div>

    <!-- 背景选择器 -->
    <div
      v-if="showBackgroundPicker && modelPath && !loadError && !webglLost"
      class="absolute bottom-3 left-1/2 -translate-x-1/2 z-20 flex items-center gap-2 px-3 py-2 rounded-full transition-all duration-300"
      style="background: rgba(60, 50, 40, 0.45); backdrop-filter: blur(12px); border: 1px solid rgba(255,255,255,0.1);"
    >
      <button
        v-for="bg in BACKGROUND_OPTIONS"
        :key="bg.value"
        @click="currentBg = bg.value"
        :title="bg.label"
        class="w-7 h-7 rounded-full border-2 transition-all duration-200 hover:scale-115 flex-shrink-0"
        :class="currentBg === bg.value
          ? 'border-white shadow-lg scale-110'
          : 'border-white/20 opacity-65 hover:opacity-90 hover:border-white/50'"
        :style="{ background: bg.preview }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick, type Ref } from 'vue'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { loadModel } from './ModelLoader'

interface Props {
  modelPath?: string
  autoRotate?: boolean
  height?: number
  /** 背景主题: 'white' | 'forest' | 'seaside' | 'park' */
  background?: string
  /** 是否显示背景选择器 */
  showBackgroundPicker?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelPath: '',
  autoRotate: true,
  height: 480,
  background: 'white',
  showBackgroundPicker: true,
})

interface HTMLElementWithCleanup extends HTMLCanvasElement {
  _webglCleanup?: () => void
}

// ============================================
// 背景主题配置
// ============================================
interface BgTheme {
  value: string
  label: string
  preview: string          // CSS 渐变预览色
  skyColor: string         // 半球光天空色
  groundColor: string      // 半球光地面色
  bgColor?: string         // 纯色背景
  gradientTop?: string     // 渐变纹理顶部色
  gradientBottom?: string  // 渐变纹理底部色
}

const BACKGROUND_OPTIONS: BgTheme[] = [
  {
    value: 'white', label: '素白', preview: 'linear-gradient(135deg, #F5F0E8 0%, #E8E0D5 100%)',
    skyColor: '#ddeeff', groundColor: '#889966',
    bgColor: '#F5F0E8',
  },
  {
    value: 'forest', label: '竹林', preview: 'linear-gradient(135deg, #2D5A3D 0%, #1A3020 100%)',
    skyColor: '#3a5c3f', groundColor: '#2a3a1a',
    gradientTop: '#2D5A3D', gradientBottom: '#1A3020',
  },
  {
    value: 'seaside', label: '海岸', preview: 'linear-gradient(135deg, #7BA3B8 0%, #3D6878 100%)',
    skyColor: '#9bc3d8', groundColor: '#5a7a90',
    gradientTop: '#7BA3B8', gradientBottom: '#3D6878',
  },
  {
    value: 'park', label: '园林', preview: 'linear-gradient(135deg, #C4A96A 0%, #5A7D3A 100%)',
    skyColor: '#d4b97a', groundColor: '#6a8a4a',
    gradientTop: '#C4A96A', gradientBottom: '#5A7D3A',
  },
]

const containerRef: Ref<HTMLElement | null> = ref(null)
const loading: Ref<boolean> = ref(true)
const loadError: Ref<string | null> = ref(null)
const webglLost: Ref<boolean> = ref(false)
const currentBg: Ref<string> = ref(props.background)

let renderer: THREE.WebGLRenderer | null = null
let scene: THREE.Scene | null = null
let camera: THREE.PerspectiveCamera | null = null
let controls: OrbitControls | null = null
let animationId: number | null = null
let modelCleanup: (() => void) | null = null
let currentModel: THREE.Group | null = null
let groundMesh: THREE.Mesh | null = null
let hemiLight: THREE.HemisphereLight | null = null
let bgTexture: THREE.Texture | THREE.CanvasTexture | null = null
let resizeObserver: ResizeObserver | null = null

function initScene(): void {
  if (!containerRef.value) return

  const container = containerRef.value
  const width = container.clientWidth
  const height = container.clientHeight

  // Renderer
  renderer = new THREE.WebGLRenderer({
    antialias: true,
    alpha: true,
    powerPreference: 'high-performance',
    failIfMajorPerformanceCaveat: false,
  })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  renderer.setSize(width, height)
  renderer.shadowMap.enabled = true
  renderer.shadowMap.type = THREE.PCFSoftShadowMap
  // 色调映射 + 曝光（提亮暗部，保留高光细节）
  renderer.toneMapping = THREE.ACESFilmicToneMapping
  renderer.toneMappingExposure = 1.2
  renderer.outputColorSpace = THREE.SRGBColorSpace
  container.appendChild(renderer.domElement)

  // Scene
  scene = new THREE.Scene()

  // Camera
  camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 50)
  camera.position.set(0, 1.5, 4)

  // ============================================
  // 灯光系统（方案一：增强光照 + 方案二：Tone Mapping）
  // ============================================

  // 环境光 — 整体基础亮度
  const ambientLight = new THREE.AmbientLight(0xffffff, 1.5)
  scene.add(ambientLight)

  // 半球光 — 模拟天空/地面漫反射，消除死黑暗面
  const hemi = new THREE.HemisphereLight(
    0xddeeff, // 天空色（冷白）
    0x889966, // 地面色（暖土色）
    0.8,      // 强度
  )
  hemiLight = hemi
  scene.add(hemi)

  // 主方向光（Key Light）— 右前方 45° 打光
  const keyLight = new THREE.DirectionalLight(0xffffff, 1.8)
  keyLight.position.set(5, 5, 5)
  keyLight.castShadow = true
  keyLight.shadow.mapSize.width = 1024
  keyLight.shadow.mapSize.height = 1024
  scene.add(keyLight)

  // 补光（Fill Light）— 左下方弱光，减轻阴影
  const fillLight = new THREE.DirectionalLight(0xffffff, 0.6)
  fillLight.position.set(-3, 2, -3)
  scene.add(fillLight)

  // Ground plane for contact shadows
  const groundGeometry = new THREE.PlaneGeometry(10, 10)
  const groundMaterial = new THREE.ShadowMaterial({ opacity: 0.4 })
  const ground = new THREE.Mesh(groundGeometry, groundMaterial)
  ground.rotation.x = -Math.PI / 2
  ground.position.y = -1.5
  ground.receiveShadow = true
  scene.add(ground)
  groundMesh = ground

  // Orbit Controls
  controls = new OrbitControls(camera, renderer.domElement)
  controls.enablePan = true
  controls.enableZoom = true
  controls.enableRotate = true
  controls.autoRotate = props.autoRotate
  controls.autoRotateSpeed = 1.5
  controls.minDistance = 2
  controls.maxDistance = 8
  controls.maxPolarAngle = Math.PI / 1.5
  controls.update()

  // 应用初始背景
  applyBackground(currentBg.value)

  // WebGL context events
  const canvas = renderer.domElement as HTMLElementWithCleanup
  const onContextLost = (event: Event) => {
    event.preventDefault()
    console.warn('[ModelViewer] WebGL 上下文丢失')
    webglLost.value = true
  }
  const onContextRestored = () => {
    console.log('[ModelViewer] WebGL 上下文已恢复')
    webglLost.value = false
    retryLoad()
  }
  canvas.addEventListener('webglcontextlost', onContextLost)
  canvas.addEventListener('webglcontextrestored', onContextRestored)
  canvas._webglCleanup = (): void => {
    canvas.removeEventListener('webglcontextlost', onContextLost)
    canvas.removeEventListener('webglcontextrestored', onContextRestored)
  }

  // Resize handler
  resizeObserver = new ResizeObserver(() => {
    if (renderer && container) {
      const w = container.clientWidth
      const h = container.clientHeight
      renderer.setSize(w, h)
      if (camera) {
        camera.aspect = w / h
        camera.updateProjectionMatrix()
      }
    }
  })
  resizeObserver.observe(container)

  // Animation loop
  function animate(): void {
    animationId = requestAnimationFrame(animate)
    if (controls) controls.update()
    if (renderer && scene && camera) {
      renderer.render(scene, camera)
    }
  }
  animate()
}

function loadModelIntoScene(url: string): void {
  if (!url || !scene) return

  // 清除旧模型
  if (currentModel) {
    scene.remove(currentModel)
    currentModel = null
  }

  // 取消之前的加载
  if (modelCleanup) {
    modelCleanup()
    modelCleanup = null
  }

  loading.value = true
  loadError.value = null

  modelCleanup = loadModel(
    url,
    // onLoad
    (modelScene: THREE.Group) => {
      if (!scene) return
      currentModel = modelScene
      scene.add(modelScene)

      // 模型已居中于原点，更新旋转中心和地面阴影位置
      const box = new THREE.Box3().setFromObject(modelScene)
      const modelCenter = box.getCenter(new THREE.Vector3())
      const modelSize = box.getSize(new THREE.Vector3())

      if (controls) {
        controls.target.copy(modelCenter)
        controls.update()
      }

      // 地面阴影跟随模型底部
      if (groundMesh) {
        groundMesh.position.y = modelCenter.y - modelSize.y / 2 - 0.05
      }

      loading.value = false
    },
    // onProgress (not used for now)
    undefined,
    // onError
    (msg: string) => {
      loading.value = false
      loadError.value = msg
    },
  )
}

function retryLoad(): void {
  webglLost.value = false
  loadError.value = null
  loading.value = true
  loadModelIntoScene(props.modelPath)
}

// ============================================
// 背景纹理生成（Canvas2D 程序化渐变）
// ============================================

/**
 * 生成 256×256 纵向渐变纹理
 * 纹理会被映射为 scene.background，AlwaysStencil 使其自动适配视口
 */
function generateBackgroundTexture(topColor: string, bottomColor: string): THREE.CanvasTexture {
  const size = 256
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')!

  const gradient = ctx.createLinearGradient(0, 0, 0, size)
  gradient.addColorStop(0, topColor)
  gradient.addColorStop(1, bottomColor)
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, size, size)

  const tex = new THREE.CanvasTexture(canvas)
  tex.mapping = THREE.EquirectangularReflectionMapping
  tex.colorSpace = THREE.SRGBColorSpace
  tex.minFilter = THREE.LinearMipmapLinearFilter
  tex.magFilter = THREE.LinearFilter
  tex.generateMipmaps = true
  return tex
}

/**
 * 应用背景主题：纯色直接设 Color，渐变用 Canvas2D 纹理
 * 同时调整半球光颜色与背景协调
 */
function applyBackground(theme: string): void {
  if (!scene) return

  const cfg = BACKGROUND_OPTIONS.find(b => b.value === theme)
  if (!cfg) return

  // 清理旧纹理
  if (bgTexture) {
    bgTexture.dispose()
    bgTexture = null
  }

  // 设置场景背景
  if (cfg.bgColor) {
    // 纯色
    scene.background = new THREE.Color(cfg.bgColor)
  } else if (cfg.gradientTop && cfg.gradientBottom) {
    // 渐变纹理
    bgTexture = generateBackgroundTexture(cfg.gradientTop, cfg.gradientBottom)
    scene.background = bgTexture
  }

  // 同步半球光颜色
  if (hemiLight) {
    hemiLight.color.set(cfg.skyColor)
    hemiLight.groundColor.set(cfg.groundColor)
  }
}

// 监听背景切换
watch(currentBg, (newBg: string) => {
  applyBackground(newBg)
})

function disposeScene(): void {
  // 取消模型加载
  if (modelCleanup) {
    modelCleanup()
    modelCleanup = null
  }

  // 停止动画循环
  if (animationId) {
    cancelAnimationFrame(animationId)
    animationId = null
  }

  // 清理模型
  if (currentModel && scene) {
    scene.remove(currentModel)
    currentModel.traverse((child: THREE.Object3D) => {
      const mesh = child as THREE.Mesh
      if (mesh.geometry) mesh.geometry.dispose()
      if (mesh.material) {
        if (Array.isArray(mesh.material)) {
          mesh.material.forEach((m: THREE.Material) => m.dispose())
        } else {
          mesh.material.dispose()
        }
      }
    })
    currentModel = null
  }

  // 清理控制器
  if (controls) {
    controls.dispose()
    controls = null
  }

  // 清理渲染器
  if (renderer) {
    const canvas = renderer.domElement as HTMLElementWithCleanup
    if (canvas._webglCleanup) {
      canvas._webglCleanup()
      delete canvas._webglCleanup
    }
    renderer.dispose()
    if (canvas.parentNode) {
      canvas.parentNode.removeChild(canvas)
    }
    renderer = null
  }

  // 清理背景纹理
  if (bgTexture) {
    bgTexture.dispose()
    bgTexture = null
  }

  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }

  scene = null
  camera = null
}

// 监听 modelPath 变化
watch(
  () => props.modelPath,
  (newPath: string) => {
    if (newPath && scene) {
      loadModelIntoScene(newPath)
    }
  },
)

onMounted(async () => {
  await nextTick()
  if (props.modelPath && containerRef.value) {
    initScene()
    loadModelIntoScene(props.modelPath)
  } else if (!props.modelPath) {
    loading.value = false
  }
})

onUnmounted(() => {
  disposeScene()
})
</script>

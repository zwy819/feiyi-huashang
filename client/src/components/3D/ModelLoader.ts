import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js'
import * as THREE from 'three'

/** 大模型加载超时时间（毫秒），90MB 级别文件需要较长时间 */
const LOAD_TIMEOUT_MS = 180000 // 3 分钟

/**
 * loadModel — GLTF/GLB 模型加载器
 *
 * 纯函数模块，不依赖任何框架。返回 cleanup 函数用于取消加载。
 *
 * @param url        - 模型 URL
 * @param onLoad     - 加载完成回调
 * @param onProgress - 进度回调
 * @param onError    - 错误回调
 * @returns cleanup  - 调用以取消加载
 */
export function loadModel(
  url: string,
  onLoad: (scene: THREE.Group) => void,
  onProgress?: (percent: number) => void,
  onError?: (message: string) => void,
): () => void {
  if (!url) return () => {}

  let cancelled = false
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  const loader = new GLTFLoader()

  console.log(`[ModelLoader] 开始加载: ${url}`)

  // 超时保护
  timeoutId = setTimeout(() => {
    if (!cancelled) {
      const msg = `模型加载超时（>${LOAD_TIMEOUT_MS / 1000}秒），文件可能过大`
      console.error(`[ModelLoader] ${msg}: ${url}`)
      if (onError) onError(msg)
    }
  }, LOAD_TIMEOUT_MS)

  loader.load(
    url,
    // onLoad — 加载成功
    (gltf) => {
      if (cancelled) return
      if (timeoutId) clearTimeout(timeoutId)
      console.log(`[ModelLoader] 加载完成: ${url}`)

      try {
        // 计算包围盒，将模型几何中心对齐到原点
        const box = new THREE.Box3().setFromObject(gltf.scene)
        const center = box.getCenter(new THREE.Vector3())

        gltf.scene.position.set(-center.x, -center.y, -center.z)

        // 遍历子对象启用阴影
        gltf.scene.traverse((child) => {
          if (child instanceof THREE.Mesh) {
            child.castShadow = true
            child.receiveShadow = true
          }
        })
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : String(err)
        console.error('[ModelLoader] 模型处理出错:', msg)
      }

      if (onLoad) onLoad(gltf.scene)
    },
    // onProgress
    (event) => {
      if (cancelled) return
      if (event.total > 0) {
        const pct = ((event.loaded / event.total) * 100).toFixed(1)
        console.log(`[ModelLoader] ${url} 加载进度: ${pct}%`)
        if (onProgress) onProgress(Number(pct))
      }
    },
    // onError
    (error: unknown) => {
      if (cancelled) return
      if (timeoutId) clearTimeout(timeoutId)
      const msg =
        error instanceof Error
          ? error.message
          : typeof error === 'string'
            ? error
            : '模型加载失败'
      console.error(`[ModelLoader] 加载失败: ${url} — ${msg}`)
      if (onError) onError(msg)
    },
  )

  return () => {
    cancelled = true
    if (timeoutId) clearTimeout(timeoutId)
  }
}

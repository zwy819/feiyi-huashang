import {
  computed,
  ref,
  watch,
  watchEffect,
  onUnmounted,
  inject,
  type Ref,
  type App,
} from 'vue'
import type { User } from '@/types/user'

/**
 * useFetch - 通用数据获取 Composable
 */
export function useFetch<T>(
  fetcher: () => Promise<T>,
  deps: Ref[] = [],
  immediate = true,
) {
  const data = ref<T | null>(null) as Ref<T | null>
  const loading = ref(immediate)
  const error = ref<string | null>(null)

  const execute = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await fetcher()
      data.value = result
    } catch (err: unknown) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  if (immediate) {
    watchEffect(execute)
  }

  return { data, loading, error, refetch: execute }
}

/**
 * useAuth - 用户认证状态 Composable
 */
export function useAuth() {
  const user = ref<User | null>(null) as Ref<User | null>

  // 从 localStorage 恢复登录状态
  const storedUser = localStorage.getItem('user')
  if (storedUser) {
    try {
      user.value = JSON.parse(storedUser) as User
    } catch {
      user.value = null
    }
  }

  const login = (token: string, userData: User) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(userData))
    user.value = userData
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    user.value = null
  }

  return {
    user,
    isAuthenticated: computed(() => !!user.value),
    login,
    logout,
  }
}

/**
 * useDebounce - 防抖 Composable
 */
export function useDebounce<T>(valueRef: Ref<T>, delay = 300): Ref<T> {
  const debouncedValue = ref(valueRef.value) as Ref<T>

  let timer: ReturnType<typeof setTimeout> | null = null
  watch(valueRef, (val) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      debouncedValue.value = val
    }, delay)
  })

  onUnmounted(() => {
    if (timer) clearTimeout(timer)
  })

  return debouncedValue
}

// ===== Toast Types =====

interface ToastItem {
  id: number
  message: string
  type: string
}

type ShowToastFn = (message: string, type?: string) => void

/**
 * useToast - 获取 showToast 方法
 */
export function useToast(): { showToast: ShowToastFn } {
  const injected = inject<ShowToastFn | null>('showToast', null)
  if (injected) return { showToast: injected }

  const fallback: ShowToastFn =
    typeof window !== 'undefined' && window.__showToast
      ? (msg, type) => window.__showToast!(msg, type)
      : () => {}
  return { showToast: fallback }
}

// ===== Toast Plugin =====

let toastId = 0

/**
 * Toast Vue Plugin — 提供全局 showToast 方法
 */
export const toastPlugin = {
  install(app: App) {
    const toasts = ref<ToastItem[]>([])

    const removeToast = (id: number) => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }

    const showToast: ShowToastFn = (message, type = 'error') => {
      const id = ++toastId
      toasts.value = [...toasts.value.slice(-4), { id, message, type }]
      setTimeout(() => removeToast(id), 4000)
    }

    // 挂载到 window 供 api.ts 拦截器使用
    if (typeof window !== 'undefined') {
      window.__showToast = showToast
    }

    app.provide('showToast', showToast)
    app.config.globalProperties.$toast = showToast

    // 注册全局 Toast 容器组件
    app.provide('toasts', toasts)
    app.provide('removeToast', removeToast)
  },
}

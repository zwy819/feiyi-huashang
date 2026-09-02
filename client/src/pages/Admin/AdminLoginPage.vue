<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-mogao-lapis-950 via-mogao-lapis-900 to-mogao-lapis-800">
    <!-- 装饰元素 -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-10 left-1/4 w-72 h-72 rounded-full bg-mogao-cinnabar/5 blur-3xl" />
      <div class="absolute bottom-20 right-1/4 w-96 h-96 rounded-full bg-mogao-gold/5 blur-3xl" />
    </div>

    <!-- 登录卡片 -->
    <div class="relative w-full max-w-md mx-4">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="font-serif text-5xl text-mogao-cinnabar-400 mb-3">非</div>
        <h1 class="font-serif text-2xl font-bold text-white tracking-wider">非遗华裳</h1>
        <div class="w-12 h-0.5 bg-mogao-cinnabar-500 mx-auto mt-3" />
        <p class="text-mogao-leadwhite-300/60 text-sm mt-3">后台管理系统</p>
      </div>

      <!-- 表单卡片 -->
      <div class="bg-white/95 backdrop-blur rounded-xl shadow-2xl p-8 border border-white/10">
        <form @submit.prevent="handleLogin" class="space-y-5">
          <!-- 用户名 -->
          <div>
            <label class="block text-xs text-mogao-lapis-600 mb-1.5 font-medium">管理员账号</label>
            <input
              v-model="form.username"
              type="text"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                     placeholder:text-mogao-ochre-400 transition-colors"
              placeholder="输入管理员用户名"
              autocomplete="username"
              :disabled="loading"
            />
          </div>

          <!-- 密码 -->
          <div>
            <label class="block text-xs text-mogao-lapis-600 mb-1.5 font-medium">密码</label>
            <input
              v-model="form.password"
              type="password"
              class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                     focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                     placeholder:text-mogao-ochre-400 transition-colors"
              placeholder="输入密码"
              autocomplete="current-password"
              :disabled="loading"
            />
          </div>

          <!-- 错误提示 -->
          <Transition name="fade">
            <p
              v-if="error"
              class="text-xs text-mogao-cinnabar bg-mogao-cinnabar-50 px-3.5 py-2.5 rounded-mogao"
            >
              {{ error }}
            </p>
          </Transition>

          <!-- 提交按钮 -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-mogao-lapis-800 text-white rounded-lg hover:bg-mogao-lapis-900
                   transition-colors text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed"
          >
            <span v-if="loading" class="flex items-center justify-center gap-2">
              <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              验证中...
            </span>
            <span v-else>登录后台</span>
          </button>
        </form>
      </div>

      <!-- 底部链接 -->
      <p class="text-center mt-6 text-sm text-white/40">
        <router-link to="/" class="hover:text-white/70 transition-colors">← 返回前台</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { useRouter, type Router } from 'vue-router'
import { useAuth } from '@/composables'
import { authAPI } from '@/services/api'
import type { User } from '@/types/user'

interface LoginForm {
  username: string
  password: string
}

const router: Router = useRouter()
const { isAuthenticated, login, user } = useAuth()

const form: Ref<LoginForm> = ref({ username: '', password: '' })
const error: Ref<string> = ref('')
const loading: Ref<boolean> = ref(false)

// 已登录管理员直接跳转后台
onMounted(() => {
  if (isAuthenticated.value && user.value?.role === 'admin') {
    router.replace('/admin')
  }
})

async function handleLogin(): Promise<void> {
  error.value = ''

  if (!form.value.username || !form.value.password) {
    error.value = '请输入管理员账号和密码'
    return
  }

  loading.value = true
  try {
    const result = await authAPI.login({
      username: form.value.username,
      password: form.value.password,
    })
    const data = result?.data || result

    // 检查是否为管理员角色
    if (data.role !== 'admin') {
      error.value = '该账号无管理员权限，请使用管理员账号登录'
      return
    }

    const userData: User = {
      id: data.id,
      username: data.username,
      email: data.email,
      role: data.role || 'user',
      avatar_url: '',
      created_at: '',
    }
    login(data.token, userData)
    router.replace('/admin')
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : '登录失败，请重试'
  } finally {
    loading.value = false
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

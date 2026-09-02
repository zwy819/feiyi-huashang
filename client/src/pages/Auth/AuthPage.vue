<template>
  <div class="min-h-screen flex">
    <!-- ===== 左侧：品牌展示区 ===== -->
    <div class="hidden md:flex md:w-1/2 bg-gradient-to-br from-mogao-lapis-900 via-mogao-lapis-800 to-mogao-cinnabar-800 relative overflow-hidden">
      <!-- 装饰元素 -->
      <div class="absolute inset-0 opacity-10">
        <div class="absolute top-20 left-10 w-64 h-64 rounded-full bg-mogao-gold blur-3xl" />
        <div class="absolute bottom-20 right-10 w-96 h-96 rounded-full bg-mogao-cinnabar blur-3xl" />
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-80 h-80 rounded-full bg-mogao-malachite blur-3xl" />
      </div>

      <div class="relative flex flex-col justify-center items-center w-full p-12 text-white">
        <div class="text-center">
          <div class="font-serif text-8xl mb-6">非</div>
          <h1 class="font-serif text-4xl font-bold tracking-widest mb-4">非遗华裳</h1>
          <div class="w-16 h-0.5 bg-mogao-gold mx-auto mb-6" />
          <p class="text-mogao-leadwhite-300 text-base leading-relaxed max-w-sm">
            以数字技术重现千年织造智慧，探索中国非物质文化遗产服饰的纹样之美、工艺之精、文化之魂
          </p>
        </div>

        <div class="absolute bottom-12 text-center text-mogao-leadwhite-400 text-xs">
          中国非物质文化遗产服饰 3D 数字博物馆
        </div>
      </div>
    </div>

    <!-- ===== 右侧：表单区 ===== -->
    <div class="flex-1 flex items-center justify-center bg-mogao-leadwhite-100 px-4 sm:px-6 lg:px-8">
      <div class="w-full max-w-md">
        <!-- 移动端 Logo -->
        <div class="md:hidden text-center mb-8">
          <div class="font-serif text-5xl text-mogao-cinnabar mb-2">非</div>
          <h2 class="font-serif text-2xl font-bold text-mogao-lapis-800">非遗华裳</h2>
        </div>

        <div class="bg-white rounded-lg shadow-mogao-lg border border-mogao-ochre-200 p-8">
          <!-- 选项卡 -->
          <div class="flex mb-6 border-b border-mogao-ochre-200">
            <button
              v-for="t in ['login', 'register']"
              :key="t"
              @click="switchTab(t)"
              :class="[
                'flex-1 pb-3 text-sm font-medium transition-all',
                tab === t
                  ? 'text-mogao-cinnabar border-b-2 border-mogao-cinnabar'
                  : 'text-mogao-ochre-500 hover:text-mogao-lapis-700',
              ]"
            >
              {{ t === 'login' ? '登录' : '注册' }}
            </button>
          </div>

          <!-- 表单 -->
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <!-- 用户名 -->
            <div>
              <label class="block text-xs text-mogao-lapis-600 mb-1.5">用户名</label>
              <input
                v-model="form.username"
                type="text"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400"
                placeholder="输入用户名"
                autocomplete="username"
              />
            </div>

            <!-- 邮箱（仅注册时显示） -->
            <Transition name="slide-fade">
              <div v-if="tab === 'register'">
                <label class="block text-xs text-mogao-lapis-600 mb-1.5">邮箱</label>
                <input
                  v-model="form.email"
                  type="email"
                  class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                         focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                         placeholder:text-mogao-ochre-400"
                  placeholder="输入邮箱"
                  autocomplete="email"
                />
              </div>
            </Transition>

            <!-- 密码 -->
            <div>
              <label class="block text-xs text-mogao-lapis-600 mb-1.5">密码</label>
              <input
                v-model="form.password"
                type="password"
                class="w-full px-3.5 py-2.5 text-sm border border-mogao-ochre-300 rounded-mogao
                       focus:outline-none focus:border-mogao-cinnabar focus:ring-1 focus:ring-mogao-cinnabar-300
                       placeholder:text-mogao-ochre-400"
                :placeholder="tab === 'register' ? '至少 6 位' : '输入密码'"
                :autocomplete="tab === 'login' ? 'current-password' : 'new-password'"
              />
            </div>

            <!-- 错误提示 -->
            <Transition name="fade">
              <p v-if="error" class="text-xs text-mogao-cinnabar bg-mogao-cinnabar-50 px-3.5 py-2.5 rounded-mogao">
                {{ error }}
              </p>
            </Transition>

            <!-- 提交按钮 -->
            <button
              type="submit"
              :disabled="loading"
              class="w-full btn-mogao-primary py-3 disabled:opacity-60"
            >
              <span v-if="loading" class="flex items-center justify-center gap-2">
                <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                处理中...
              </span>
              <span v-else>{{ tab === 'login' ? '登录' : '注册' }}</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue'
import { useRouter, type Router } from 'vue-router'
import { useAuth } from '@/composables'
import { authAPI } from '@/services/api'
import type { LoginForm, RegisterForm, User } from '@/types/user'

interface AuthForm {
  username: string
  email: string
  password: string
}

const router: Router = useRouter()
const { isAuthenticated, login } = useAuth()

const tab: Ref<'login' | 'register'> = ref('login')
const form: Ref<AuthForm> = ref({ username: '', email: '', password: '' })
const error: Ref<string> = ref('')
const loading: Ref<boolean> = ref(false)

// 已登录用户直接跳转首页
onMounted(() => {
  if (isAuthenticated.value) {
    router.replace('/')
  }
})

function switchTab(t: 'login' | 'register'): void {
  tab.value = t
  error.value = ''
}

async function handleSubmit(): Promise<void> {
  error.value = ''

  if (!form.value.username || !form.value.password) {
    error.value = '请填写用户名和密码'
    return
  }
  if (tab.value === 'register' && !form.value.email) {
    error.value = '请填写邮箱'
    return
  }

  loading.value = true
  try {
    if (tab.value === 'login') {
      const result = await authAPI.login({ username: form.value.username, password: form.value.password })
      const data = result?.data || result
      login(data.token, { id: data.id, username: data.username, email: data.email, role: data.role || 'user' } as User)
    } else {
      const result = await authAPI.register({ username: form.value.username, email: form.value.email, password: form.value.password })
      const data = result?.data || result
      login(data.token, { id: data.id, username: data.username, email: data.email, role: data.role || 'user' } as User)
    }
    router.replace('/')
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : '操作失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.slide-fade-enter-active {
  transition: all 0.25s ease;
}
.slide-fade-leave-active {
  transition: all 0.2s ease;
}
.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  max-height: 0;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

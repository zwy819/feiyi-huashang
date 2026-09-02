<template>
  <header class="sticky top-0 z-50 bg-white/95 backdrop-blur-sm border-b border-mogao-ochre-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <router-link
          to="/"
          class="flex items-center gap-2 font-serif text-xl font-bold text-mogao-cinnabar hover:text-mogao-cinnabar-600 transition-colors"
        >
          <img
            src="/logo.png"
            alt="非遗华裳"
            class="h-10 w-auto object-contain"
          />
          <span>非遗华裳</span>
        </router-link>

        <!-- 桌面导航 -->
        <nav class="hidden md:flex items-center gap-1">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            :exact="item.to === '/'"
            class="px-4 py-2 rounded-mogao text-sm font-medium transition-all duration-200"
            active-class="bg-mogao-cinnabar text-white shadow-mogao"
            :class="$route.path === item.to
              ? ''
              : 'text-mogao-lapis-700 hover:bg-mogao-ochre-50 hover:text-mogao-cinnabar'"
          >
            {{ item.label }}
          </router-link>

          <!-- 管理入口（仅管理员可见） -->
          <router-link
            v-if="isAdmin"
            to="/admin"
            class="px-4 py-2 rounded-mogao text-sm font-medium transition-all duration-200"
            :class="$route.path.startsWith('/admin')
              ? 'bg-mogao-cinnabar text-white shadow-mogao'
              : 'text-mogao-ochre-600 hover:bg-mogao-ochre-50 hover:text-mogao-cinnabar'"
          >
            管理
          </router-link>

          <!-- 分隔线 -->
          <span class="mx-2 w-px h-5 bg-mogao-ochre-300" />

          <!-- 用户区域 -->
          <div v-if="user" class="flex items-center gap-2 pl-1">
            <span class="text-sm text-mogao-lapis-600 font-medium">
              {{ user.username }}
            </span>
            <button
              @click="handleLogout"
              class="px-3 py-1.5 text-xs border border-mogao-ochre-300 text-mogao-ochre-600 rounded-mogao hover:bg-mogao-ochre-50 transition-colors"
            >
              退出
            </button>
          </div>
          <router-link v-else to="/login" class="px-3 py-1.5 text-xs border border-mogao-ochre-300 text-mogao-ochre-600 rounded-mogao hover:bg-mogao-ochre-50 transition-colors">
            登录
          </router-link>
        </nav>

        <!-- 移动端汉堡按钮 -->
        <button
          class="md:hidden p-2 rounded-mogao text-mogao-lapis-700 hover:bg-mogao-ochre-50"
          @click="mobileOpen = !mobileOpen"
          aria-label="切换菜单"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path v-if="mobileOpen" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            <path v-else strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 移动端菜单 -->
    <Transition name="slide">
      <nav
        v-if="mobileOpen"
        class="md:hidden overflow-hidden border-t border-mogao-ochre-200 bg-white"
      >
        <div class="px-4 py-3 space-y-1">
          <router-link
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            :exact="item.to === '/'"
            @click="mobileOpen = false"
            class="block px-4 py-2.5 rounded-mogao text-sm font-medium transition"
            :class="$route.path === item.to
              ? 'bg-mogao-cinnabar text-white'
              : 'text-mogao-lapis-700 hover:bg-mogao-ochre-50'"
          >
            {{ item.label }}
          </router-link>

          <!-- 移动端用户区域 -->
          <div class="border-t border-mogao-ochre-200 pt-2 mt-2">
            <div class="flex items-center justify-between px-4 py-2">
              <span v-if="user" class="text-sm text-mogao-lapis-600 font-medium">
                {{ user.username }}
              </span>
              <button
                v-if="user"
                @click="handleLogout()"
                class="px-3 py-1.5 text-xs border border-mogao-ochre-300 text-mogao-ochre-600 rounded-mogao hover:bg-mogao-ochre-50 transition-colors"
              >
                退出
              </button>
              <router-link
                v-else
                to="/login"
                @click="mobileOpen = false"
                class="px-3 py-1.5 text-xs border border-mogao-ochre-300 text-mogao-ochre-600 rounded-mogao hover:bg-mogao-ochre-50 transition-colors"
              >
                登录
              </router-link>
            </div>
          </div>
        </div>
      </nav>
    </Transition>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, type Ref } from 'vue'
import { useRouter, type Router } from 'vue-router'
import { useAuth } from '@/composables'
import type { User } from '@/types/user'

interface NavItem {
  to: string
  label: string
}

const navItems: NavItem[] = [
  { to: '/',          label: '每日一服' },
  { to: '/explore',   label: '服饰浏览' },
  { to: '/statistics',label: '地理分布' },
  { to: '/favorites', label: '我的收藏' },
  { to: '/submit',          label: '上传投稿' },
  { to: '/my-submissions', label: '我的投稿' },
]

// Check admin status from stored user
const isAdmin = computed(() => {
  if (!user.value || !('role' in user.value)) return false
  return (user.value as User).role === 'admin'
})

const mobileOpen: Ref<boolean> = ref(false)
const { user, logout } = useAuth()
const router: Router = useRouter()

function handleLogout(): void {
  logout()
  mobileOpen.value = false
  router.push('/login')
}
</script>

<style scoped>
.slide-enter-active {
  transition: all 0.25s ease;
}
.slide-leave-active {
  transition: all 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
}
.slide-enter-to,
.slide-leave-from {
  max-height: 400px;
  opacity: 1;
}
</style>

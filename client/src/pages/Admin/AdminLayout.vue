<template>
  <div class="flex min-h-screen bg-mogao-leadwhite-500">
    <!-- Sidebar -->
    <aside class="w-64 bg-mogao-lapis-900 text-white flex-shrink-0">
      <div class="p-6 border-b border-white/10">
        <router-link to="/" class="text-xl font-bold tracking-wide">
          <span class="text-mogao-cinnabar-400">非遗</span>华裳
        </router-link>
        <p class="text-xs text-white/50 mt-1">后台管理面板</p>
      </div>

      <nav class="p-4 space-y-1">
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-4 py-3 rounded-lg text-sm transition-colors"
          :class="
            $route.path === item.to || (item.exact !== false && $route.path.startsWith(item.to + '/'))
              ? 'bg-mogao-cinnabar-500 text-white'
              : 'text-white/70 hover:bg-white/10 hover:text-white'
          "
        >
          <span class="text-lg">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="absolute bottom-0 w-64 p-4 border-t border-white/10">
        <router-link
          to="/"
          class="flex items-center gap-2 text-sm text-white/50 hover:text-white transition-colors"
        >
          <span>←</span>
          <span>返回前台</span>
        </router-link>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 overflow-auto">
      <div class="p-8">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
interface NavItem {
  to: string
  icon: string
  label: string
  exact?: boolean
}

const navItems: NavItem[] = [
  { to: '/admin', icon: '📊', label: '仪表盘', exact: true },
  { to: '/admin/submissions', icon: '📨', label: '投稿审核' },
  { to: '/admin/models', icon: '📦', label: '模型管理' },
  { to: '/admin/garments', icon: '👘', label: '服饰管理' },
]
</script>

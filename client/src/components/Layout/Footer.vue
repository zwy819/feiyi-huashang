<template>
  <footer class="bg-mogao-lapis-900 text-mogao-leadwhite-300 mt-auto">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
        <!-- 品牌介绍 -->
        <div>
          <h3 class="font-serif text-lg font-bold text-white mb-3">非遗华裳</h3>
          <p class="text-sm leading-relaxed text-mogao-leadwhite-400">
            中国非物质文化遗产服饰3D数字博物馆。
            以数字技术守护千年织造智慧，让每一针每一线都焕发新生。
          </p>
        </div>

        <!-- 快速链接 -->
        <div>
          <h4 class="font-serif text-sm font-semibold text-white mb-3">快速链接</h4>
          <ul class="space-y-2">
            <li v-for="link in footerLinks" :key="link.to">
              <router-link
                :to="link.to"
                class="text-sm text-mogao-leadwhite-400 hover:text-mogao-gold transition-colors"
              >
                {{ link.label }}
              </router-link>
            </li>
          </ul>
        </div>

        <!-- 技术说明 -->
        <div>
          <h4 class="font-serif text-sm font-semibold text-white mb-3">技术架构</h4>
          <p class="text-xs text-mogao-leadwhite-500 space-y-1">
            <span class="block">前端：Vue 3 + Three.js</span>
            <span class="block">后端：Go + Gin + SQLite</span>
            <span class="block">配色：敦煌莫高窟色系</span>
          </p>
        </div>
      </div>

      <!-- 底部装饰线与版权 -->
      <div class="mt-8 pt-6 border-t border-mogao-lapis-700">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-2">
          <p class="text-xs text-mogao-leadwhite-500">
            &copy; {{ currentYear }} 非遗华裳 · 数字传承非物质文化遗产
          </p>
          <p class="text-xs text-mogao-leadwhite-600">
            敦煌莫高窟 · 匠心织造
          </p>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
interface FooterLink {
  to: string
  label: string
}

const footerLinks: FooterLink[] = [
  { to: '/explore', label: '服饰浏览' },
  { to: '/statistics', label: '地理分布' },
  { to: '/favorites', label: '我的收藏' },
  { to: '/report', label: '导出报告' },
]

const currentYear: number = new Date().getFullYear()
</script>

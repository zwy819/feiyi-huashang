import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/pages/Admin/AdminLoginPage.vue'),
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/Auth/AuthPage.vue'),
  },
  {
    path: '/',
    component: () => import('@/components/Layout/Layout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/pages/Home/HomePage.vue'),
      },
      {
        path: 'explore',
        name: 'Explore',
        component: () => import('@/pages/Explore/ExplorePage.vue'),
      },
      {
        path: 'detail/:id',
        name: 'Detail',
        component: () => import('@/pages/Detail/DetailPage.vue'),
      },
      {
        path: 'favorites',
        name: 'Favorites',
        component: () => import('@/pages/Favorites/FavoritesPage.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'statistics',
        name: 'Statistics',
        component: () => import('@/pages/Statistics/StatisticsPage.vue'),
      },
      {
        path: 'report',
        name: 'Report',
        component: () => import('@/pages/Report/ReportPage.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'submit',
        name: 'Submit',
        component: () => import('@/pages/Submit/SubmitPage.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'my-submissions',
        name: 'MySubmissions',
        component: () => import('@/pages/Submit/MySubmissionsPage.vue'),
        meta: { requiresAuth: true },
      },
      // Admin routes
      {
        path: 'admin',
        component: () => import('@/pages/Admin/AdminLayout.vue'),
        meta: { requiresAuth: true, requiresAdmin: true },
        children: [
          {
            path: '',
            name: 'AdminDashboard',
            component: () => import('@/pages/Admin/DashboardPage.vue'),
          },
          {
            path: 'models',
            name: 'AdminModels',
            component: () => import('@/pages/Admin/ModelsPage.vue'),
          },
          {
            path: 'garments',
            name: 'AdminGarments',
            component: () => import('@/pages/Admin/GarmentsPage.vue'),
          },
          {
            path: 'garments/new',
            name: 'AdminGarmentCreate',
            component: () => import('@/pages/Admin/GarmentEditPage.vue'),
          },
          {
            path: 'garments/:id/edit',
            name: 'AdminGarmentEdit',
            component: () => import('@/pages/Admin/GarmentEditPage.vue'),
          },
          {
            path: 'submissions',
            name: 'AdminSubmissions',
            component: () => import('@/pages/Admin/SubmissionsPage.vue'),
          },
        ],
      },
    ],
  },
  {
    // 404 重定向
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

// 登录守卫
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')

  // 管理员登录页：已登录管理员直接进后台
  if (to.path === '/admin/login') {
    const userStr = localStorage.getItem('user')
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr) as { role?: string }
        if (user.role === 'admin') {
          next('/admin')
          return
        }
      } catch {
        // 解析失败，允许继续显示登录页
      }
    }
    next()
    return
  }

  if (to.path === '/login') {
    if (token) {
      next('/')
    } else {
      next()
    }
    return
  }

  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }

  if (to.meta.requiresAdmin) {
    const userStr = localStorage.getItem('user')
    if (!userStr) {
      next('/admin/login')
      return
    }
    try {
      const user = JSON.parse(userStr) as { role?: string }
      if (user.role !== 'admin') {
        next('/')
        return
      }
    } catch {
      next('/admin/login')
      return
    }
  }

  next()
})

export default router

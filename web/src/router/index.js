import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store/user'

const routes = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layouts/BasicLayout.vue'),
    redirect: '/home',
    children: [
      {
        path: '/home',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
        meta: { title: '首页', keepAlive: true }
      },
      {
        path: '/devices',
        name: 'Devices',
        component: () => import('@/views/devices/index.vue'),
        meta: { title: '监控管理', keepAlive: true }
      },
      {
        path: '/platform',
        name: 'Platform',
        component: () => import('@/views/platform/index.vue'),
        meta: { title: '平台管理', keepAlive: true }
      },
      {
        path: '/streams',
        name: 'Streams',
        component: () => import('@/views/streams/index.vue'),
        meta: { title: '流管理', keepAlive: true }
      },
      {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/settings/index.vue'),
        meta: { title: '系统设置', keepAlive: true }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: '404' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  // 恢复 localStorage 中的用户状态
  if (!userStore.token) {
    userStore.restoreFromStorage()
  }

  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - YSIP` : 'YSIP'

  // 检查是否需要登录
  if (to.path !== '/login') {
    if (!userStore.isLoggedIn) {
      // 未登录，跳转到登录页
      next('/login')
      return
    }
  } else {
    // 已登录访问登录页，跳转到首页
    if (userStore.isLoggedIn) {
      next('/')
      return
    }
  }

  next()
})

export default router

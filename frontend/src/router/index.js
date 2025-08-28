import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import Layout from '@/views/Layout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { 
      requiresAuth: false,
      titleKey: 'common.login'
    }
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { 
          titleKey: 'nav.dashboard',
          permission: 'admin'
        }
      },
      {
        path: 'applications',
        name: 'Applications',
        component: () => import('@/views/Applications.vue'),
        meta: { 
          titleKey: 'nav.applications',
          permission: 'admin'
        }
      },
      {
        path: 'versions',
        name: 'VersionManagement',
        component: () => import('@/views/VersionManagement.vue'),
        meta: { 
          titleKey: 'nav.version',
          permission: 'admin'
        }
      },
      {
        path: 'channels',
        name: 'ChannelManagement',
        component: () => import('@/views/ChannelManagement.vue'),
        meta: { 
          titleKey: 'nav.channel',
          permission: 'admin'
        }
      },
      {
        path: 'statistics',
        name: 'Statistics',
        component: () => import('@/views/Statistics.vue'),
        meta: { 
          titleKey: 'nav.statistics',
          permission: 'admin'
        }
      },
      {
        path: 'docs',
        name: 'APIDocs',
        component: () => import('@/views/APIDocs.vue'),
        meta: { 
          titleKey: 'docs.title',
          permission: 'admin'
        }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory('/admin-ui/'),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // 初始化认证状态
  authStore.initAuth()
  
  // 如果访问登录页面且已登录，跳转到首页
  if (to.name === 'Login' && authStore.isLoggedIn) {
    next('/')
    return
  }
  
  // 检查是否需要认证
  if (to.meta.requiresAuth || to.matched.some(record => record.meta.requiresAuth)) {
    if (!authStore.isLoggedIn) {
      // 未登录，跳转到登录页
      next('/login')
      return
    }
    
    // 检查权限
    if (to.meta.permission && !authStore.hasPermission(to.meta.permission)) {
      // 权限不足，跳转到首页或显示错误页面
      console.warn('Access denied: insufficient permissions')
      next('/')
      return
    }
    
    // 尝试获取最新的用户信息
    try {
      await authStore.fetchProfile()
    } catch (error) {
      console.error('Failed to fetch profile:', error)
      // 如果获取用户信息失败，可能是token过期，跳转到登录页
      if (error.message === 'Authentication expired') {
        next('/login')
        return
      }
    }
  }
  
  next()
})

export default router

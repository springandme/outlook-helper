import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// 静态导入组件
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import EmailsView from '@/views/EmailsView.vue'
import TagsView from '@/views/TagsView.vue'
import Layout from '@/components/Layout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: Layout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: DashboardView
        },
        {
          path: 'emails',
          name: 'emails',
          component: EmailsView
        },
        {
          path: 'tags',
          name: 'tags',
          component: TagsView
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // 检查认证状态
  authStore.checkAuth()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，跳转到登录页
    next('/login')
  } else if (to.name === 'login' && authStore.isAuthenticated) {
    // 已登录用户访问登录页，跳转到首页
    next('/')
  } else {
    next()
  }
})

export default router

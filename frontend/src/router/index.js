import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('../views/Login.vue'), meta: { public: true } },
    { path: '/callback', component: () => import('../views/Callback.vue'), meta: { public: true } },
    {
      path: '/',
      component: () => import('../layouts/AppLayout.vue'),
      children: [
        { path: '', component: () => import('../views/Today.vue') },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const user = useUserStore()
  if (to.meta.public) {
    if (user.isLogin && to.path === '/login') return '/'
    return true
  }
  if (!user.isLogin) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (!user.profile) {
    try {
      await user.fetchMe()
    } catch {
      user.logout()
      return { path: '/login' }
    }
  }
  return true
})

export default router

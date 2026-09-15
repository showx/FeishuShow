<template>
  <div class="page">
    <div class="card box">
      <h2>{{ title }}</h2>
      <p class="muted">{{ detail }}</p>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authApi } from '../api'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const user = useUserStore()
const title = ref('正在完成飞书授权…')
const detail = ref('请稍候，不要关闭此页。')

onMounted(async () => {
  const err = route.query.error
  if (err) {
    title.value = '授权已取消'
    detail.value = '你可以返回重新登录，或先看演示。'
    setTimeout(() => router.replace('/login'), 1600)
    return
  }
  const code = route.query.code
  const state = route.query.state
  if (!code) {
    title.value = '缺少授权码'
    detail.value = '请从登录页重新发起飞书授权。'
    return
  }
  try {
    const res = await authApi.exchange(code, state)
    user.setSession(res.data.token, res.data.user)
    router.replace('/')
  } catch (e) {
    title.value = '授权失败'
    detail.value = e.message || '兑换登录态失败'
  }
})
</script>

<style scoped>
.page {
  min-height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
}
.box { width: 420px; padding: 28px 24px; text-align: center; }
h2 { margin-bottom: 8px; }
</style>

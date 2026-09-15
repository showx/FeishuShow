<template>
  <div class="page">
    <div class="hero">
      <div class="mark">飞</div>
      <h1>FeishuShow</h1>
      <p class="lead">把飞书里的今天摊开：日程、文档、待办，再一键出纪要。</p>
    </div>
    <div class="card box">
      <button class="btn btn-block" :disabled="!status.feishuConfigured || loading" @click="goFeishu">
        {{ status.feishuConfigured ? '用飞书账号登录' : '尚未配置飞书应用' }}
      </button>
      <button class="btn btn-ghost btn-block" :disabled="loading" @click="goDemo">先看演示数据</button>
      <p class="hint" v-if="!status.feishuConfigured">
        在仓库根目录复制 <code>.env.example</code> 为 <code>.env</code>，填入自建应用的 App ID / Secret，
        重定向 URL 设为 <code>http://localhost:5173/callback</code>。
      </p>
      <p class="hint" v-else>登录后只读取你本人权限内的日程、文档和任务，不会越权。</p>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authApi } from '../api'
import { useUserStore } from '../stores/user'
import { toast } from '../utils'

const router = useRouter()
const route = useRoute()
const user = useUserStore()
const loading = ref(false)
const status = reactive({ feishuConfigured: false, demoAvailable: true })

onMounted(async () => {
  try {
    const res = await authApi.status()
    Object.assign(status, res.data)
  } catch (e) {
    toast(e.message || '无法连接后端，请先启动 API')
  }
})

async function goFeishu() {
  loading.value = true
  try {
    const res = await authApi.feishuUrl()
    window.location.href = res.data.url
  } catch (e) {
    toast(e.message || '无法打开飞书授权')
    loading.value = false
  }
}

async function goDemo() {
  loading.value = true
  try {
    const res = await authApi.demo()
    user.setSession(res.data.token, res.data.user)
    router.push(route.query.redirect || '/')
  } catch (e) {
    toast(e.message || '演示登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  background:
    radial-gradient(900px 420px at 50% -10%, #d7e4ff 0%, transparent 60%),
    var(--bg);
}
.hero { text-align: center; margin-bottom: 28px; }
.mark {
  width: 56px;
  height: 56px;
  margin: 0 auto 12px;
  border-radius: 14px;
  background: var(--brand);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 700;
  box-shadow: 0 10px 24px rgba(51, 112, 255, 0.28);
}
h1 { font-size: 28px; letter-spacing: 0.02em; }
.lead { margin-top: 10px; color: var(--muted); max-width: 420px; }
.box { width: 400px; max-width: 100%; padding: 24px; }
.btn-block + .btn-block { margin-top: 10px; }
.hint { margin-top: 16px; color: var(--muted); font-size: 12px; line-height: 1.7; }
code { background: #f2f3f5; padding: 1px 5px; border-radius: 4px; }
</style>

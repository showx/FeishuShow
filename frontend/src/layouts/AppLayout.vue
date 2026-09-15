<template>
  <div class="shell">
    <header class="top">
      <div class="brand" @click="$router.push('/')">
        <span class="mark">飞</span>
        <div>
          <strong>FeishuShow</strong>
          <span class="sub">今日雷达</span>
        </div>
      </div>
      <div class="spacer"></div>
      <span class="mode" :class="user.profile?.demo ? 'demo' : 'live'">
        {{ user.profile?.demo ? '演示数据' : '已连接飞书' }}
      </span>
      <div class="who">
        <span class="avatar">
          <img v-if="user.profile?.avatar" :src="user.profile.avatar" alt="" />
          <template v-else>{{ initial(user.profile?.name) }}</template>
        </span>
        <span>{{ user.profile?.name }}</span>
        <button class="btn btn-ghost btn-sm" @click="logout">退出</button>
      </div>
    </header>
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { initial } from '../utils'

const router = useRouter()
const user = useUserStore()

function logout() {
  user.logout()
  router.push('/login')
}
</script>

<style scoped>
.shell { min-height: 100%; padding-top: var(--header-h); }
.top {
  position: fixed;
  inset: 0 0 auto 0;
  height: var(--header-h);
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line-soft);
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.mark {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--brand);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}
.brand strong { display: block; font-size: 14px; line-height: 1.2; }
.sub { font-size: 12px; color: var(--muted); }
.spacer { flex: 1; }
.mode {
  font-size: 12px;
  padding: 3px 8px;
  border-radius: 99px;
}
.mode.demo { background: var(--warning-soft); color: var(--warning); }
.mode.live { background: var(--success-soft); color: var(--success); }
.who { display: flex; align-items: center; gap: 8px; }
.main { padding: 20px 24px 40px; max-width: 1280px; margin: 0 auto; }
</style>

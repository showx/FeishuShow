<template>
  <div>
    <div class="head">
      <div>
        <h2>今天 · {{ board.date || '—' }}</h2>
        <p class="muted">日程、发呆文档和待办摊在一屏。点一场会，右侧出纪要。</p>
      </div>
      <button class="btn btn-ghost" :disabled="loading" @click="load">刷新</button>
    </div>

    <div class="warns" v-if="board.warnings?.length">
      <p v-for="(w, i) in board.warnings" :key="i">{{ w }}</p>
    </div>

    <div class="stats">
      <div class="card stat">
        <div class="k">今日日程</div>
        <div class="v">{{ board.stats.meetings || 0 }}</div>
        <div class="s" v-if="board.stats.live">{{ board.stats.live }} 场进行中</div>
      </div>
      <div class="card stat">
        <div class="k">待办</div>
        <div class="v">{{ board.stats.tasks || 0 }}</div>
        <div class="s danger" v-if="board.stats.overdue">{{ board.stats.overdue }} 项过期</div>
      </div>
      <div class="card stat">
        <div class="k">发呆文档</div>
        <div class="v warn">{{ board.stats.staleDocs || 0 }}</div>
        <div class="s">超过 30 天未改</div>
      </div>
      <div class="card stat">
        <div class="k">最近会话</div>
        <div class="v">{{ board.stats.chats || 0 }}</div>
      </div>
    </div>

    <div class="grid">
      <section class="card pane">
        <h3>今日日程</h3>
        <div v-if="!board.events?.length" class="empty">今天没有日程</div>
        <ul class="list">
          <li
            v-for="ev in board.events"
            :key="ev.id"
            :class="{ active: selected?.id === ev.id }"
            @click="selected = ev"
          >
            <div class="row">
              <strong>{{ ev.title }}</strong>
              <span :class="['tag', 'tag-' + ev.status]">{{ statusLabel(ev.status) }}</span>
            </div>
            <div class="meta">
              <span>{{ fmtTime(ev.start) }} – {{ fmtTime(ev.end) }}</span>
              <span v-if="ev.location"> · {{ ev.location }}</span>
              <span v-if="ev.hasMeeting"> · 飞书会议</span>
            </div>
            <div class="people" v-if="ev.attendees?.length">{{ ev.attendees.join('、') }}</div>
          </li>
        </ul>
      </section>

      <section class="card pane">
        <h3>文档雷达</h3>
        <div v-if="!board.docs?.length" class="empty">还没有最近文档</div>
        <ul class="list">
          <li v-for="doc in board.docs" :key="doc.token">
            <div class="row">
              <a v-if="doc.url" :href="doc.url" target="_blank" rel="noreferrer">{{ doc.name }}</a>
              <strong v-else>{{ doc.name }}</strong>
              <span :class="['tag', doc.stale ? 'tag-stale' : 'tag-' + (doc.type || 'docx')]">
                {{ doc.stale ? `发呆 ${doc.daysIdle} 天` : typeLabel(doc.type) }}
              </span>
            </div>
            <div class="meta">上次编辑 {{ fmtDateTime(doc.modifiedAt) }}</div>
          </li>
        </ul>
        <h3 class="sub">待办</h3>
        <div v-if="!board.tasks?.length" class="empty slim">没有未完成任务</div>
        <ul class="list compact">
          <li v-for="t in board.tasks" :key="t.id">
            <div class="row">
              <span>{{ t.title }}</span>
              <span v-if="t.overdue" class="tag tag-overdue">过期</span>
              <span v-else-if="t.due" class="muted">{{ fmtDateTime(t.due) }}</span>
            </div>
          </li>
        </ul>
        <h3 class="sub" v-if="board.chats?.length">最近会话</h3>
        <ul class="list compact" v-if="board.chats?.length">
          <li v-for="ch in board.chats" :key="ch.id">
            <div class="row">
              <span>{{ ch.name }}</span>
              <span class="muted">{{ ch.memberCount }} 人</span>
            </div>
          </li>
        </ul>
      </section>

      <section class="card pane minutes">
        <h3>会后纪要</h3>
        <p class="muted pick" v-if="!selected">点左侧一场会议，生成决议 / 待办 / 风险。</p>
        <template v-else>
          <div class="pick-title">{{ selected.title }}</div>
          <p class="muted">{{ fmtTime(selected.start) }} – {{ fmtTime(selected.end) }} · {{ statusLabel(selected.status) }}</p>
          <div class="actions">
            <button class="btn" :disabled="busy" @click="makeMinutes(false)">生成纪要卡片</button>
            <button class="btn btn-ghost" :disabled="busy || isDemo" @click="makeMinutes(true)">
              写回飞书
            </button>
          </div>
          <p class="muted" v-if="isDemo">演示模式只在本地出卡片，连上飞书后才能写回文档和任务。</p>
        </template>

        <div v-if="result" class="result">
          <div class="row">
            <strong>{{ result.title }}</strong>
            <span class="tag tag-fresh">{{ result.source }}</span>
          </div>
          <a v-if="result.docUrl" :href="result.docUrl" target="_blank" rel="noreferrer">打开飞书文档</a>
          <div class="block" v-if="result.decisions?.length">
            <h4>决议</h4>
            <ul>
              <li v-for="(d, i) in result.decisions" :key="'d' + i">{{ d }}</li>
            </ul>
          </div>
          <div class="block" v-if="result.actions?.length">
            <h4>待办</h4>
            <ul>
              <li v-for="(a, i) in result.actions" :key="'a' + i">
                <b>{{ a.owner || '待指定' }}</b> {{ a.title }}
                <span class="muted"> {{ a.dueHint }}</span>
              </li>
            </ul>
          </div>
          <div class="block" v-if="result.risks?.length">
            <h4>风险</h4>
            <ul>
              <li v-for="(r, i) in result.risks" :key="'r' + i">{{ r }}</li>
            </ul>
          </div>
          <details>
            <summary>Markdown</summary>
            <pre>{{ result.markdown }}</pre>
          </details>
        </div>

        <h3 class="sub" v-if="history.length">最近生成</h3>
        <ul class="list compact" v-if="history.length">
          <li v-for="h in history" :key="h.id">
            <div class="row">
              <span>{{ h.title }}</span>
              <a v-if="h.docUrl" :href="h.docUrl" target="_blank" rel="noreferrer">文档</a>
            </div>
            <div class="meta">{{ fmtDateTime(h.createdAt) }}</div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { skillApi, todayApi } from '../api'
import { useUserStore } from '../stores/user'
import { fmtDateTime, fmtTime, statusLabel, toast, typeLabel } from '../utils'

const user = useUserStore()
const loading = ref(false)
const busy = ref(false)
const selected = ref(null)
const result = ref(null)
const history = ref([])
const board = ref({
  date: '',
  stats: {},
  events: [],
  tasks: [],
  docs: [],
  chats: [],
  warnings: [],
})

const isDemo = computed(() => Boolean(user.profile?.demo))

async function load() {
  loading.value = true
  try {
    const res = await todayApi.get()
    board.value = res.data
    if (!selected.value && res.data.events?.length) {
      selected.value = res.data.events.find((e) => e.status === 'ended') || res.data.events[0]
    }
  } catch (e) {
    toast(e.message || '加载今日数据失败')
  } finally {
    loading.value = false
  }
}

async function loadHistory() {
  try {
    const res = await skillApi.history()
    history.value = res.data || []
  } catch {
    history.value = []
  }
}

async function makeMinutes(writeBack) {
  if (!selected.value) return
  busy.value = true
  try {
    const res = await skillApi.minutes({
      eventId: selected.value.id,
      calendarId: selected.value.calendarId,
      writeBack,
    })
    result.value = res.data
    toast(writeBack && res.data.docUrl ? '已写回飞书' : '纪要已生成')
    await loadHistory()
  } catch (e) {
    toast(e.message || '生成失败')
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  await load()
  await loadHistory()
})
</script>

<style scoped>
.head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}
.head h2 { font-size: 20px; }
.warns {
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--warning-soft);
  color: var(--warning);
  font-size: 12px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.stat { padding: 16px 18px; }
.stat .k { color: var(--muted); font-size: 12px; }
.stat .v { margin-top: 4px; font-size: 28px; font-weight: 700; color: var(--brand); }
.stat .v.warn { color: var(--warning); }
.stat .s { margin-top: 4px; color: var(--faint); font-size: 12px; }
.stat .s.danger { color: var(--danger); }
.grid {
  display: grid;
  grid-template-columns: 1.1fr 1fr 1.05fr;
  gap: 14px;
  align-items: start;
}
.pane { padding: 16px 16px 12px; min-height: 420px; }
.pane h3 { font-size: 14px; margin-bottom: 12px; }
.pane h3.sub { margin-top: 18px; }
.list { list-style: none; }
.list li {
  padding: 10px 8px;
  border-radius: 8px;
  border-bottom: 1px solid var(--line-soft);
}
.list li:last-child { border-bottom: none; }
.pane:first-child li { cursor: pointer; }
.pane:first-child li:hover,
.pane:first-child li.active { background: var(--brand-light); }
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.row strong, .row a, .row span:first-child {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.meta, .people { margin-top: 4px; color: var(--muted); font-size: 12px; }
.people { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.compact li { padding: 8px; }
.empty.slim { padding: 16px 8px; }
.minutes .pick { margin-bottom: 8px; }
.pick-title { font-weight: 600; margin-bottom: 4px; }
.actions { display: flex; gap: 8px; margin: 12px 0; flex-wrap: wrap; }
.result { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--line-soft); }
.block { margin-top: 12px; }
.block h4 { font-size: 13px; margin-bottom: 6px; }
.block ul { padding-left: 18px; color: var(--ink); }
.block li { margin: 4px 0; }
details { margin-top: 12px; color: var(--muted); }
pre {
  margin-top: 8px;
  padding: 10px;
  background: #f7f8fa;
  border-radius: 8px;
  white-space: pre-wrap;
  font-size: 12px;
  max-height: 220px;
  overflow: auto;
}
@media (max-width: 1100px) {
  .stats, .grid { grid-template-columns: 1fr; }
}
</style>

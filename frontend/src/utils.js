export function toast(message) {
  const el = document.createElement('div')
  el.className = 'toast'
  el.textContent = message
  document.body.appendChild(el)
  setTimeout(() => el.remove(), 2400)
}

export function initial(name) {
  const s = (name || '?').trim()
  return s.slice(0, 1)
}

export function fmtTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
}

export function fmtDateTime(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false })
}

export function statusLabel(s) {
  return { upcoming: '未开始', live: '进行中', ended: '已结束' }[s] || s
}

export function typeLabel(t) {
  return { docx: '文档', doc: '文档', sheet: '表格', bitable: '多维表', slides: '幻灯片', mindnote: '思维笔记', file: '文件' }[t] || t
}

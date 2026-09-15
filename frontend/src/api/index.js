import http from './http'

export const authApi = {
  status: () => http.get('/auth/status'),
  feishuUrl: () => http.get('/auth/feishu/url'),
  exchange: (code, state) => http.post('/auth/feishu/exchange', { code, state }),
  demo: () => http.post('/auth/demo'),
  me: () => http.get('/me'),
}

export const todayApi = {
  get: () => http.get('/today'),
}

export const skillApi = {
  minutes: (payload) => http.post('/skills/minutes', payload),
  history: () => http.get('/skills/minutes'),
}

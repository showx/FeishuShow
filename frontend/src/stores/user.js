import { defineStore } from 'pinia'
import { authApi } from '../api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    profile: null,
  }),
  getters: {
    isLogin: (s) => Boolean(s.token),
  },
  actions: {
    setSession(token, user) {
      this.token = token
      this.profile = user
      localStorage.setItem('token', token)
    },
    logout() {
      this.token = ''
      this.profile = null
      localStorage.removeItem('token')
    },
    async fetchMe() {
      if (!this.token) return
      const res = await authApi.me()
      this.profile = res.data
    },
  },
})

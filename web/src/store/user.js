import { defineStore } from 'pinia'
import { login as loginApi, logout as logoutApi } from '@/api/user'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    userInfo: null
  }),

  getters: {
    isLoggedIn: state => !!state.token,
    userName: state => state.userInfo?.name ?? ''
  },

  actions: {
    setToken(token) {
      this.token = token
      localStorage.setItem('token', token)
    },

    setUserInfo(userInfo) {
      this.userInfo = userInfo
      localStorage.setItem('userInfo', JSON.stringify(userInfo))
    },

    /**
     * 用户登录
     * @param {Object} loginForm - 登录表单数据
     * @returns {Promise}
     */
    async login(loginForm) {
      const res = await loginApi(loginForm)
      // res.data 包含 { access_token, username, name, role, user_id, avatar }
      this.setToken(res.data.access_token)
      this.setUserInfo({
        username: res.data.username,
        name: res.data.name,
        role: res.data.role,
        user_id: res.data.user_id,
        avatar: res.data.avatar
      })
      return res
    },

    /**
     * 用户登出
     * @returns {Promise}
     */
    async logout() {
      try {
        await logoutApi()
      } catch (error) {
        // 登出失败也继续执行清理逻辑
      }
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
    },

    /**
     * 从 localStorage 恢复用户状态
     */
    restoreFromStorage() {
      const token = localStorage.getItem('token')
      const userInfo = localStorage.getItem('userInfo')
      if (token) {
        this.token = token
      }
      if (userInfo) {
        this.userInfo = JSON.parse(userInfo)
      }
    }
  }
})

import { defineStore } from 'pinia'
import request, { authApi } from '@/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    refreshToken: localStorage.getItem('refresh_token') || null,
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    isAuthenticated: false
  }),

  getters: {
    isLoggedIn: (state) => !!state.token && !!state.user,
    currentUser: (state) => state.user,
    userRole: (state) => state.user?.role,
    isAdmin: (state) => ['admin', 'superadmin'].includes(state.user?.role),
    isSuperAdmin: (state) => state.user?.role === 'superadmin'
  },

  actions: {
    // 登录
    async login(username, password) {
      try {
        const response = await authApi.post('/login', {
          username,
          password
        })

        if (response.code === 200) {
          const { token, refresh_token, user } = response.data
          
          // 保存认证信息
          this.token = token
          this.refreshToken = refresh_token
          this.user = user
          this.isAuthenticated = true

          // 持久化到本地存储
          localStorage.setItem('token', token)
          localStorage.setItem('refresh_token', refresh_token)
          localStorage.setItem('user', JSON.stringify(user))

          

          return {
            success: true,
            user
          }
        } else {
          throw new Error(response.message || '登录失败')
        }
      } catch (error) {
        console.error('Login error:', error)
        return {
          success: false,
          message: error.response?.data?.message || error.message || '登录失败'
        }
      }
    },

    // 登出
    async logout() {
      try {
        // 调用登出API
        if (this.refreshToken) {
          await authApi.post('/logout', {
            refresh_token: this.refreshToken
          })
        }
      } catch (error) {
        console.error('Logout API error:', error)
        // 即使API调用失败，也要清除本地状态
      } finally {
        // 清除本地状态
        this.clearAuth()
      }
    },

    // 清除认证信息
    clearAuth() {
      this.token = null
      this.refreshToken = null
      this.user = null
      this.isAuthenticated = false

      // 清除本地存储
      localStorage.removeItem('token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('user')

      
    },

    // 刷新令牌
    async refreshAccessToken() {
      if (!this.refreshToken) {
        this.clearAuth()
        return false
      }

      try {
        const response = await authApi.post('/refresh', {
          refresh_token: this.refreshToken
        })

        if (response.code === 200) {
          const { token, refresh_token, user } = response.data
          
          this.token = token
          this.refreshToken = refresh_token
          this.user = user

          localStorage.setItem('token', token)
          localStorage.setItem('refresh_token', refresh_token)
          localStorage.setItem('user', JSON.stringify(user))

          api.defaults.headers.common['Authorization'] = `Bearer ${token}`

          return true
        } else {
          throw new Error('Token refresh failed')
        }
      } catch (error) {
        console.error('Token refresh error:', error)
        this.clearAuth()
        return false
      }
    },

    // 获取用户配置文件
    async fetchProfile() {
      try {
        const response = await request.get('/profile')
        
        if (response.code === 200) {
          this.user = response.data
          localStorage.setItem('user', JSON.stringify(this.user))
          return this.user
        }
      } catch (error) {
        console.error('Fetch profile error:', error)
        if (error.response?.status === 401) {
          // 尝试刷新令牌
          const refreshed = await this.refreshAccessToken()
          if (!refreshed) {
            this.clearAuth()
            throw new Error('Authentication expired')
          }
          // 递归重试
          return this.fetchProfile()
        }
        throw error
      }
    },

    // 修改密码
    async changePassword(currentPassword, newPassword) {
      try {
        const requestData = {
          new_password: newPassword
        }
        
        // 只有当前密码不为空时才添加
        if (currentPassword && currentPassword.trim()) {
          requestData.current_password = currentPassword
        } else {
          requestData.current_password = ""
        }

        const response = await request.post('/change-password', requestData)

        if (response.code === 200) {
          return { success: true }
        } else {
          throw new Error(response.message || '密码修改失败')
        }
      } catch (error) {
        return {
          success: false,
          message: error.response?.data?.message || error.message || '密码修改失败'
        }
      }
    },

    // 初始化认证状态
    initAuth() {
      if (this.token && this.user) {
        this.isAuthenticated = true
        
      }
    },

    // 检查权限
    hasPermission(permission) {
      if (!this.user) return false
      
      switch (permission) {
        case 'admin':
          return ['admin', 'superadmin'].includes(this.user.role)
        case 'superadmin':
          return this.user.role === 'superadmin'
        default:
          return true
      }
    }
  }
})





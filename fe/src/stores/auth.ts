import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userApi, type User, type LoginRequest } from '@/api/user'
import { isAdmin } from '@/utils/permissions'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref<string>(localStorage.getItem('token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refreshToken') || '')
  const user = ref<User | null>(null)
  const loading = ref(false)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)
  const isAdminUser = computed(() => {
    if (!user.value) return false
    return isAdmin(user.value.permissions)
  })

  // 登录
  const login = async (loginData: LoginRequest) => {
    loading.value = true
    try {
      const response = await userApi.login(loginData)
      
      token.value = response.accessToken
      refreshToken.value = response.refreshToken
      user.value = response.user
      
      // 保存到localStorage
      localStorage.setItem('token', response.accessToken)
      localStorage.setItem('refreshToken', response.refreshToken)
      
      return response
    } catch (error) {
      throw error
    } finally {
      loading.value = false
    }
  }

  // 刷新token
  const refresh = async () => {
    if (!refreshToken.value) {
      throw new Error('No refresh token available')
    }
    
    try {
      const response = await userApi.refreshToken({ refreshToken: refreshToken.value })
      
      token.value = response.accessToken
      refreshToken.value = response.refreshToken
      user.value = response.user
      
      localStorage.setItem('token', response.accessToken)
      localStorage.setItem('refreshToken', response.refreshToken)
      
      return response
    } catch (error) {
      logout()
      throw error
    }
  }

  // 获取用户信息
  const fetchUserInfo = async () => {
    try {
      const userInfo = await userApi.getUserInfo()
      user.value = userInfo
      return userInfo
    } catch (error) {
      logout()
      throw error
    }
  }

  // 登出
  const logout = () => {
    token.value = ''
    refreshToken.value = ''
    user.value = null
    
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  // 初始化 - 如果有token则获取用户信息
  const initialize = async () => {
    if (token.value) {
      try {
        await fetchUserInfo()
      } catch (error) {
        // 如果获取用户信息失败，清除token但不抛出错误
        logout()
        console.log('Token已过期或无效，已清除认证信息')
      }
    }
  }

  return {
    // 状态
    token,
    refreshToken,
    user,
    loading,
    
    // 计算属性
    isAuthenticated,
    isAdmin: isAdminUser,
    
    // 方法
    login,
    refresh,
    fetchUserInfo,
    logout,
    initialize
  }
})
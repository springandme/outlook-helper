import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { authAPI, type User, type LoginRequest } from '@/api'
import { ElMessage } from 'element-plus'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user_info') || 'null'))
  const isLoading = ref(false)

  const isAuthenticated = computed(() => !!token.value)

  // 登录
  const login = async (credentials: LoginRequest) => {
    isLoading.value = true
    try {
      const response = await authAPI.login(credentials)
      if (response.data.success && response.data.data) {
        const { token: authToken, user: userInfo } = response.data.data

        // 保存到localStorage
        localStorage.setItem('auth_token', authToken)
        localStorage.setItem('user_info', JSON.stringify(userInfo))

        // 更新store状态
        token.value = authToken
        user.value = userInfo

        ElMessage.success('登录成功')
        return true
      } else {
        ElMessage.error(response.data.message || '登录失败')
        return false
      }
    } catch (error: any) {
      const message = error.response?.data?.message || '登录失败，请检查网络连接'
      ElMessage.error(message)
      return false
    } finally {
      isLoading.value = false
    }
  }

  // 登出
  const logout = async () => {
    try {
      await authAPI.logout()
    } catch (error) {
      console.error('Logout API error:', error)
    } finally {
      // 清除本地存储
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')

      // 清除store状态
      token.value = null
      user.value = null

      ElMessage.success('已退出登录')
    }
  }

  // 检查token有效性
  const checkAuth = () => {
    const storedToken = localStorage.getItem('auth_token')
    const storedUser = localStorage.getItem('user_info')

    if (storedToken && storedUser) {
      token.value = storedToken
      user.value = JSON.parse(storedUser)
      return true
    }

    return false
  }

  return {
    token,
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    checkAuth
  }
})

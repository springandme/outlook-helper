<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="login-logo">
          <svg class="login-icon" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
            <path d="M896 42.666667H341.333333c-25.6 0-42.666667 17.066667-42.666666 42.666666v42.666667l341.333333 106.666667L938.666667 128V85.333333c0-25.6-17.066667-42.666667-42.666667-42.666666z" fill="#0364B8"></path>
            <path d="M1024 507.733333c0-8.533333-4.266667-17.066667-12.8-21.333333l-366.933333-209.066667s-4.266667 0-4.266667-4.266666c-8.533333-4.266667-12.8-4.266667-21.333333-4.266667s-17.066667 0-21.333334 4.266667c0 0-4.266667 0-4.266666 4.266666l-366.933334 209.066667c-8.533333 4.266667-12.8 12.8-12.8 21.333333s4.266667 17.066667 12.8 21.333334l366.933334 209.066666s4.266667 0 4.266666 4.266667c8.533333 4.266667 12.8 4.266667 21.333334 4.266667s17.066667 0 21.333333-4.266667c0 0 4.266667 0 4.266667-4.266667l366.933333-209.066666c8.533333-4.266667 12.8-12.8 12.8-21.333334z" fill="#0A2767"></path>
            <path d="M640 128H298.666667v298.666667l341.333333 341.333333h298.666667v-341.333333z" fill="#28A8EA"></path>
            <path d="M640 128h298.666667v298.666667h-298.666667z" fill="#50D9FF"></path>
            <path d="M298.666667 426.666667h341.333333v426.666666H298.666667z" fill="#0078D4"></path>
            <path d="M546.133333 810.666667H51.2C21.333333 810.666667 0 789.333333 0 759.466667V264.533333C0 234.666667 21.333333 213.333333 51.2 213.333333h499.2c25.6 0 46.933333 21.333333 46.933333 51.2v499.2c0 25.6-21.333333 46.933333-51.2 46.933334z" fill="#0078D4"></path>
            <path d="M157.866667 426.666667c12.8-25.6 29.866667-46.933333 55.466666-64 25.6-12.8 55.466667-21.333333 89.6-21.333334 29.866667 0 59.733333 8.533333 81.066667 21.333334 25.6 12.8 42.666667 34.133333 55.466667 59.733333 12.8 25.6 17.066667 55.466667 17.066666 85.333333 0 34.133333-8.533333 64-21.333333 89.6-8.533333 29.866667-29.866667 51.2-51.2 64-25.6 12.8-55.466667 21.333333-85.333333 21.333334-34.133333 0-59.733333-8.533333-85.333334-21.333334-25.6-12.8-42.666667-34.133333-55.466666-59.733333-12.8-25.6-21.333333-55.466667-21.333334-85.333333 0-34.133333 8.533333-64 21.333334-89.6z m59.733333 145.066666c8.533333 17.066667 17.066667 29.866667 29.866667 42.666667 12.8 8.533333 29.866667 12.8 51.2 12.8s38.4-4.266667 51.2-17.066667c12.8-8.533333 25.6-25.6 29.866666-42.666666 8.533333-12.8 12.8-34.133333 12.8-55.466667s-4.266667-38.4-8.533333-55.466667-17.066667-29.866667-29.866667-42.666666c-17.066667-12.8-34.133333-17.066667-55.466666-17.066667s-38.4 4.266667-51.2 17.066667c-12.8 8.533333-25.6 25.6-34.133334 42.666666-4.266667 12.8-8.533333 34.133333-8.533333 55.466667s4.266667 42.666667 12.8 59.733333z" fill="#FFFFFF"></path>
          </svg>
          <h1>{{ appTitle }}</h1>
        </div>
        <p>请输入授权令牌登录系统</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="auth_token">
          <el-input
            v-model="loginForm.auth_token"
            placeholder="请输入授权码"
            size="large"
            :prefix-icon="Key"
            show-password
            clearable
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-button"
            :loading="authStore.isLoading"
            @click="handleLogin"
          >
            {{ authStore.isLoading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>


    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Key } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { LoginRequest } from '@/api'

const router = useRouter()
const authStore = useAuthStore()

const appTitle = import.meta.env.VITE_APP_TITLE || 'Outlook取件助手'

const loginFormRef = ref<FormInstance>()
const loginForm = reactive<LoginRequest>({
  auth_token: ''
})

const loginRules: FormRules = {
  auth_token: [
    { required: true, message: '请输入授权码', trigger: 'blur' },
    { min: 8, max: 100, message: '授权码长度在 8 到 100 个字符', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return

  try {
    const valid = await loginFormRef.value.validate()
    if (!valid) return

    const success = await authStore.login(loginForm)
    if (success) {
      console.log('Login successful, navigating to dashboard...')
      try {
        await router.push('/')
        console.log('Navigation successful')
      } catch (navError) {
        console.error('Navigation error:', navError)
        // 如果路由跳转失败，尝试直接跳转到仪表盘
        window.location.href = '/'
      }
    }
  } catch (error) {
    console.error('Login validation error:', error)
  }
}

onMounted(() => {
  // 如果已经登录，直接跳转到首页
  if (authStore.isAuthenticated) {
    router.push('/')
  }


})
</script>

<style scoped>
.login-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: #f5f5f5;
  background-image:
    linear-gradient(90deg, rgba(200, 200, 200, 0.1) 1px, transparent 1px),
    linear-gradient(rgba(200, 200, 200, 0.1) 1px, transparent 1px);
  background-size: 20px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  overflow: hidden;
}

.login-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 40px;
  width: 100%;
  max-width: 400px;
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 8px;
}

.login-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.login-header h1 {
  color: #2c3e50;
  font-size: 28px;
  font-weight: 600;
  margin: 0;
}

.login-header p {
  color: #7f8c8d;
  font-size: 14px;
  margin: 0;
}

.login-form {
  margin-bottom: 20px;
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
}



:deep(.el-input__wrapper) {
  border-radius: 8px;
}

:deep(.el-button) {
  border-radius: 8px;
}
</style>

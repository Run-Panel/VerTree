<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>{{ $t('login.title') }}</h1>
        <p>{{ $t('login.subtitle') }}</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            :prefix-icon="User"
            :placeholder="$t('login.usernamePlaceholder')"
            size="large"
            :disabled="loading"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            :prefix-icon="Lock"
            type="password"
            :placeholder="$t('login.passwordPlaceholder')"
            size="large"
            :disabled="loading"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            @click="handleLogin"
          >
            {{ loading ? $t('login.loggingIn') : $t('login.loginButton') }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <p>{{ $t('login.defaultAccount') }}</p>
        <p class="warning">{{ $t('login.warningMessage') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElNotification } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()

const loginFormRef = ref()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [
    { required: true, message: t('login.usernameRequired'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('login.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('common.passwordLengthError'), trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  try {
    await loginFormRef.value.validate()
    loading.value = true

    const result = await authStore.login(loginForm.username, loginForm.password)
    
    if (result.success) {
      ElNotification({
        title: t('login.loginSuccess'),
        message: `${t('login.welcomeBack')}，${result.user.username}!`,
        type: 'success',
        duration: 3000
      })

      router.push('/')
    } else {
      ElMessage.error(result.message || t('login.loginFailed'))
    }
  } catch (error) {
    console.error('Login error:', error)
    ElMessage.error(t('login.loginFailed') + ': ' + (error.message || t('login.networkError')))
  } finally {
    loading.value = false
  }
}

// Check if already authenticated
onMounted(() => {
  if (authStore.isAuthenticated) {
    router.push('/')
  }
})
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--bg-tertiary) 0%, var(--primary-ultralight) 30%, var(--primary-light) 100%);
  padding: var(--spacing-2xl);
  position: relative;
}

.login-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(circle at 20% 80%, rgba(22, 119, 255, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 20%, rgba(82, 196, 26, 0.1) 0%, transparent 50%);
  pointer-events: none;
}

.login-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-primary);
  border-radius: var(--border-radius-2xl);
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-light);
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  padding: var(--spacing-4xl) var(--spacing-4xl) var(--spacing-2xl);
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-hover) 100%);
  color: var(--text-inverse);
  position: relative;
}

.login-header::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--secondary-color) 0%, var(--primary-color) 100%);
}

.login-header h1 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.login-header p {
  margin: 0;
  opacity: 0.95;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-medium);
}

.login-form {
  padding: var(--spacing-4xl);
}

.login-footer {
  text-align: center;
  padding: 0 var(--spacing-4xl) var(--spacing-4xl);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  border-top: 1px solid var(--border-light);
}

.login-footer p {
  margin: var(--spacing-xs) 0;
}

.login-footer .warning {
  color: var(--warning-color);
  font-weight: var(--font-weight-medium);
  background: rgba(250, 173, 20, 0.1);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-md);
  border: 1px solid rgba(250, 173, 20, 0.2);
}

/* Element Plus style overrides */
:deep(.el-form-item) {
  margin-bottom: var(--spacing-2xl);
}

:deep(.el-input__wrapper) {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-medium);
  box-shadow: var(--shadow-xs);
  transition: all var(--transition-fast);
}

:deep(.el-input__wrapper:hover) {
  border-color: var(--primary-light);
}

:deep(.el-input__wrapper.is-focus) {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(22, 119, 255, 0.1);
}

:deep(.el-input__inner) {
  color: var(--text-primary);
  font-weight: var(--font-weight-medium);
}

:deep(.el-input__prefix-inner) {
  color: var(--text-tertiary);
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-hover) 100%);
  border: none;
  border-radius: var(--border-radius-lg);
  font-weight: var(--font-weight-semibold);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-fast);
}

:deep(.el-button--primary:hover) {
  background: linear-gradient(135deg, var(--primary-hover) 0%, var(--primary-color) 100%);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

:deep(.el-button--primary:active) {
  transform: translateY(0);
}

/* Responsive design */
@media (max-width: 768px) {
  .login-container {
    padding: var(--spacing-lg);
  }
  
  .login-card {
    max-width: 100%;
  }
  
  .login-header {
    padding: var(--spacing-3xl) var(--spacing-2xl) var(--spacing-xl);
  }
  
  .login-header h1 {
    font-size: var(--font-size-2xl);
  }
  
  .login-form {
    padding: var(--spacing-2xl);
  }
  
  .login-footer {
    padding: 0 var(--spacing-2xl) var(--spacing-2xl);
  }
}

@media (max-width: 480px) {
  .login-container {
    padding: var(--spacing-md);
  }
  
  .login-header {
    padding: var(--spacing-2xl) var(--spacing-xl) var(--spacing-lg);
  }
  
  .login-header h1 {
    font-size: var(--font-size-xl);
  }
  
  .login-form {
    padding: var(--spacing-xl);
  }
  
  .login-footer {
    padding: 0 var(--spacing-xl) var(--spacing-xl);
  }
}
</style>

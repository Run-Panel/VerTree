<template>
  <div class="app-container">
    <!-- Modern Header -->
    <el-header class="app-header">
      <div class="header-left">
        <div class="brand">
          <div class="brand-icon">
            <el-icon size="28"><Box /></el-icon>
          </div>
          <div class="brand-text">
            <h2 class="brand-title">{{ $t('layout.appTitle') }}</h2>
          </div>
        </div>
      </div>
      
      <div class="header-right">
        <!-- API Documentation -->
        <el-button 
          class="header-action-btn docs-btn" 
          text 
          @click="$router.push('/docs')"
        >
          <el-icon><Document /></el-icon>
          <span class="mobile-hidden">API文档</span>
        </el-button>

        <!-- Language Switch -->
        <el-dropdown @command="handleLanguageCommand" class="language-dropdown">
          <el-button class="header-action-btn" text>
            <el-icon><Globe /></el-icon>
            <span class="mobile-hidden">{{ currentLocale === 'zh' ? $t('layout.chinese') : $t('layout.english') }}</span>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="zh">{{ $t('layout.chinese') }}</el-dropdown-item>
              <el-dropdown-item command="en">{{ $t('layout.english') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <!-- User Menu -->
        <el-dropdown @command="handleUserCommand" trigger="click" class="user-dropdown">
          <div class="user-profile">
            <el-avatar class="user-avatar" :size="36" :src="userAvatar">
              <el-icon><User /></el-icon>
            </el-avatar>
            <div class="user-info mobile-hidden">
              <span class="username">{{ currentUser?.username || 'Admin' }}</span>
              <span class="user-role">{{ $t('common.admin') || 'Administrator' }}</span>
            </div>
            <el-icon class="dropdown-arrow mobile-hidden"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">
                <el-icon><User /></el-icon>
                <span>{{ $t('layout.profile') }}</span>
              </el-dropdown-item>
              <el-dropdown-item command="change-password">
                <el-icon><Lock /></el-icon>
                <span>{{ $t('layout.changePassword') }}</span>
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>
                <span>{{ $t('layout.logout') }}</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container class="main-container">
      <!-- Modern Sidebar -->
      <el-aside :width="sidebarWidth" class="app-sidebar" :class="{ collapsed: sidebarCollapsed }">
        <div class="sidebar-header">
          <el-button 
            class="sidebar-toggle mobile-hidden" 
            text 
            @click="toggleSidebar"
            :icon="sidebarCollapsed ? Expand : Fold"
          />
        </div>
        
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          :collapse="sidebarCollapsed"
          :unique-opened="true"
        >
          <el-menu-item index="/dashboard" class="menu-item">
            <el-icon><Odometer /></el-icon>
            <template #title>{{ $t('nav.dashboard') }}</template>
          </el-menu-item>
          
          <el-menu-item index="/applications" class="menu-item">
            <el-icon><Files /></el-icon>
            <template #title>{{ $t('nav.applications') }}</template>
          </el-menu-item>
          
          <el-menu-item index="/versions" class="menu-item">
            <el-icon><Box /></el-icon>
            <template #title>{{ $t('nav.version') }}</template>
          </el-menu-item>
          
          <el-menu-item index="/channels" class="menu-item">
            <el-icon><Switch /></el-icon>
            <template #title>{{ $t('nav.channel') }}</template>
          </el-menu-item>
          
          <el-menu-item index="/statistics" class="menu-item">
            <el-icon><TrendCharts /></el-icon>
            <template #title>{{ $t('nav.statistics') }}</template>
          </el-menu-item>
          
          <el-menu-item index="/github" class="menu-item">
            <el-icon><Connection /></el-icon>
            <template #title>{{ $t('nav.github') }}</template>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- Main Content Area -->
      <el-main class="app-main">
        <div class="content-wrapper">
          <transition name="fade" mode="out-in">
            <router-view />
          </transition>
        </div>
      </el-main>
    </el-container>

    <!-- First Login Password Change Dialog -->
    <el-dialog
      v-model="showFirstLoginDialog"
      :title="$t('layout.firstLoginPasswordChange')"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <div class="dialog-content">
        <el-alert
          :title="$t('layout.firstLoginAlert')"
          type="warning"
          :closable="false"
          show-icon
          style="margin-bottom: 20px;"
        />
        
        <el-form
          ref="passwordForm"
          :model="passwordFormData"
          :rules="passwordRules"
          label-width="120px"
        >
          <el-form-item :label="$t('common.newPassword')" prop="newPassword">
            <el-input
              v-model="passwordFormData.newPassword"
              type="password"
              show-password
              :placeholder="$t('common.enterNewPassword')"
            />
          </el-form-item>
          
          <el-form-item :label="$t('common.confirmPassword')" prop="confirmPassword">
            <el-input
              v-model="passwordFormData.confirmPassword"
              type="password"
              show-password
              :placeholder="$t('common.confirmNewPassword')"
            />
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <el-button type="primary" @click="handlePasswordChange" style="width: 100%;">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Regular Password Change Dialog -->
    <el-dialog
      v-model="showChangePasswordDialog"
      :title="$t('layout.changePassword')"
      width="500px"
    >
      <el-form
        ref="passwordForm"
        :model="passwordFormData"
        :rules="passwordRules"
        label-width="120px"
      >
        <el-form-item :label="$t('common.currentPassword')" prop="currentPassword">
          <el-input
            v-model="passwordFormData.currentPassword"
            type="password"
            show-password
            :placeholder="$t('common.enterCurrentPassword')"
          />
        </el-form-item>
        
        <el-form-item :label="$t('common.newPassword')" prop="newPassword">
          <el-input
            v-model="passwordFormData.newPassword"
            type="password"
            show-password
            :placeholder="$t('common.enterNewPassword')"
          />
        </el-form-item>
        
        <el-form-item :label="$t('common.confirmPassword')" prop="confirmPassword">
          <el-input
            v-model="passwordFormData.confirmPassword"
            type="password"
            show-password
            :placeholder="$t('common.confirmNewPassword')"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="cancelPasswordChange">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handlePasswordChange">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, ElDialog, ElForm, ElFormItem, ElInput, ElButton } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import { Expand, Fold, Files, Document, Connection } from '@element-plus/icons-vue'

const router = useRouter()
const { locale, t } = useI18n()
const authStore = useAuthStore()

// Reactive state
const sidebarCollapsed = ref(false)
const userAvatar = ref('')
const showFirstLoginDialog = ref(false)
const showChangePasswordDialog = ref(false)
const passwordForm = ref(null)
const passwordFormData = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// Computed properties
const currentLocale = computed(() => locale.value)
const currentUser = computed(() => authStore.currentUser)
const sidebarWidth = computed(() => sidebarCollapsed.value ? '80px' : '280px')

// Password form rules
const passwordRules = {
  currentPassword: [
    { 
      validator: (rule, value, callback) => {
        if (!currentUser.value?.first_login && (!value || value.trim() === '')) {
          callback(new Error(t('common.currentPasswordRequired')))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ],
  newPassword: [
    { required: true, message: t('common.newPasswordRequired'), trigger: 'blur' },
    { min: 6, message: t('common.passwordLengthError'), trigger: 'blur' },
    { 
      validator: (rule, value, callback) => {
        if (value && value.length >= 6) {
          const hasLetter = /[a-zA-Z]/.test(value)
          const hasNumber = /[0-9]/.test(value)
          if (!hasLetter || !hasNumber) {
            callback(new Error(t('common.passwordComplexityError')))
          } else {
            callback()
          }
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ],
  confirmPassword: [
    { required: true, message: t('common.confirmPasswordRequired'), trigger: 'blur' },
    { 
      validator: (rule, value, callback) => {
        if (value !== passwordFormData.value.newPassword) {
          callback(new Error(t('common.passwordMismatch')))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ]
}

// Check for first login on mount
onMounted(() => {
  if (currentUser.value?.first_login) {
    setTimeout(() => {
      showFirstLoginPasswordDialog()
    }, 1000)
  }
})

// Sidebar toggle function
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// Handle language switch
const handleLanguageCommand = (command) => {
  locale.value = command
  ElMessage.success(t('common.success'))
}

// Handle user menu commands
const handleUserCommand = async (command) => {
  switch (command) {
    case 'profile':
      ElMessage.info(t('common.developing') || 'Feature under development')
      break
    
    case 'change-password':
      showChangePasswordDialogForm()
      break
    
    case 'logout':
      await handleLogout()
      break
  }
}

// Handle logout with confirmation
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm(
      t('layout.confirmLogout'),
      t('common.confirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    )
    
    await authStore.logout()
    ElMessage.success(t('layout.logoutSuccess'))
    router.push('/login')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Logout error:', error)
      ElMessage.error(t('layout.logoutFailed'))
    }
  }
}

// Show change password dialog
// Show first login password dialog (mandatory)
const showFirstLoginPasswordDialog = () => {
  showFirstLoginDialog.value = true
  passwordFormData.value = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

// Show regular change password dialog
const showChangePasswordDialogForm = () => {
  showChangePasswordDialog.value = true
  passwordFormData.value = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

// Handle password change
const handlePasswordChange = async () => {
  if (!passwordForm.value) return
  
  try {
    await passwordForm.value.validate()
    
    const currentPassword = currentUser.value?.first_login ? '' : passwordFormData.value.currentPassword
    const result = await authStore.changePassword(currentPassword, passwordFormData.value.newPassword)
    
    if (result.success) {
      ElMessage.success(t('common.passwordChangeSuccess'))
      
      // Close dialogs
      showFirstLoginDialog.value = false
      showChangePasswordDialog.value = false
      
      // If it was first login, refresh user data and show success
      if (currentUser.value?.first_login) {
        await authStore.fetchProfile() // Refresh user data
        ElMessage.success(t('layout.firstLoginCompleted'))
      } else {
        // For regular password change, logout and redirect to login
        ElMessage.success(t('common.passwordChangeSuccess') + ', ' + t('layout.pleaseLoginAgain'))
        await authStore.logout()
        router.push('/login')
      }
    } else {
      ElMessage.error(result.message || t('common.passwordChangeFailed'))
    }
  } catch (error) {
    ElMessage.error(t('common.passwordChangeFailed') + ': ' + error.message)
  }
}

// Cancel password change (only allowed for non-first-login)
const cancelPasswordChange = () => {
  if (currentUser.value?.first_login) {
    ElMessage.warning(t('layout.firstLoginMandatory'))
    return
  }
  
  showChangePasswordDialog.value = false
  passwordFormData.value = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}
</script>

<style scoped>
/* Modern App Container */
.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
}

/* Modern Header Styles */
.app-header {
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--spacing-2xl);
  height: var(--header-height);
  position: relative;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.brand-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  border-radius: var(--border-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-inverse);
  box-shadow: var(--shadow-sm);
}

.brand-title {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.header-action-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-lg);
  color: var(--text-secondary);
  transition: all var(--transition-normal);
}

.header-action-btn:hover {
  background: var(--bg-tertiary);
  color: var(--primary-color);
}

.user-profile {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm);
  border-radius: var(--border-radius-lg);
  cursor: pointer;
  transition: all var(--transition-normal);
}

.user-profile:hover {
  background: var(--bg-tertiary);
}

.user-avatar {
  box-shadow: var(--shadow-sm);
  border: 2px solid var(--primary-light);
}

.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.username {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  line-height: 1.2;
}

.user-role {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  line-height: 1.2;
}

.dropdown-arrow {
  color: var(--text-tertiary);
  transition: transform var(--transition-fast);
}

.user-dropdown:hover .dropdown-arrow {
  transform: rotate(180deg);
}

/* Main Container */
.main-container {
  flex: 1;
  overflow: hidden;
}

/* Modern Sidebar Styles */
.app-sidebar {
  background: var(--bg-dark);
  transition: width var(--transition-normal);
  border-right: 1px solid var(--bg-dark-light);
  box-shadow: var(--shadow-md);
  position: relative;
  z-index: 90;
}

.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 0 var(--spacing-lg);
  border-bottom: 1px solid var(--bg-dark-light);
}

.sidebar-toggle {
  color: var(--text-inverse) !important;
  background: transparent !important;
  border: none !important;
  padding: var(--spacing-sm) !important;
  border-radius: var(--border-radius-md) !important;
  transition: all var(--transition-normal) !important;
}

.sidebar-toggle:hover {
  background: var(--bg-dark-light) !important;
  color: var(--primary-color) !important;
}

.sidebar-menu {
  border-right: none !important;
  background: transparent !important;
  height: calc(100vh - var(--header-height) - 60px);
  padding: var(--spacing-lg) 0;
}

.sidebar-menu :deep(.el-menu-item) {
  margin: 0 var(--spacing-md) var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-lg) !important;
  color: rgba(255, 255, 255, 0.8) !important;
  font-weight: var(--font-weight-medium);
  transition: all var(--transition-normal) !important;
  height: 48px;
  line-height: 48px;
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background: var(--bg-dark-light) !important;
  color: var(--primary-color) !important;
  transform: translateX(4px);
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, var(--primary-color), var(--primary-hover)) !important;
  color: var(--text-inverse) !important;
  font-weight: var(--font-weight-semibold);
  box-shadow: var(--shadow-sm);
}

.sidebar-menu :deep(.el-menu-item .el-icon) {
  margin-right: var(--spacing-md);
  font-size: 18px;
}

/* Collapsed Sidebar Styles */
.app-sidebar.collapsed {
  width: var(--sidebar-collapsed-width) !important;
}

.app-sidebar.collapsed .sidebar-menu :deep(.el-menu-item) {
  justify-content: center;
  margin: 0 var(--spacing-sm) var(--spacing-sm) var(--spacing-sm);
}

/* Main Content Area */
.app-main {
  background: var(--bg-secondary);
  padding: var(--spacing-2xl);
  overflow-y: auto;
  position: relative;
}

.content-wrapper {
  max-width: 1600px;
  margin: 0 auto;
  min-height: calc(100vh - var(--header-height) - var(--spacing-4xl));
}

/* Responsive Design */
@media (max-width: 768px) {
  .app-header {
    padding: 0 var(--spacing-lg);
  }
  
  .brand-title {
    font-size: var(--font-size-lg);
  }
  
  .header-right {
    gap: var(--spacing-sm);
  }
  
  .app-sidebar {
    position: fixed;
    left: -280px;
    top: var(--header-height);
    height: calc(100vh - var(--header-height));
    z-index: 1000;
    transition: left var(--transition-normal);
  }
  
  .app-sidebar.mobile-open {
    left: 0;
  }
  
  .app-main {
    padding: var(--spacing-lg);
    margin-left: 0 !important;
  }
  
  .sidebar-header {
    display: none;
  }
}

@media (max-width: 480px) {
  .app-header {
    padding: 0 var(--spacing-md);
  }
  
  .brand-title {
    display: none;
  }
  
  .user-info {
    display: none;
  }
  
  .app-main {
    padding: var(--spacing-md);
  }
}

/* Dark Theme Support */
[data-theme="dark"] .app-header {
  background: var(--bg-secondary);
  border-bottom-color: var(--border-color);
}

[data-theme="dark"] .brand-icon {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
}

/* Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-normal);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="formRules"
    label-width="120px"
    class="github-repo-form"
  >
    <el-form-item label="仓库URL" prop="repository_url">
      <el-input
        v-model="form.repository_url"
        placeholder="https://github.com/owner/repo"
        :disabled="loading"
        @blur="parseRepositoryUrl"
      >
        <template #prepend>
          <el-icon><Link /></el-icon>
        </template>
        <template #append>
          <el-button 
            :loading="validating" 
            @click="validateRepository"
            :disabled="!form.repository_url"
          >
            验证
          </el-button>
        </template>
      </el-input>
      <div class="form-help-text">
        输入GitHub仓库URL，支持公开和私有仓库
      </div>
    </el-form-item>

    <el-form-item label="所有者" prop="owner_name">
      <el-input 
        v-model="form.owner_name" 
        :disabled="loading"
        placeholder="GitHub用户名或组织名"
      />
    </el-form-item>

    <el-form-item label="仓库名" prop="repo_name">
      <el-input 
        v-model="form.repo_name" 
        :disabled="loading"
        placeholder="仓库名称"
      />
    </el-form-item>

    <el-form-item label="分支" prop="branch_name">
      <el-input 
        v-model="form.branch_name" 
        :disabled="loading"
        placeholder="默认为 main"
      />
    </el-form-item>

    <!-- Authentication Type Selection -->
    <el-form-item label="认证方式" prop="auth_type">
      <el-radio-group v-model="form.auth_type" @change="onAuthTypeChange">
        <el-radio value="personal_token">
          <div class="auth-option">
            <div class="auth-title">Personal Access Token</div>
            <div class="auth-desc">适用于个人仓库或小型团队</div>
          </div>
        </el-radio>
        <el-radio value="github_app">
          <div class="auth-option">
            <div class="auth-title">GitHub Apps <el-tag type="success" size="small">推荐</el-tag></div>
            <div class="auth-desc">适用于组织仓库，更安全、权限更精细</div>
          </div>
        </el-radio>
      </el-radio-group>
    </el-form-item>

    <!-- Personal Token Authentication -->
    <template v-if="form.auth_type === 'personal_token'">
      <el-form-item label="访问令牌" prop="access_token">
        <el-input
          v-model="form.access_token"
          type="password"
          show-password
          :disabled="loading"
          placeholder="GitHub Personal Access Token"
        >
          <template #prepend>
            <el-icon><Key /></el-icon>
          </template>
          <template #append>
            <el-button 
              :loading="testingToken" 
              @click="testToken"
              :disabled="!form.access_token"
            >
              测试
            </el-button>
          </template>
        </el-input>
        <div class="form-help-text">
          需要具有 <code>repo</code> 和 <code>admin:repo_hook</code> 权限的GitHub Token
        </div>
      </el-form-item>
    </template>

    <!-- GitHub Apps Authentication -->
    <template v-if="form.auth_type === 'github_app'">
      <div class="github-apps-section">
        <el-alert
          title="GitHub Apps 配置向导"
          description="GitHub Apps 提供更安全的认证方式，特别适用于组织仓库。我们将引导您完成设置。"
          type="info"
          :closable="false"
          show-icon
        />
        
        <el-steps :active="currentStep" :space="200" class="github-apps-steps">
          <el-step title="创建GitHub App" description="在GitHub上创建应用" />
          <el-step title="获取配置信息" description="收集必要的认证信息" />
          <el-step title="安装配置" description="安装并测试连接" />
        </el-steps>

        <!-- Step 1: GitHub App Creation Guide -->
        <div v-if="currentStep === 0" class="step-content">
          <h3>第一步：创建GitHub App</h3>
          <el-card class="guide-card">
            <p><strong>还没有GitHub App？</strong> 请按照以下步骤创建：</p>
            <ol>
              <li>访问GitHub组织设置页面或个人设置</li>
              <li>进入 "Developer settings" > "GitHub Apps"</li>
              <li>点击 "New GitHub App"</li>
              <li>
                填写基本信息：
                <ul>
                  <li><strong>App name</strong>: VerTree-YourOrg (必须唯一)</li>
                  <li><strong>Homepage URL</strong>: 您的VerTree部署地址</li>
                </ul>
              </li>
              <li>
                设置权限：
                <ul>
                  <li><strong>Repository permissions</strong>:
                    <ul>
                      <li>Contents: Read</li>
                      <li>Metadata: Read</li>
                      <li>Webhooks: Write</li>
                    </ul>
                  </li>
                </ul>
              </li>
              <li>勾选 "Release" 事件</li>
              <li>点击创建应用</li>
            </ol>
            
            <div class="guide-actions">
              <el-button 
                type="primary" 
                @click="openGitHubAppsPage"
                icon="External"
              >
                去GitHub创建应用
              </el-button>
              <el-button @click="currentStep = 1">
                我已经有GitHub App了
              </el-button>
            </div>
          </el-card>
        </div>

        <!-- Step 2: Configuration Collection -->
        <div v-if="currentStep === 1" class="step-content">
          <h3>第二步：填写GitHub App配置信息</h3>
          
          <el-form-item label="GitHub App ID" prop="github_app_id">
            <el-input
              v-model="form.github_app_id"
              placeholder="例如：123456"
              :disabled="loading"
            >
              <template #prepend>
                <el-icon><Stamp /></el-icon>
              </template>
            </el-input>
            <div class="form-help-text">
              在GitHub App详情页面找到App ID，通常是一个6位数字
            </div>
          </el-form-item>

          <el-form-item label="私钥" prop="private_key">
            <el-input
              v-model="form.private_key"
              type="textarea"
              :rows="8"
              placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;...&#10;-----END RSA PRIVATE KEY-----"
              :disabled="loading"
            />
            <div class="form-help-text">
              在GitHub App设置页面生成并下载私钥文件，将完整内容粘贴到此处
            </div>
          </el-form-item>

          <el-form-item label="Installation ID" prop="installation_id">
            <el-input
              v-model="form.installation_id"
              placeholder="例如：12345678"
              :disabled="loading"
            >
              <template #prepend>
                <el-icon><Monitor /></el-icon>
              </template>
              <template #append>
                <el-button 
                  :loading="loadingInstallations" 
                  @click="getInstallations"
                  :disabled="!form.github_app_id || !form.private_key"
                >
                  获取列表
                </el-button>
              </template>
            </el-input>
            <div class="form-help-text">
              安装GitHub App后获得的Installation ID，或点击"获取列表"自动获取
            </div>
          </el-form-item>

          <!-- Installation List -->
          <div v-if="installationsList.length > 0" class="installations-section">
            <h4>可用的安装列表：</h4>
            <el-card 
              v-for="installation in installationsList" 
              :key="installation.id"
              class="installation-card"
              :class="{ 'selected': form.installation_id == installation.id }"
              @click="selectInstallation(installation)"
            >
              <div class="installation-info">
                <div class="installation-header">
                  <el-avatar :src="installation.account.avatar_url" :size="40" />
                  <div class="account-info">
                    <div class="account-name">{{ installation.account.login }}</div>
                    <div class="account-type">{{ installation.account.type }}</div>
                  </div>
                  <el-tag v-if="form.installation_id == installation.id" type="success">已选择</el-tag>
                </div>
                <div class="installation-details">
                  <span>ID: {{ installation.id }}</span>
                  <span>权限: {{ installation.permissions ? Object.keys(installation.permissions).length : 0 }} 项</span>
                  <span>仓库: {{ installation.repository_selection === 'all' ? '全部' : '选择的' }}</span>
                </div>
              </div>
            </el-card>
          </div>

          <div class="step-actions">
            <el-button @click="currentStep = 0">上一步</el-button>
            <el-button 
              type="primary" 
              @click="currentStep = 2"
              :disabled="!form.github_app_id || !form.private_key || !form.installation_id"
            >
              下一步
            </el-button>
          </div>
        </div>

        <!-- Step 3: Test and Complete -->
        <div v-if="currentStep === 2" class="step-content">
          <h3>第三步：测试连接</h3>
          
          <el-card class="test-section">
            <div class="test-header">
              <h4>连接测试</h4>
              <el-button 
                type="primary" 
                @click="testGitHubAppConnection"
                :loading="testingGitHubApp"
                :disabled="!form.github_app_id || !form.private_key || !form.installation_id"
              >
                测试GitHub App连接
              </el-button>
            </div>
            
            <div v-if="githubAppTestResult" class="test-result">
              <el-alert
                :title="githubAppTestResult.success ? '连接测试成功！' : '连接测试失败'"
                :type="githubAppTestResult.success ? 'success' : 'error'"
                :description="githubAppTestResult.message"
                show-icon
                :closable="false"
              />
              
              <div v-if="githubAppTestResult.success && githubAppTestResult.data" class="app-info">
                <h4>GitHub App 信息：</h4>
                <el-descriptions :column="2" size="small">
                  <el-descriptions-item label="App名称">
                    {{ githubAppTestResult.data.app_name || '未知' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="所有者">
                    {{ githubAppTestResult.data.owner || '未知' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="安装状态">
                    <el-tag type="success">已安装</el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="权限">
                    {{ githubAppTestResult.data.permissions || '未知' }}
                  </el-descriptions-item>
                </el-descriptions>
              </div>
            </div>
          </el-card>

          <div class="step-actions">
            <el-button @click="currentStep = 1">上一步</el-button>
            <el-button 
              type="success" 
              @click="completeGitHubAppSetup"
              :disabled="!githubAppTestResult?.success"
            >
              完成配置
            </el-button>
          </div>
        </div>
      </div>
    </template>

    <el-form-item label="默认发布渠道" prop="default_channel">
      <el-select v-model="form.default_channel" :disabled="loading" style="width: 100%">
        <el-option label="Stable（稳定版）" value="stable" />
        <el-option label="Beta（测试版）" value="beta" />
        <el-option label="Alpha（内测版）" value="alpha" />
        <el-option label="Development（开发版）" value="development" />
      </el-select>
      <div class="form-help-text">
        GitHub Release发布后将自动同步到此渠道
      </div>
    </el-form-item>

    <el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="启用同步" label-width="80px">
            <el-switch
              v-model="form.is_active"
              :disabled="loading"
              active-text="启用"
              inactive-text="禁用"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="自动同步" label-width="80px">
            <el-switch
              v-model="form.auto_sync"
              :disabled="loading"
              active-text="启用"
              inactive-text="禁用"
            />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form-item>

    <el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="自动发布" label-width="80px">
            <el-switch
              v-model="form.auto_publish"
              :disabled="loading"
              active-text="启用"
              inactive-text="禁用"
            />
          </el-form-item>
        </el-col>
      </el-row>
      <div class="form-help-text">
        自动发布：同步后自动发布版本到对应渠道
      </div>
    </el-form-item>

    <!-- Validation Results -->
    <el-form-item v-if="validationResult">
      <el-alert
        :title="validationResult.success ? '验证成功' : '验证失败'"
        :type="validationResult.success ? 'success' : 'error'"
        :description="validationResult.message"
        show-icon
        :closable="false"
      />
      <div v-if="validationResult.success && validationResult.data" class="validation-info">
        <h4>仓库信息：</h4>
        <el-descriptions :column="2" size="small">
          <el-descriptions-item label="仓库全名">
            {{ validationResult.data.repository?.full_name || '未知' }}
          </el-descriptions-item>
          <el-descriptions-item label="描述">
            {{ validationResult.data.repository?.description || '无描述' }}
          </el-descriptions-item>
          <el-descriptions-item label="默认分支">
            {{ validationResult.data.repository?.default_branch || '未知' }}
          </el-descriptions-item>
          <el-descriptions-item label="可见性">
            <el-tag :type="validationResult.data.repository?.private ? 'warning' : 'success'">
              {{ validationResult.data.repository?.private ? '私有' : '公开' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="最新Release" :span="2">
            <div v-if="validationResult.data.latest_release" class="release-info">
              <div class="release-header">
                <el-tag type="success" size="small">{{ validationResult.data.latest_release.tag_name }}</el-tag>
                <span class="release-name">{{ validationResult.data.latest_release.name }}</span>
              </div>
              <div class="release-date">
                发布时间：{{ formatReleaseDate(validationResult.data.latest_release.published_at) }}
              </div>
              <div v-if="validationResult.data.latest_release.assets?.length" class="release-assets">
                <span class="assets-label">资产文件：</span>
                <el-tag 
                  v-for="asset in validationResult.data.latest_release.assets.slice(0, 3)" 
                  :key="asset.id"
                  size="small"
                  class="asset-tag"
                >
                  {{ asset.name }} ({{ formatFileSize(asset.size) }})
                </el-tag>
                <span v-if="validationResult.data.latest_release.assets.length > 3" class="more-assets">
                  +{{ validationResult.data.latest_release.assets.length - 3 }}个文件
                </span>
              </div>
            </div>
            <span v-else class="no-release">无Release</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-form-item>

    <!-- Token Test Results -->
    <el-form-item v-if="tokenTestResult">
      <el-alert
        :title="tokenTestResult.success ? 'Token验证成功' : 'Token验证失败'"
        :type="tokenTestResult.success ? 'success' : 'error'"
        :description="tokenTestResult.message"
        show-icon
        :closable="false"
      />
      <div v-if="tokenTestResult.success && tokenTestResult.permissions" class="token-permissions">
        <h4>Token权限：</h4>
        <el-tag
          v-for="permission in tokenTestResult.permissions"
          :key="permission"
          size="small"
          class="permission-tag"
        >
          {{ permission }}
        </el-tag>
      </div>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Link, Key, Stamp, Monitor } from '@element-plus/icons-vue'
import { 
  validateGitHubRepository, 
  testGitHubToken,
  getGitHubAppInstallations,
  testGitHubApp
} from '@/api/github'

// Props & Emits
const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'validation-change'])

// Refs
const formRef = ref()
const validating = ref(false)
const testingToken = ref(false)
const validationResult = ref(null)
const tokenTestResult = ref(null)

// GitHub Apps related refs
const currentStep = ref(0)
const loadingInstallations = ref(false)
const testingGitHubApp = ref(false)
const installationsList = ref([])
const githubAppTestResult = ref(null)

// Form data
const form = reactive({
  repository_url: '',
  owner_name: '',
  repo_name: '',
  branch_name: 'main',
  auth_type: 'personal_token',
  access_token: '',
  github_app_id: '',
  installation_id: '',
  private_key: '',
  default_channel: 'stable',
  is_active: true,
  auto_sync: true,
  auto_publish: false,
  ...props.modelValue
})

// Form rules - dynamic based on auth type
const getFormRules = () => {
  const baseRules = {
    repository_url: [
      { required: true, message: '请输入GitHub仓库URL', trigger: 'blur' },
      { 
        pattern: /^https:\/\/github\.com\/[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+\/?$/,
        message: '请输入有效的GitHub仓库URL',
        trigger: 'blur'
      }
    ],
    owner_name: [
      { required: true, message: '请输入仓库所有者', trigger: 'blur' },
      { min: 1, max: 100, message: '所有者名称长度在 1 到 100 个字符', trigger: 'blur' }
    ],
    repo_name: [
      { required: true, message: '请输入仓库名称', trigger: 'blur' },
      { min: 1, max: 100, message: '仓库名称长度在 1 到 100 个字符', trigger: 'blur' }
    ],
    branch_name: [
      { required: true, message: '请输入分支名称', trigger: 'blur' },
      { min: 1, max: 100, message: '分支名称长度在 1 到 100 个字符', trigger: 'blur' }
    ],
    auth_type: [
      { required: true, message: '请选择认证方式', trigger: 'change' }
    ],
    default_channel: [
      { required: true, message: '请选择默认发布渠道', trigger: 'change' }
    ]
  }

  if (form.auth_type === 'personal_token') {
    baseRules.access_token = [
      { required: true, message: '请输入GitHub访问令牌', trigger: 'blur' },
      { min: 20, message: 'Token长度至少20个字符', trigger: 'blur' }
    ]
  } else if (form.auth_type === 'github_app') {
    baseRules.github_app_id = [
      { required: true, message: '请输入GitHub App ID', trigger: 'blur' },
      { pattern: /^\d+$/, message: 'App ID必须是数字', trigger: 'blur' }
    ]
    baseRules.installation_id = [
      { required: true, message: '请输入Installation ID', trigger: 'blur' },
      { pattern: /^\d+$/, message: 'Installation ID必须是数字', trigger: 'blur' }
    ]
    baseRules.private_key = [
      { required: true, message: '请输入GitHub App私钥', trigger: 'blur' },
      { 
        validator: (rule, value, callback) => {
          if (!value.includes('-----BEGIN') || !value.includes('-----END')) {
            callback(new Error('私钥格式不正确，必须包含完整的PEM格式'))
          } else {
            callback()
          }
        }, 
        trigger: 'blur' 
      }
    ]
  }

  return baseRules
}

const formRules = ref(getFormRules())

// Methods
const parseRepositoryUrl = () => {
  const url = form.repository_url.trim()
  if (!url) return
  
  // Parse GitHub URL
  const match = url.match(/^https:\/\/github\.com\/([a-zA-Z0-9_.-]+)\/([a-zA-Z0-9_.-]+)\/?$/)
  if (match) {
    form.owner_name = match[1]
    form.repo_name = match[2]
    
    // Clear previous validation results
    validationResult.value = null
    tokenTestResult.value = null
  }
}

const validateRepository = async () => {
  if (!form.repository_url || !form.owner_name || !form.repo_name) {
    ElMessage.warning('请先填写完整的仓库信息')
    return
  }
  
  validating.value = true
  validationResult.value = null
  
  try {
    const response = await validateGitHubRepository({
      repository_url: form.repository_url,
      owner_name: form.owner_name,
      repo_name: form.repo_name,
      access_token: form.access_token || undefined
    })
    
    if (response.data) {
      validationResult.value = {
        success: true,
        message: '仓库验证成功',
        data: response.data
      }
      
      // Auto-fill branch name if not set
      if (!form.branch_name && response.data.repository?.default_branch) {
        form.branch_name = response.data.repository.default_branch
      }
      
      emit('validation-change', true)
    }
  } catch (error) {
    validationResult.value = {
      success: false,
      message: error.response?.data?.message || '仓库验证失败'
    }
    emit('validation-change', false)
  } finally {
    validating.value = false
  }
}

const testToken = async () => {
  if (!form.access_token) {
    ElMessage.warning('请先输入GitHub访问令牌')
    return
  }
  
  testingToken.value = true
  tokenTestResult.value = null
  
  try {
    const response = await testGitHubToken({
      access_token: form.access_token,
      repository_url: form.repository_url || undefined
    })
    
    if (response.data) {
      tokenTestResult.value = {
        success: true,
        message: 'Token验证成功',
        permissions: response.data.permissions || []
      }
    }
  } catch (error) {
    tokenTestResult.value = {
      success: false,
      message: error.response?.data?.message || 'Token验证失败'
    }
  } finally {
    testingToken.value = false
  }
}

const validate = async () => {
  if (!formRef.value) return false
  
  try {
    await formRef.value.validate()
    return true
  } catch {
    return false
  }
}

const clearValidate = () => {
  formRef.value?.clearValidate()
  validationResult.value = null
  tokenTestResult.value = null
  githubAppTestResult.value = null
}

// GitHub Apps specific methods
const onAuthTypeChange = () => {
  // Reset validation results when auth type changes
  validationResult.value = null
  tokenTestResult.value = null
  githubAppTestResult.value = null
  installationsList.value = []
  currentStep.value = 0
  
  // Update form rules
  formRules.value = getFormRules()
  
  // Clear validation
  clearValidate()
}

const openGitHubAppsPage = () => {
  // Determine if user likely has an organization based on repository URL
  let url = 'https://github.com/settings/apps'
  
  if (form.repository_url) {
    const match = form.repository_url.match(/^https:\/\/github\.com\/([a-zA-Z0-9_.-]+)\//)
    if (match && match[1]) {
      // Try organization first
      url = `https://github.com/organizations/${match[1]}/settings/apps`
    }
  }
  
  window.open(url, '_blank')
}

const getInstallations = async () => {
  if (!form.github_app_id || !form.private_key) {
    ElMessage.warning('请先填写GitHub App ID和私钥')
    return
  }
  
  loadingInstallations.value = true
  installationsList.value = []
  
  try {
    const response = await getGitHubAppInstallations({
      github_app_id: form.github_app_id,
      private_key: form.private_key
    })
    
    if (response.data?.installations) {
      installationsList.value = response.data.installations
      ElMessage.success(`找到 ${installationsList.value.length} 个可用的安装`)
    } else {
      ElMessage.warning('没有找到可用的GitHub App安装')
    }
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '获取安装列表失败')
    console.error('Failed to get installations:', error)
  } finally {
    loadingInstallations.value = false
  }
}

const selectInstallation = (installation) => {
  form.installation_id = installation.id.toString()
  ElMessage.success(`已选择安装：${installation.account.login}`)
}

const testGitHubAppConnection = async () => {
  if (!form.github_app_id || !form.private_key || !form.installation_id) {
    ElMessage.warning('请先填写完整的GitHub App配置信息')
    return
  }
  
  testingGitHubApp.value = true
  githubAppTestResult.value = null
  
  try {
    const response = await testGitHubApp({
      github_app_id: form.github_app_id,
      private_key: form.private_key,
      installation_id: form.installation_id,
      repository_url: form.repository_url || undefined
    })
    
    if (response.data) {
      githubAppTestResult.value = {
        success: true,
        message: 'GitHub App连接测试成功',
        data: response.data
      }
      ElMessage.success('GitHub App连接测试成功！')
    }
  } catch (error) {
    githubAppTestResult.value = {
      success: false,
      message: error.response?.data?.message || 'GitHub App连接测试失败'
    }
    ElMessage.error(error.response?.data?.message || 'GitHub App连接测试失败')
  } finally {
    testingGitHubApp.value = false
  }
}

const completeGitHubAppSetup = () => {
  ElMessage.success('GitHub Apps配置完成！您现在可以保存仓库绑定了。')
  emit('validation-change', true)
}

// Utility functions for formatting
const formatReleaseDate = (dateString) => {
  if (!dateString) return '未知'
  try {
    return new Date(dateString).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return '未知'
  }
}

const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  
  const sizes = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  
  while (size >= 1024 && i < sizes.length - 1) {
    size /= 1024
    i++
  }
  
  return `${size.toFixed(i === 0 ? 0 : 1)} ${sizes[i]}`
}

// Watch form changes
watch(form, (newValue) => {
  emit('update:modelValue', { ...newValue })
}, { deep: true })

// Watch auth type changes to update rules
watch(() => form.auth_type, () => {
  formRules.value = getFormRules()
}, { immediate: true })

// Expose methods
defineExpose({
  validate,
  clearValidate,
  validateRepository,
  testToken
})
</script>

<style scoped>
.github-repo-form {
  max-width: 600px;
}

.form-help-text {
  margin-top: 5px;
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}

.form-help-text code {
  background: #f5f7fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
  font-size: 11px;
}

.validation-info {
  margin-top: 16px;
  padding: 16px;
  background: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #e1f5fe;
}

.validation-info h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #303133;
}

.token-permissions {
  margin-top: 16px;
  padding: 16px;
  background: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #e1f5fe;
}

.token-permissions h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #303133;
}

.permission-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}

/* Release Info Styles */
.release-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.release-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.release-name {
  font-weight: 500;
  color: #303133;
}

.release-date {
  font-size: 12px;
  color: #909399;
}

.release-assets {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
}

.assets-label {
  font-size: 12px;
  color: #606266;
  margin-right: 4px;
}

.asset-tag {
  margin: 0;
}

.more-assets {
  font-size: 12px;
  color: #909399;
  margin-left: 4px;
}

.no-release {
  color: #909399;
  font-style: italic;
}

:deep(.el-descriptions-item__content) {
  word-break: break-all;
}

/* GitHub Apps Section Styles */
.github-apps-section {
  margin-top: 20px;
}

.github-apps-steps {
  margin: 24px 0;
}

.step-content {
  margin-top: 24px;
}

.step-content h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #303133;
}

.guide-card {
  margin-bottom: 16px;
}

.guide-card ol {
  margin: 16px 0;
  padding-left: 20px;
}

.guide-card li {
  margin-bottom: 8px;
  line-height: 1.6;
}

.guide-card ul {
  margin: 8px 0;
  padding-left: 20px;
}

.guide-actions {
  margin-top: 20px;
  text-align: center;
}

.guide-actions .el-button {
  margin: 0 8px;
}

.step-actions {
  margin-top: 24px;
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.step-actions .el-button {
  margin: 0 8px;
}

/* Authentication Type Selection */
.auth-option {
  margin-left: 8px;
}

.auth-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.auth-desc {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}

/* Installation List Styles */
.installations-section {
  margin-top: 16px;
}

.installations-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #303133;
}

.installation-card {
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.installation-card:hover {
  border-color: #409eff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.installation-card.selected {
  border-color: #67c23a;
  background-color: #f0f9ff;
}

.installation-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.installation-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.account-info {
  flex: 1;
}

.account-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 2px;
}

.account-type {
  font-size: 12px;
  color: #909399;
  text-transform: capitalize;
}

.installation-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 12px;
  color: #606266;
}

/* Test Section Styles */
.test-section {
  margin-bottom: 16px;
}

.test-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.test-header h4 {
  margin: 0;
  font-size: 14px;
  color: #303133;
}

.test-result {
  margin-top: 16px;
}

.app-info {
  margin-top: 16px;
  padding: 16px;
  background: #f0f9ff;
  border-radius: 6px;
  border: 1px solid #e1f5fe;
}

.app-info h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #303133;
}

/* Radio Group Spacing */
:deep(.el-radio) {
  width: 100%;
  margin-bottom: 12px;
  margin-right: 0;
}

:deep(.el-radio__label) {
  width: 100%;
}

/* Form Item Spacing for GitHub Apps */
.github-apps-section .el-form-item {
  margin-bottom: 20px;
}

/* Alert Styling */
.github-apps-section .el-alert {
  margin-bottom: 20px;
}

/* Responsive Design */
@media (max-width: 768px) {
  .github-apps-steps {
    margin: 16px 0;
  }
  
  .github-apps-steps :deep(.el-steps) {
    flex-direction: column;
  }
  
  .installation-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .installation-details {
    flex-direction: column;
    gap: 4px;
  }
  
  .guide-actions {
    text-align: left;
  }
  
  .guide-actions .el-button {
    display: block;
    width: 100%;
    margin: 8px 0;
  }
}
</style>

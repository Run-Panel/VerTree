<template>
  <div class="github-management-container">
    <div class="page-header">
      <h1>GitHub集成管理</h1>
      <p>管理应用的GitHub仓库绑定和版本同步</p>
    </div>

    <!-- Statistics Cards -->
    <div class="stats-section">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ stats.total_repositories }}</div>
              <div class="stat-label">绑定仓库</div>
            </div>
            <el-icon class="stat-icon"><Collection /></el-icon>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ stats.active_repositories }}</div>
              <div class="stat-label">活跃仓库</div>
            </div>
            <el-icon class="stat-icon success"><Check /></el-icon>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ stats.syncing_repositories }}</div>
              <div class="stat-label">同步中</div>
            </div>
            <el-icon class="stat-icon warning"><Loading /></el-icon>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ stats.total_releases }}</div>
              <div class="stat-label">同步版本</div>
            </div>
            <el-icon class="stat-icon info"><DocumentCopy /></el-icon>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- Repository List -->
    <div class="content-section">
      <div class="section-header">
        <h2>仓库列表</h2>
        <div class="header-actions">
          <el-button @click="refreshData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            绑定仓库
          </el-button>
        </div>
      </div>

      <!-- Filters -->
      <div class="filters-section">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-select v-model="filters.app_id" placeholder="选择应用" clearable @change="handleFilterChange">
              <el-option
                v-for="app in applications"
                :key="app.app_id"
                :label="app.name"
                :value="app.app_id"
              />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.sync_status" placeholder="同步状态" clearable @change="handleFilterChange">
              <el-option label="全部" value="" />
              <el-option label="成功" value="success" />
              <el-option label="进行中" value="syncing" />
              <el-option label="失败" value="failed" />
              <el-option label="等待中" value="pending" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filters.is_active" placeholder="状态" clearable @change="handleFilterChange">
              <el-option label="全部" value="" />
              <el-option label="启用" :value="true" />
              <el-option label="禁用" :value="false" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-input
              v-model="filters.search"
              placeholder="搜索仓库名或所有者"
              @input="handleSearchChange"
            >
              <template #prepend>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
        </el-row>
      </div>

      <el-table
        :data="repositories"
        v-loading="loading"
        class="repositories-table"
        @row-click="handleRowClick"
      >
        <el-table-column label="仓库信息" min-width="300">
          <template #default="scope">
            <div class="repo-info-cell">
              <div class="repo-name">
                <el-link :href="scope.row.repository_url" target="_blank" type="primary">
                  {{ scope.row.owner_name }}/{{ scope.row.repo_name }}
                </el-link>
                <el-tag size="small" class="branch-tag">{{ scope.row.branch_name }}</el-tag>
              </div>
              <div class="app-name">{{ scope.row.app_name }}</div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="同步状态" width="120">
          <template #default="scope">
            <div class="sync-status">
              <el-tag
                :type="getSyncStatusType(scope.row.last_sync_status)"
                size="small"
              >
                {{ getSyncStatusText(scope.row.last_sync_status) }}
              </el-tag>
              <div v-if="scope.row.last_sync_at" class="sync-time">
                {{ formatRelativeTime(scope.row.last_sync_at) }}
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="版本统计" width="100">
          <template #default="scope">
            <div class="version-stats">
              <el-tooltip content="已同步版本数" placement="top">
                <span class="stat-badge">{{ scope.row.releases_count || 0 }}</span>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="配置" width="150">
          <template #default="scope">
            <div class="config-badges">
              <el-tooltip content="自动同步" placement="top">
                <el-tag :type="scope.row.auto_sync ? 'success' : 'info'" size="small">
                  自动同步
                </el-tag>
              </el-tooltip>
              <el-tooltip content="自动发布" placement="top">
                <el-tag :type="scope.row.auto_publish ? 'warning' : 'info'" size="small">
                  自动发布
                </el-tag>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'danger'" size="small">
              {{ scope.row.is_active ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button 
              size="small" 
              @click.stop="syncRepository(scope.row)"
              :loading="scope.row.syncing"
              :disabled="!scope.row.is_active"
            >
              同步
            </el-button>
            <el-button size="small" @click.stop="viewReleases(scope.row)">
              版本
            </el-button>
            <el-dropdown @command="(cmd) => handleAction(cmd, scope.row)" trigger="click">
              <el-button size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="toggle">
                    {{ scope.row.is_active ? '禁用' : '启用' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handlePageSizeChange"
          @current-change="handleCurrentPageChange"
        />
      </div>
    </div>

    <!-- Create/Edit Repository Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingRepo ? '编辑仓库绑定' : '绑定GitHub仓库'"
      width="700px"
      @close="resetForm"
      destroy-on-close
    >
      <el-form
        ref="appFormRef"
        :model="appForm"
        :rules="appFormRules"
        label-width="100px"
      >
        <el-form-item label="应用" prop="app_id">
          <el-select 
            v-model="appForm.app_id" 
            placeholder="选择应用"
            style="width: 100%"
            :disabled="!!editingRepo"
          >
            <el-option
              v-for="app in applications"
              :key="app.app_id"
              :label="app.name"
              :value="app.app_id"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <GitHubRepositoryForm
        ref="githubFormRef"
        v-model="githubForm"
        :loading="submitting"
        @validation-change="handleValidationChange"
      />
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="submitForm" 
            :loading="submitting"
            :disabled="!isFormValid"
          >
            {{ editingRepo ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- Repository Releases Dialog -->
    <el-dialog
      v-model="showReleasesDialog"
      title="GitHub版本发布"
      width="900px"
      @close="resetReleasesDialog"
    >
      <div class="releases-header">
        <h3>{{ selectedRepo?.owner_name }}/{{ selectedRepo?.repo_name }} - 版本列表</h3>
        <el-button 
          type="primary" 
          size="small" 
          @click="syncRepository(selectedRepo)"
          :loading="selectedRepo?.syncing"
        >
          <el-icon><Refresh /></el-icon>
          同步版本
        </el-button>
      </div>

      <el-table :data="releases" v-loading="releasesLoading">
        <el-table-column label="版本" width="150">
          <template #default="scope">
            <div class="version-cell">
              <div class="version-name">{{ scope.row.tag_name }}</div>
              <el-tag 
                v-if="scope.row.is_prerelease" 
                type="warning" 
                size="small"
              >
                预发布
              </el-tag>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="release_name" label="发布名称" min-width="200" />
        
        <el-table-column label="同步状态" width="100">
          <template #default="scope">
            <el-tag 
              :type="getSyncStatusType(scope.row.sync_status)" 
              size="small"
            >
              {{ getSyncStatusText(scope.row.sync_status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="文件大小" width="100">
          <template #default="scope">
            {{ formatFileSize(scope.row.file_size) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="published_at" label="发布时间" width="150">
          <template #default="scope">
            {{ formatDate(scope.row.published_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button 
              size="small" 
              @click="downloadRelease(scope.row)"
              :disabled="scope.row.sync_status !== 'success'"
            >
              下载
            </el-button>
            <el-button 
              size="small" 
              type="primary"
              @click="viewReleaseDetail(scope.row)"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Refresh,
  Search,
  Collection,
  Check,
  Loading,
  DocumentCopy,
  ArrowDown
} from '@element-plus/icons-vue'
import {
  getAllGitHubRepositories,
  createGitHubRepository,
  updateGitHubRepository,
  deleteGitHubRepository,
  syncGitHubRepository,
  getGitHubReleases
} from '@/api/github'
import { getApplications } from '@/api/application'
import GitHubRepositoryForm from '@/components/GitHubRepositoryForm.vue'

// Reactive data
const loading = ref(false)
const repositories = ref([])
const applications = ref([])
const releases = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// Statistics
const stats = reactive({
  total_repositories: 0,
  active_repositories: 0,
  syncing_repositories: 0,
  total_releases: 0
})

// Filters
const filters = reactive({
  app_id: '',
  sync_status: '',
  is_active: '',
  search: ''
})

// Repository form
const showCreateDialog = ref(false)
const submitting = ref(false)
const editingRepo = ref(null)
const appFormRef = ref()
const githubFormRef = ref()
const isFormValid = ref(false)

const appForm = reactive({
  app_id: ''
})

const appFormRules = {
  app_id: [
    { required: true, message: '请选择应用', trigger: 'change' }
  ]
}

const githubForm = reactive({
  repository_url: '',
  owner_name: '',
  repo_name: '',
  branch_name: 'main',
  access_token: '',
  default_channel: 'stable',
  is_active: true,
  auto_sync: true,
  auto_publish: false
})

// Releases dialog
const showReleasesDialog = ref(false)
const releasesLoading = ref(false)
const selectedRepo = ref(null)

// Search debounce
let searchTimeout = null
let refreshInterval = null

// Methods
const fetchRepositories = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value,
      ...filters
    }
    
    // Remove empty filters
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    
    const response = await getAllGitHubRepositories(params)
    
    if (response.data) {
      repositories.value = response.data || []
      total.value = response.pagination?.total || 0
      
      // Update statistics
      updateStatistics(repositories.value)
    }
  } catch (error) {
    ElMessage.error('获取仓库列表失败')
    console.error('Failed to fetch repositories:', error)
  } finally {
    loading.value = false
  }
}

const fetchApplications = async () => {
  try {
    const response = await getApplications({ limit: 1000 })
    applications.value = response.data || []
  } catch (error) {
    console.error('Failed to fetch applications:', error)
  }
}

const updateStatistics = (repos) => {
  stats.total_repositories = repos.length
  stats.active_repositories = repos.filter(r => r.is_active).length
  stats.syncing_repositories = repos.filter(r => r.last_sync_status === 'syncing').length
  stats.total_releases = repos.reduce((sum, r) => sum + (r.releases_count || 0), 0)
}

const handleFilterChange = () => {
  currentPage.value = 1
  fetchRepositories()
}

const handleSearchChange = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchRepositories()
  }, 500)
}

const handlePageSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  fetchRepositories()
}

const handleCurrentPageChange = (page) => {
  currentPage.value = page
  fetchRepositories()
}

const handleRowClick = (row) => {
  viewReleases(row)
}

const refreshData = () => {
  fetchRepositories()
}

const syncRepository = async (repo) => {
  try {
    repo.syncing = true
    await syncGitHubRepository(repo.id, { force: true })
    ElMessage.success('同步请求已发送，请稍后查看结果')
    
    // Refresh data after 2 seconds
    setTimeout(() => {
      fetchRepositories()
    }, 2000)
  } catch (error) {
    ElMessage.error('同步失败')
    console.error('Failed to sync repository:', error)
  } finally {
    repo.syncing = false
  }
}

const viewReleases = async (repo) => {
  selectedRepo.value = repo
  showReleasesDialog.value = true
  await fetchReleases()
}

const fetchReleases = async () => {
  if (!selectedRepo.value) return
  
  releasesLoading.value = true
  try {
    const response = await getGitHubReleases(selectedRepo.value.id)
    releases.value = response.data || []
  } catch (error) {
    ElMessage.error('获取版本列表失败')
    console.error('Failed to fetch releases:', error)
  } finally {
    releasesLoading.value = false
  }
}

const handleAction = async (command, repo) => {
  switch (command) {
    case 'edit':
      editRepository(repo)
      break
    case 'toggle':
      await toggleRepository(repo)
      break
    case 'delete':
      await deleteRepository(repo)
      break
  }
}

const editRepository = (repo) => {
  editingRepo.value = repo
  appForm.app_id = repo.app_id
  Object.assign(githubForm, {
    repository_url: repo.repository_url,
    owner_name: repo.owner_name,
    repo_name: repo.repo_name,
    branch_name: repo.branch_name,
    access_token: '', // Don't pre-fill for security
    default_channel: repo.default_channel,
    is_active: repo.is_active,
    auto_sync: repo.auto_sync,
    auto_publish: repo.auto_publish
  })
  showCreateDialog.value = true
}

const toggleRepository = async (repo) => {
  try {
    await updateGitHubRepository(repo.id, {
      is_active: !repo.is_active
    })
    ElMessage.success(`仓库已${repo.is_active ? '禁用' : '启用'}`)
    fetchRepositories()
  } catch (error) {
    ElMessage.error('操作失败')
    console.error('Failed to toggle repository:', error)
  }
}

const deleteRepository = async (repo) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除仓库绑定 "${repo.owner_name}/${repo.repo_name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteGitHubRepository(repo.id)
    ElMessage.success('仓库绑定已删除')
    fetchRepositories()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('Failed to delete repository:', error)
    }
  }
}

const resetForm = () => {
  editingRepo.value = null
  appForm.app_id = ''
  Object.assign(githubForm, {
    repository_url: '',
    owner_name: '',
    repo_name: '',
    branch_name: 'main',
    access_token: '',
    default_channel: 'stable',
    is_active: true,
    auto_sync: true,
    auto_publish: false
  })
  isFormValid.value = false
  appFormRef.value?.clearValidate()
  githubFormRef.value?.clearValidate()
}

const resetReleasesDialog = () => {
  selectedRepo.value = null
  releases.value = []
}

const submitForm = async () => {
  // Validate forms
  const appValid = await appFormRef.value?.validate().catch(() => false)
  const githubValid = await githubFormRef.value?.validate().catch(() => false)
  
  if (!appValid || !githubValid) return
  
  submitting.value = true
  try {
    const formData = {
      ...githubForm,
      app_id: appForm.app_id
    }
    
    if (editingRepo.value) {
      await updateGitHubRepository(editingRepo.value.id, formData)
      ElMessage.success('仓库绑定更新成功')
    } else {
      await createGitHubRepository(formData)
      ElMessage.success('仓库绑定创建成功')
    }
    
    showCreateDialog.value = false
    fetchRepositories()
  } catch (error) {
    ElMessage.error(editingRepo.value ? '更新失败' : '创建失败')
    console.error('Failed to save repository:', error)
  } finally {
    submitting.value = false
  }
}

const handleValidationChange = (isValid) => {
  isFormValid.value = isValid
}

const downloadRelease = (release) => {
  if (release.download_url) {
    window.open(release.download_url, '_blank')
  } else {
    ElMessage.warning('文件尚未同步完成')
  }
}

const viewReleaseDetail = (release) => {
  // TODO: Implement release detail view
  ElMessage.info('功能开发中...')
}

// Utility functions
const getSyncStatusType = (status) => {
  const statusMap = {
    success: 'success',
    syncing: 'warning',
    failed: 'danger',
    pending: 'info'
  }
  return statusMap[status] || 'info'
}

const getSyncStatusText = (status) => {
  const statusMap = {
    success: '成功',
    syncing: '同步中',
    failed: '失败',
    pending: '等待'
  }
  return statusMap[status] || '未知'
}

const formatRelativeTime = (dateString) => {
  if (!dateString) return ''
  
  const date = new Date(dateString)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 0) return `${days}天前`
  if (hours > 0) return `${hours}小时前`
  if (minutes > 0) return `${minutes}分钟前`
  return '刚刚'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatFileSize = (bytes) => {
  if (!bytes) return ''
  
  const sizes = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < sizes.length - 1) {
    bytes /= 1024
    i++
  }
  
  return `${bytes.toFixed(1)} ${sizes[i]}`
}

// Computed properties
const canSubmit = computed(() => {
  return isFormValid.value && appForm.app_id
})

// Lifecycle
onMounted(() => {
  fetchApplications()
  fetchRepositories()
  
  // Auto refresh every 30 seconds for syncing repositories
  refreshInterval = setInterval(() => {
    if (repositories.value.some(r => r.last_sync_status === 'syncing')) {
      fetchRepositories()
    }
  }, 30000)
})

// Cleanup
watch(() => showReleasesDialog.value, (show) => {
  if (!show) {
    resetReleasesDialog()
  }
})

// Clean up intervals on component unmount
onBeforeUnmount(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
})
</script>

<style scoped>
.github-management-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  margin: 0 0 10px 0;
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.page-header p {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

/* Statistics Cards */
.stats-section {
  margin-bottom: 30px;
}

.stat-card {
  height: 100px;
  display: flex;
  align-items: center;
  position: relative;
  overflow: hidden;
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-top: 8px;
}

.stat-icon {
  font-size: 40px;
  color: #dcdfe6;
  position: absolute;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
}

.stat-icon.success {
  color: #67c23a;
}

.stat-icon.warning {
  color: #e6a23c;
}

.stat-icon.info {
  color: #409eff;
}

/* Content Section */
.content-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.filters-section {
  margin-bottom: 20px;
  padding: 16px;
  background: #fafbfc;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

/* Table Styles */
.repositories-table {
  width: 100%;
  margin-bottom: 20px;
}

.repo-info-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.repo-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.branch-tag {
  background: #f0f9ff;
  color: #0369a1;
  border: 1px solid #7dd3fc;
}

.app-name {
  font-size: 12px;
  color: #909399;
}

.sync-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.sync-time {
  font-size: 11px;
  color: #909399;
}

.version-stats {
  text-align: center;
}

.stat-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #e1f5fe;
  color: #0277bd;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.config-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

/* Dialog Styles */
.releases-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.releases-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.version-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version-name {
  font-weight: 600;
  color: #303133;
}

/* Responsive */
@media (max-width: 768px) {
  .stats-section :deep(.el-col) {
    margin-bottom: 20px;
  }
  
  .filters-section :deep(.el-col) {
    margin-bottom: 12px;
  }
  
  .header-actions {
    flex-direction: column;
    gap: 8px;
  }
}
</style>

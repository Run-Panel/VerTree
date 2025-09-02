<template>
  <div class="github-sync-status">
    <!-- Repository Sync Status -->
    <div v-if="repository" class="repository-status">
      <div class="status-header">
        <div class="repo-info">
          <el-icon class="repo-icon"><Collection /></el-icon>
          <span class="repo-name">{{ repository.owner_name }}/{{ repository.repo_name }}</span>
          <el-tag :type="getRepositoryStatusType(repository)" size="small">
            {{ getRepositoryStatusText(repository) }}
          </el-tag>
        </div>
        <div class="status-actions">
          <el-button 
            size="small" 
            :loading="syncing" 
            @click="syncRepository"
            :disabled="!repository.is_active"
          >
            <el-icon><Refresh /></el-icon>
            同步
          </el-button>
        </div>
      </div>

      <div v-if="repository.last_sync_at" class="sync-info">
        <el-row :gutter="16">
          <el-col :span="8">
            <div class="info-item">
              <span class="label">最后同步：</span>
              <span class="value">{{ formatRelativeTime(repository.last_sync_at) }}</span>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="info-item">
              <span class="label">同步版本：</span>
              <span class="value">{{ repository.releases_count || 0 }} 个</span>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="info-item">
              <span class="label">同步状态：</span>
              <el-tag :type="getSyncStatusType(repository.last_sync_status)" size="small">
                {{ getSyncStatusText(repository.last_sync_status) }}
              </el-tag>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- Sync Progress -->
      <div v-if="showProgress && syncProgress" class="sync-progress">
        <div class="progress-header">
          <span class="progress-title">{{ syncProgress.title }}</span>
          <span class="progress-percentage">{{ syncProgress.percentage }}%</span>
        </div>
        <el-progress 
          :percentage="syncProgress.percentage" 
          :status="syncProgress.status"
          :show-text="false"
        />
        <div v-if="syncProgress.message" class="progress-message">
          {{ syncProgress.message }}
        </div>
      </div>

      <!-- Sync Logs -->
      <div v-if="showLogs && syncLogs.length > 0" class="sync-logs">
        <div class="logs-header">
          <span class="logs-title">同步日志</span>
          <el-button size="small" text @click="clearLogs">
            <el-icon><Delete /></el-icon>
            清空
          </el-button>
        </div>
        <div class="logs-content">
          <div 
            v-for="(log, index) in syncLogs" 
            :key="index"
            class="log-item"
            :class="log.level"
          >
            <span class="log-time">{{ formatTime(log.timestamp) }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Batch Sync Status -->
    <div v-if="batchSync" class="batch-sync-status">
      <div class="batch-header">
        <h4>批量同步状态</h4>
        <el-tag :type="getBatchStatusType(batchSync.status)" size="small">
          {{ getBatchStatusText(batchSync.status) }}
        </el-tag>
      </div>

      <div class="batch-progress">
        <div class="progress-summary">
          <span>进度：{{ batchSync.completed }}/{{ batchSync.total }}</span>
          <span>{{ Math.round((batchSync.completed / batchSync.total) * 100) }}%</span>
        </div>
        <el-progress 
          :percentage="Math.round((batchSync.completed / batchSync.total) * 100)"
          :status="batchSync.status === 'failed' ? 'exception' : 'success'"
        />
      </div>

      <div v-if="batchSync.repositories" class="batch-repositories">
        <div 
          v-for="repo in batchSync.repositories" 
          :key="repo.id"
          class="batch-repo-item"
        >
          <div class="repo-info">
            <span class="repo-name">{{ repo.owner_name }}/{{ repo.repo_name }}</span>
            <el-tag :type="getSyncStatusType(repo.status)" size="small">
              {{ getSyncStatusText(repo.status) }}
            </el-tag>
          </div>
          <div v-if="repo.error" class="repo-error">
            {{ repo.error }}
          </div>
        </div>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="error-display">
      <el-alert
        :title="error.title || '同步错误'"
        :description="error.message"
        type="error"
        show-icon
        :closable="true"
        @close="clearError"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Collection, Refresh, Delete } from '@element-plus/icons-vue'
import { syncGitHubRepository } from '@/api/github'

// Props
const props = defineProps({
  repository: {
    type: Object,
    default: null
  },
  batchSync: {
    type: Object,
    default: null
  },
  showProgress: {
    type: Boolean,
    default: true
  },
  showLogs: {
    type: Boolean,
    default: false
  },
  autoRefresh: {
    type: Boolean,
    default: true
  },
  refreshInterval: {
    type: Number,
    default: 5000 // 5 seconds
  }
})

// Emits
const emit = defineEmits(['sync-completed', 'sync-failed', 'refresh-required'])

// Reactive data
const syncing = ref(false)
const syncProgress = ref(null)
const syncLogs = ref([])
const error = ref(null)

// Auto refresh timer
let refreshTimer = null

// Methods
const syncRepository = async () => {
  if (!props.repository || syncing.value) return
  
  syncing.value = true
  error.value = null
  
  try {
    addLog('info', '开始同步仓库...')
    
    await syncGitHubRepository(props.repository.id, { force: true })
    
    addLog('success', '同步请求已发送')
    ElMessage.success('同步请求已发送，请稍后查看结果')
    
    emit('sync-completed', props.repository)
    
    // Start monitoring sync progress
    if (props.showProgress) {
      startProgressMonitoring()
    }
    
  } catch (error) {
    const errorMessage = error.response?.data?.message || '同步失败'
    addLog('error', errorMessage)
    handleError('同步失败', errorMessage)
    emit('sync-failed', props.repository, error)
  } finally {
    syncing.value = false
  }
}

const startProgressMonitoring = () => {
  syncProgress.value = {
    title: '正在同步版本...',
    percentage: 0,
    status: 'active',
    message: '正在获取GitHub Release列表'
  }
  
  // Simulate progress updates
  let progress = 0
  const progressTimer = setInterval(() => {
    progress += 10
    syncProgress.value.percentage = Math.min(progress, 90)
    
    if (progress === 20) {
      syncProgress.value.message = '正在下载Release文件'
    } else if (progress === 50) {
      syncProgress.value.message = '正在创建版本记录'
    } else if (progress === 80) {
      syncProgress.value.message = '正在完成同步'
    }
    
    if (progress >= 90) {
      clearInterval(progressTimer)
      // Check actual status after a delay
      setTimeout(checkSyncStatus, 2000)
    }
  }, 1000)
}

const checkSyncStatus = () => {
  // This should fetch the actual sync status from the API
  // For now, simulate completion
  if (syncProgress.value) {
    syncProgress.value.percentage = 100
    syncProgress.value.status = 'success'
    syncProgress.value.message = '同步完成'
    addLog('success', '仓库同步完成')
    
    setTimeout(() => {
      syncProgress.value = null
      emit('refresh-required')
    }, 1000)
  }
}

const addLog = (level, message) => {
  if (!props.showLogs) return
  
  syncLogs.value.unshift({
    timestamp: new Date(),
    level,
    message
  })
  
  // Keep only last 50 logs
  if (syncLogs.value.length > 50) {
    syncLogs.value = syncLogs.value.slice(0, 50)
  }
}

const clearLogs = () => {
  syncLogs.value = []
}

const handleError = (title, message) => {
  error.value = { title, message }
}

const clearError = () => {
  error.value = null
}

// Utility functions
const getRepositoryStatusType = (repo) => {
  if (!repo.is_active) return 'info'
  return getSyncStatusType(repo.last_sync_status)
}

const getRepositoryStatusText = (repo) => {
  if (!repo.is_active) return '已禁用'
  return getSyncStatusText(repo.last_sync_status)
}

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
    success: '同步成功',
    syncing: '同步中',
    failed: '同步失败',
    pending: '等待同步'
  }
  return statusMap[status] || '未知'
}

const getBatchStatusType = (status) => {
  const statusMap = {
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return statusMap[status] || 'info'
}

const getBatchStatusText = (status) => {
  const statusMap = {
    running: '进行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
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

const formatTime = (date) => {
  return date.toLocaleTimeString('zh-CN')
}

// Auto refresh
const startAutoRefresh = () => {
  if (!props.autoRefresh || refreshTimer) return
  
  refreshTimer = setInterval(() => {
    if (props.repository?.last_sync_status === 'syncing') {
      emit('refresh-required')
    }
  }, props.refreshInterval)
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// Lifecycle
onMounted(() => {
  if (props.autoRefresh) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
})

// Watch for prop changes
watch(() => props.autoRefresh, (newValue) => {
  if (newValue) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})

watch(() => props.repository?.last_sync_status, (newStatus) => {
  if (newStatus === 'syncing' && props.showProgress && !syncProgress.value) {
    startProgressMonitoring()
  } else if (newStatus !== 'syncing' && syncProgress.value) {
    // Update progress based on actual status
    if (newStatus === 'success') {
      syncProgress.value.percentage = 100
      syncProgress.value.status = 'success'
      syncProgress.value.message = '同步成功'
      setTimeout(() => {
        syncProgress.value = null
      }, 2000)
    } else if (newStatus === 'failed') {
      syncProgress.value.status = 'exception'
      syncProgress.value.message = '同步失败'
      setTimeout(() => {
        syncProgress.value = null
      }, 3000)
    }
  }
})
</script>

<style scoped>
.github-sync-status {
  width: 100%;
}

/* Repository Status */
.repository-status {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background: #fafbfc;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.repo-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.repo-icon {
  color: #606266;
}

.repo-name {
  font-weight: 600;
  color: #303133;
}

.sync-info {
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item .label {
  font-size: 12px;
  color: #909399;
}

.info-item .value {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

/* Sync Progress */
.sync-progress {
  margin: 16px 0;
  padding: 16px;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e1f5fe;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.progress-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.progress-percentage {
  font-size: 12px;
  color: #606266;
}

.progress-message {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

/* Sync Logs */
.sync-logs {
  margin-top: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #f5f7fa;
}

.logs-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.logs-content {
  max-height: 200px;
  overflow-y: auto;
  padding: 8px;
}

.log-item {
  display: flex;
  gap: 8px;
  padding: 4px 8px;
  margin-bottom: 2px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.4;
}

.log-item.info {
  background: #f0f9ff;
  color: #0369a1;
}

.log-item.success {
  background: #f0fdf4;
  color: #15803d;
}

.log-item.error {
  background: #fef2f2;
  color: #dc2626;
}

.log-time {
  flex-shrink: 0;
  font-family: monospace;
}

.log-message {
  flex: 1;
}

/* Batch Sync Status */
.batch-sync-status {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  background: #fff;
}

.batch-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.batch-header h4 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.batch-progress {
  margin-bottom: 16px;
}

.progress-summary {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
}

.batch-repositories {
  border-top: 1px solid #e4e7ed;
  padding-top: 16px;
}

.batch-repo-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px;
  margin-bottom: 8px;
  background: #f5f7fa;
  border-radius: 4px;
}

.batch-repo-item .repo-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.batch-repo-item .repo-name {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
}

.repo-error {
  font-size: 12px;
  color: #f56565;
  padding: 4px 8px;
  background: #fed7d7;
  border-radius: 4px;
}

/* Error Display */
.error-display {
  margin: 16px 0;
}

/* Responsive */
@media (max-width: 768px) {
  .status-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .batch-repo-item .repo-info {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>


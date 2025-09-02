<template>
  <el-card class="github-version-card" :class="{ 'syncing': isSyncing }">
    <div class="version-header">
      <div class="version-info">
        <div class="version-title">
          <el-icon class="github-icon"><DataBoard /></el-icon>
          <span class="version-name">{{ version.tag_name }}</span>
          <el-tag v-if="version.is_prerelease" type="warning" size="small">
            预发布
          </el-tag>
          <el-tag v-if="version.is_draft" type="info" size="small">
            草稿
          </el-tag>
        </div>
        <div class="version-subtitle">
          <span class="release-name">{{ version.release_name }}</span>
          <span class="publish-time">{{ formatDate(version.published_at) }}</span>
        </div>
      </div>
      
      <div class="version-actions">
        <el-tooltip content="同步状态" placement="top">
          <el-tag :type="getSyncStatusType(version.sync_status)" size="small">
            {{ getSyncStatusText(version.sync_status) }}
          </el-tag>
        </el-tooltip>
        
        <el-dropdown @command="handleAction" trigger="click">
          <el-button size="small" circle>
            <el-icon><MoreFilled /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item 
                command="download" 
                :disabled="version.sync_status !== 'success'"
              >
                <el-icon><Download /></el-icon>
                下载文件
              </el-dropdown-item>
              <el-dropdown-item command="sync">
                <el-icon><Refresh /></el-icon>
                重新同步
              </el-dropdown-item>
              <el-dropdown-item command="view-details">
                <el-icon><View /></el-icon>
                查看详情
              </el-dropdown-item>
              <el-dropdown-item 
                command="create-version" 
                :disabled="version.sync_status !== 'success'"
                divided
              >
                <el-icon><Plus /></el-icon>
                创建版本记录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- Version Description -->
    <div v-if="version.body" class="version-description">
      <div class="description-content" :class="{ 'expanded': showFullDescription }">
        <div v-html="formattedDescription"></div>
      </div>
      <el-button 
        v-if="version.body.length > 200"
        size="small" 
        text 
        @click="showFullDescription = !showFullDescription"
        class="expand-button"
      >
        {{ showFullDescription ? '收起' : '展开' }}
        <el-icon>
          <ArrowDown v-if="!showFullDescription" />
          <ArrowUp v-else />
        </el-icon>
      </el-button>
    </div>

    <!-- Version Statistics -->
    <div class="version-stats">
      <el-row :gutter="16">
        <el-col :span="8">
          <div class="stat-item">
            <span class="stat-label">文件大小</span>
            <span class="stat-value">{{ formatFileSize(version.file_size) }}</span>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-item">
            <span class="stat-label">下载次数</span>
            <span class="stat-value">{{ version.download_count || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-item">
            <span class="stat-label">同步时间</span>
            <span class="stat-value">{{ formatRelativeTime(version.synced_at) }}</span>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- Sync Progress -->
    <div v-if="isSyncing" class="sync-progress">
      <div class="progress-header">
        <span class="progress-title">正在同步版本...</span>
        <span class="progress-percentage">{{ syncProgress }}%</span>
      </div>
      <el-progress 
        :percentage="syncProgress" 
        :status="syncStatus"
        :show-text="false"
        :stroke-width="6"
      />
      <div v-if="syncMessage" class="progress-message">
        {{ syncMessage }}
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="syncError" class="sync-error">
      <el-alert
        :title="syncError.title || '同步错误'"
        :description="syncError.message"
        type="error"
        size="small"
        show-icon
        :closable="true"
        @close="clearSyncError"
      />
    </div>

    <!-- Repository Info -->
    <div class="repository-info">
      <div class="repo-link">
        <el-icon><Link /></el-icon>
        <el-link 
          :href="`${repository.repository_url}/releases/tag/${version.tag_name}`" 
          target="_blank"
          type="primary"
          size="small"
        >
          在GitHub上查看
        </el-link>
      </div>
      <div class="repo-channel">
        <el-icon><Switch /></el-icon>
        <span class="channel-text">默认渠道：{{ repository.default_channel }}</span>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DataBoard,
  MoreFilled,
  Download,
  Refresh,
  View,
  Plus,
  Link,
  Switch,
  ArrowDown,
  ArrowUp
} from '@element-plus/icons-vue'
import { marked } from 'marked'

// Props
const props = defineProps({
  version: {
    type: Object,
    required: true
  },
  repository: {
    type: Object,
    required: true
  }
})

// Emits
const emit = defineEmits([
  'sync-version',
  'download-version', 
  'view-details',
  'create-version'
])

// Reactive data
const showFullDescription = ref(false)
const isSyncing = ref(false)
const syncProgress = ref(0)
const syncStatus = ref('active')
const syncMessage = ref('')
const syncError = ref(null)

// Computed properties
const formattedDescription = computed(() => {
  if (!props.version.body) return ''
  
  let description = props.version.body
  
  // Limit description length if not expanded
  if (!showFullDescription.value && description.length > 200) {
    description = description.substring(0, 200) + '...'
  }
  
  // Convert markdown to HTML (basic support)
  try {
    return marked(description, { 
      breaks: true,
      sanitize: true,
      headerIds: false
    })
  } catch (error) {
    console.warn('Failed to parse markdown:', error)
    return description.replace(/\n/g, '<br>')
  }
})

// Methods
const handleAction = async (command) => {
  switch (command) {
    case 'download':
      handleDownload()
      break
    case 'sync':
      await handleSync()
      break
    case 'view-details':
      handleViewDetails()
      break
    case 'create-version':
      handleCreateVersion()
      break
  }
}

const handleDownload = () => {
  if (props.version.download_url) {
    emit('download-version', props.version)
    window.open(props.version.download_url, '_blank')
  } else {
    ElMessage.warning('文件尚未同步完成')
  }
}

const handleSync = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要重新同步版本 "${props.version.tag_name}" 吗？`,
      '确认同步',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    startSync()
    emit('sync-version', props.version)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Sync error:', error)
    }
  }
}

const handleViewDetails = () => {
  emit('view-details', props.version)
}

const handleCreateVersion = () => {
  emit('create-version', props.version)
}

const startSync = () => {
  isSyncing.value = true
  syncProgress.value = 0
  syncStatus.value = 'active'
  syncMessage.value = '正在初始化同步...'
  syncError.value = null
  
  // Simulate sync progress
  const progressTimer = setInterval(() => {
    syncProgress.value += 10
    
    if (syncProgress.value === 20) {
      syncMessage.value = '正在下载文件...'
    } else if (syncProgress.value === 50) {
      syncMessage.value = '正在验证文件完整性...'
    } else if (syncProgress.value === 80) {
      syncMessage.value = '正在创建版本记录...'
    } else if (syncProgress.value >= 100) {
      clearInterval(progressTimer)
      finishSync()
    }
  }, 500)
}

const finishSync = () => {
  syncProgress.value = 100
  syncStatus.value = 'success'
  syncMessage.value = '同步完成'
  
  setTimeout(() => {
    isSyncing.value = false
    syncProgress.value = 0
    syncMessage.value = ''
  }, 1500)
  
  ElMessage.success('版本同步成功')
}

const clearSyncError = () => {
  syncError.value = null
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
    success: '已同步',
    syncing: '同步中',
    failed: '同步失败',
    pending: '等待同步'
  }
  return statusMap[status] || '未知'
}

const formatFileSize = (bytes) => {
  if (!bytes) return '未知'
  
  const sizes = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < sizes.length - 1) {
    bytes /= 1024
    i++
  }
  
  return `${bytes.toFixed(1)} ${sizes[i]}`
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatRelativeTime = (dateString) => {
  if (!dateString) return '未同步'
  
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

// Watch for external sync status changes
watch(() => props.version.sync_status, (newStatus) => {
  if (newStatus === 'syncing' && !isSyncing.value) {
    startSync()
  } else if (newStatus === 'failed' && isSyncing.value) {
    syncStatus.value = 'exception'
    syncMessage.value = '同步失败'
    syncError.value = {
      title: '版本同步失败',
      message: '请检查网络连接或稍后重试'
    }
    
    setTimeout(() => {
      isSyncing.value = false
    }, 2000)
  } else if (newStatus === 'success' && isSyncing.value) {
    finishSync()
  }
})
</script>

<style scoped>
.github-version-card {
  margin-bottom: 16px;
  transition: all 0.3s ease;
  border: 1px solid #e4e7ed;
}

.github-version-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.github-version-card.syncing {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

/* Version Header */
.version-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.version-info {
  flex: 1;
}

.version-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.github-icon {
  color: #606266;
  font-size: 16px;
}

.version-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.version-subtitle {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.release-name {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.publish-time {
  font-size: 12px;
  color: #909399;
}

.version-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Version Description */
.version-description {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.description-content {
  max-height: 80px;
  overflow: hidden;
  transition: max-height 0.3s ease;
}

.description-content.expanded {
  max-height: none;
}

.description-content :deep(h1),
.description-content :deep(h2),
.description-content :deep(h3),
.description-content :deep(h4),
.description-content :deep(h5),
.description-content :deep(h6) {
  margin: 8px 0 4px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.description-content :deep(p) {
  margin: 4px 0;
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}

.description-content :deep(ul),
.description-content :deep(ol) {
  margin: 4px 0;
  padding-left: 20px;
}

.description-content :deep(li) {
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}

.description-content :deep(code) {
  background: #f0f2f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
  font-size: 12px;
}

.expand-button {
  margin-top: 8px;
  font-size: 12px;
}

/* Version Statistics */
.version-stats {
  margin-bottom: 16px;
  padding: 12px;
  background: #fafbfc;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

/* Sync Progress */
.sync-progress {
  margin-bottom: 16px;
  padding: 12px;
  background: #e3f2fd;
  border-radius: 6px;
  border: 1px solid #bbdefb;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.progress-title {
  font-size: 13px;
  font-weight: 500;
  color: #1976d2;
}

.progress-percentage {
  font-size: 12px;
  color: #1976d2;
}

.progress-message {
  margin-top: 8px;
  font-size: 12px;
  color: #1565c0;
}

/* Sync Error */
.sync-error {
  margin-bottom: 16px;
}

/* Repository Info */
.repository-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #909399;
}

.repo-link,
.repo-channel {
  display: flex;
  align-items: center;
  gap: 4px;
}

.channel-text {
  color: #606266;
}

/* Responsive */
@media (max-width: 768px) {
  .version-header {
    flex-direction: column;
    gap: 12px;
  }
  
  .version-actions {
    align-self: stretch;
    justify-content: flex-end;
  }
  
  .repository-info {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
  
  .version-stats :deep(.el-col) {
    margin-bottom: 8px;
  }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .github-version-card {
    background: #1a1a1a;
    border-color: #404040;
  }
  
  .version-description {
    background: #2a2a2a;
  }
  
  .version-stats {
    background: #252525;
    border-color: #404040;
  }
}
</style>


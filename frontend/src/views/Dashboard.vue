<template>
  <div class="dashboard">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">{{ $t('dashboard.welcome') }}</h1>
        <p class="page-subtitle">{{ $t('dashboard.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="refreshData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          <span class="mobile-hidden">{{ $t('common.refresh') }}</span>
        </el-button>
      </div>
    </div>

    <!-- Statistics Overview -->
    <div class="stats-section">
      <h2 class="section-title">{{ $t('dashboard.overview') }}</h2>
      <el-row :gutter="24" class="stats-grid">
        <el-col :xs="12" :sm="12" :md="6" :lg="6">
          <div class="stats-card" @click="navigateToUsers">
            <div class="icon gradient-primary">
              <el-icon size="24"><User /></el-icon>
            </div>
            <div class="content">
              <div class="value">{{ formatNumber(stats.totalUsers || 0) }}</div>
              <div class="label">{{ $t('stats.totalUsers') }}</div>
            </div>
            <div class="trend">
              <el-icon color="#52c41a"><TrendCharts /></el-icon>
              <span class="trend-text">+12%</span>
            </div>
          </div>
        </el-col>
        
        <el-col :xs="12" :sm="12" :md="6" :lg="6">
          <div class="stats-card" @click="navigateToDownloads">
            <div class="icon gradient-success">
              <el-icon size="24"><Download /></el-icon>
            </div>
            <div class="content">
              <div class="value">{{ formatNumber(stats.totalDownloads || 0) }}</div>
              <div class="label">{{ $t('stats.totalDownloads') }}</div>
            </div>
            <div class="trend">
              <el-icon color="#52c41a"><TrendCharts /></el-icon>
              <span class="trend-text">+8%</span>
            </div>
          </div>
        </el-col>
        
        <el-col :xs="12" :sm="12" :md="6" :lg="6">
          <div class="stats-card" @click="navigateToStatistics">
            <div class="icon gradient-warning">
              <el-icon size="24"><TrendCharts /></el-icon>
            </div>
            <div class="content">
              <div class="value">{{ (stats.successRate || 0).toFixed(1) }}%</div>
              <div class="label">{{ $t('stats.successRate') }}</div>
            </div>
            <div class="trend">
              <el-icon color="#52c41a"><TrendCharts /></el-icon>
              <span class="trend-text">+2.3%</span>
            </div>
          </div>
        </el-col>
        
        <el-col :xs="12" :sm="12" :md="6" :lg="6">
          <div class="stats-card" @click="navigateToVersions">
            <div class="icon gradient-secondary">
              <el-icon size="24"><Box /></el-icon>
            </div>
            <div class="content">
              <div class="value">{{ formatNumber(versions.length) }}</div>
              <div class="label">{{ $t('stats.totalVersions') }}</div>
            </div>
            <div class="trend">
              <el-icon color="#52c41a"><TrendCharts /></el-icon>
              <span class="trend-text">+5</span>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- Content Grid -->
    <el-row :gutter="24" class="content-grid">
      <!-- Latest Versions Status -->
      <el-col :xs="24" :lg="16">
        <div class="modern-card">
          <div class="modern-card-header">
            <div class="card-title">
              <el-icon><Box /></el-icon>
              <span>{{ $t('dashboard.latestVersions') }}</span>
            </div>
            <div class="card-actions">
              <el-button text @click="navigateToVersions">
                {{ $t('common.viewAll') || 'View All' }}
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
          
          <div class="modern-card-body">
            <el-table 
              :data="latestVersions" 
              v-loading="loading"
              class="modern-table"
            >
              <el-table-column prop="channel" :label="$t('version.channel')" width="120">
                <template #default="scope">
                  <el-tag 
                    :type="getChannelTagType(scope.row.channel)"
                    effect="light"
                    round
                  >
                    {{ scope.row.channel }}
                  </el-tag>
                </template>
              </el-table-column>
              
              <el-table-column prop="version" :label="$t('version.versionNumber')" width="120">
                <template #default="scope">
                  <span class="version-text">{{ scope.row.version }}</span>
                </template>
              </el-table-column>
              
              <el-table-column prop="title" :label="$t('version.versionTitle')" show-overflow-tooltip />
              
              <el-table-column prop="publish_time" :label="$t('dashboard.publishTime')" width="160" class-name="mobile-hidden">
                <template #default="scope">
                  <div class="time-container">
                    <el-icon size="14"><Clock /></el-icon>
                    <span>{{ formatDate(scope.row.publish_time) }}</span>
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column :label="$t('dashboard.adoptionRate')" width="140" class-name="mobile-hidden">
                <template #default="scope">
                  <div class="adoption-container">
                    <el-progress 
                      :percentage="getAdoptionRate(scope.row.version)" 
                      :color="getProgressColor(getAdoptionRate(scope.row.version))"
                      :stroke-width="8"
                      :show-text="false"
                    />
                    <span class="adoption-text">{{ getAdoptionRate(scope.row.version) }}%</span>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-col>

      <!-- Quick Actions & System Status -->
      <el-col :xs="24" :lg="8">
        <!-- Quick Actions -->
        <div class="modern-card quick-actions-card">
          <div class="modern-card-header">
            <div class="card-title">
              <el-icon><Lightning /></el-icon>
              <span>{{ $t('dashboard.quickActions') }}</span>
            </div>
          </div>
          
          <div class="modern-card-body">
            <div class="action-buttons">
              <el-button 
                type="primary" 
                class="action-btn"
                @click="navigateToVersionCreate"
              >
                <el-icon><Plus /></el-icon>
                {{ $t('version.createVersion') }}
              </el-button>
              
              <el-button 
                class="action-btn"
                @click="navigateToChannels"
              >
                <el-icon><Setting /></el-icon>
                {{ $t('nav.channel') }}
              </el-button>
              
              <el-button 
                class="action-btn"
                @click="navigateToStatistics"
              >
                <el-icon><PieChart /></el-icon>
                {{ $t('nav.statistics') }}
              </el-button>
            </div>
          </div>
        </div>

        <!-- System Status -->
        <div class="modern-card system-status-card">
          <div class="modern-card-header">
            <div class="card-title">
              <el-icon><Monitor /></el-icon>
              <span>{{ $t('dashboard.systemStatus') }}</span>
            </div>
          </div>
          
          <div class="modern-card-body">
            <div class="status-items">
              <div class="status-item">
                <div class="status-indicator online"></div>
                <div class="status-info">
                  <span class="status-label">{{ $t('common.apiStatus') || 'API Status' }}</span>
                  <span class="status-value">{{ $t('common.online') || 'Online' }}</span>
                </div>
              </div>
              
              <div class="status-item">
                <div class="status-indicator online"></div>
                <div class="status-info">
                  <span class="status-label">{{ $t('common.database') || 'Database' }}</span>
                  <span class="status-value">{{ $t('common.online') || 'Online' }}</span>
                </div>
              </div>
              
              <div class="status-item">
                <div class="status-indicator warning"></div>
                <div class="status-info">
                  <span class="status-label">{{ $t('common.storage') || 'Storage' }}</span>
                  <span class="status-value">{{ $t('common.warning') || 'Warning' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getVersions } from '@/api/version'
import { getStats } from '@/api/stats'

const router = useRouter()
const loading = ref(false)
const stats = ref({})
const versions = ref([])
const latestVersions = ref([])

// Data fetching functions
const fetchStats = async () => {
  try {
    const response = await getStats({ period: '7d', action: 'all' })
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

const fetchVersions = async () => {
  try {
    const response = await getVersions({ limit: 100 })
    versions.value = response.data || []
    
    // Get latest published version for each channel
    const channelMap = new Map()
    versions.value
      .filter(v => v.is_published)
      .forEach(version => {
        if (!channelMap.has(version.channel) || 
            new Date(version.publish_time) > new Date(channelMap.get(version.channel).publish_time)) {
          channelMap.set(version.channel, version)
        }
      })
    
    latestVersions.value = Array.from(channelMap.values()).slice(0, 5) // Limit to 5 latest
  } catch (error) {
    console.error('Failed to fetch versions:', error)
  }
}

const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchStats(), fetchVersions()])
  } finally {
    loading.value = false
  }
}

// Navigation functions
const navigateToUsers = () => {
  // Future implementation for user management
  console.log('Navigate to users')
}

const navigateToDownloads = () => {
  router.push('/statistics')
}

const navigateToVersions = () => {
  router.push('/versions')
}

const navigateToChannels = () => {
  router.push('/channels')
}

const navigateToStatistics = () => {
  router.push('/statistics')
}

const navigateToVersionCreate = () => {
  router.push('/versions?action=create')
}

// Utility functions
const formatNumber = (num) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  const now = new Date()
  const diffTime = Math.abs(now - date)
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays === 1) return 'Today'
  if (diffDays === 2) return 'Yesterday'
  if (diffDays <= 7) return `${diffDays} days ago`
  
  return date.toLocaleDateString()
}

const getChannelTagType = (channel) => {
  const typeMap = {
    'stable': 'success',
    'beta': 'warning',
    'alpha': 'info'
  }
  return typeMap[channel] || 'info'
}

const getAdoptionRate = (version) => {
  // Mock adoption rate calculation based on version and stats
  const versionStats = stats.value.versionDistribution || {}
  const total = Object.values(versionStats).reduce((sum, count) => sum + count, 0)
  const versionCount = versionStats[version] || 0
  return total > 0 ? Math.round((versionCount / total) * 100) : Math.floor(Math.random() * 100)
}

const getProgressColor = (percentage) => {
  if (percentage > 70) return '#52c41a'
  if (percentage > 40) return '#faad14'
  return '#ff4d4f'
}

onMounted(() => {
  refreshData()
})
</script>

<style scoped>
/* Dashboard Container */
.dashboard {
  padding: 0;
  min-height: 100%;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-3xl);
  padding: var(--spacing-2xl);
  background: linear-gradient(135deg, var(--primary-ultralight), var(--bg-primary));
  border-radius: var(--border-radius-2xl);
  border: 1px solid var(--border-light);
}

.header-content h1 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-content p {
  margin: 0;
  font-size: var(--font-size-lg);
  color: var(--text-secondary);
}

.header-actions {
  display: flex;
  gap: var(--spacing-md);
}

/* Statistics Section */
.stats-section {
  margin-bottom: var(--spacing-3xl);
}

.section-title {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-xl);
  padding-left: var(--spacing-sm);
}

.stats-grid {
  margin-bottom: var(--spacing-2xl);
}

/* Modern Stats Cards */
.stats-card {
  background: var(--bg-primary);
  padding: var(--spacing-2xl);
  border-radius: var(--border-radius-2xl);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  cursor: pointer;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  min-height: 140px;
}

.stats-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--primary-color), var(--secondary-color));
}

.stats-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.stats-card .icon {
  width: 60px;
  height: 60px;
  border-radius: var(--border-radius-xl);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-inverse);
  font-size: 24px;
  align-self: flex-start;
}

.stats-card .content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.stats-card .value {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: var(--spacing-xs);
}

.stats-card .label {
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stats-card .trend {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  align-self: flex-end;
  position: absolute;
  top: var(--spacing-lg);
  right: var(--spacing-lg);
}

.trend-text {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--success-color);
}

/* Content Grid */
.content-grid {
  gap: var(--spacing-2xl);
}

/* Modern Card Styles */
.modern-card {
  background: var(--bg-primary);
  border-radius: var(--border-radius-2xl);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  overflow: hidden;
  margin-bottom: var(--spacing-2xl);
  transition: all var(--transition-normal);
}

.modern-card:hover {
  box-shadow: var(--shadow-md);
}

.modern-card-header {
  padding: var(--spacing-xl) var(--spacing-2xl);
  background: linear-gradient(135deg, var(--bg-primary), var(--bg-secondary));
  border-bottom: 1px solid var(--border-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.card-title .el-icon {
  color: var(--primary-color);
  font-size: 20px;
}

.card-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.modern-card-body {
  padding: var(--spacing-2xl);
}

/* Table Enhancements */
.modern-table {
  border-radius: var(--border-radius-lg);
  overflow: hidden;
}

.version-text {
  font-family: var(--font-family-mono);
  font-weight: var(--font-weight-semibold);
  background: var(--bg-tertiary);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius-sm);
  font-size: var(--font-size-xs);
}

.time-container {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.adoption-container {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.adoption-text {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
  min-width: 32px;
}

/* Quick Actions Card */
.quick-actions-card {
  margin-bottom: var(--spacing-xl);
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.action-btn {
  width: 100%;
  justify-content: flex-start;
  height: 48px;
  border-radius: var(--border-radius-lg) !important;
  font-weight: var(--font-weight-medium) !important;
  transition: all var(--transition-normal) !important;
}

.action-btn .el-icon {
  margin-right: var(--spacing-sm);
}

/* System Status Card */
.system-status-card {
  margin-bottom: 0;
}

.status-items {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.status-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--bg-secondary);
  border-radius: var(--border-radius-lg);
  transition: background var(--transition-normal);
}

.status-item:hover {
  background: var(--bg-tertiary);
}

.status-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-indicator.online {
  background: var(--success-color);
  box-shadow: 0 0 8px rgba(82, 196, 26, 0.4);
}

.status-indicator.warning {
  background: var(--warning-color);
  box-shadow: 0 0 8px rgba(250, 173, 20, 0.4);
}

.status-indicator.error {
  background: var(--error-color);
  box-shadow: 0 0 8px rgba(255, 77, 79, 0.4);
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.status-label {
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
}

.status-value {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

/* Responsive Design */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: var(--spacing-lg);
    align-items: stretch;
    padding: var(--spacing-lg);
  }
  
  .header-content h1 {
    font-size: var(--font-size-3xl);
  }
  
  .header-content p {
    font-size: var(--font-size-md);
  }
  
  .stats-card {
    padding: var(--spacing-lg);
    min-height: 120px;
  }
  
  .stats-card .value {
    font-size: var(--font-size-3xl);
  }
  
  .stats-card .icon {
    width: 48px;
    height: 48px;
  }
  
  .trend {
    position: static !important;
    align-self: flex-start !important;
    margin-top: var(--spacing-sm);
  }
  
  .modern-card-header {
    padding: var(--spacing-lg);
  }
  
  .modern-card-body {
    padding: var(--spacing-lg);
  }
  
  .card-title {
    font-size: var(--font-size-md);
  }
  
  .action-btn {
    height: 44px;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    gap: var(--spacing-md) !important;
  }
  
  .stats-card {
    padding: var(--spacing-md);
    min-height: 100px;
  }
  
  .stats-card .value {
    font-size: var(--font-size-2xl);
  }
  
  .stats-card .label {
    font-size: var(--font-size-xs);
  }
  
  .section-title {
    font-size: var(--font-size-xl);
  }
}

/* Dark theme adjustments */
[data-theme="dark"] .stats-card {
  background: var(--bg-secondary);
  border-color: var(--border-color);
}

[data-theme="dark"] .modern-card {
  background: var(--bg-secondary);
  border-color: var(--border-color);
}

[data-theme="dark"] .status-item {
  background: var(--bg-tertiary);
}

/* Animation classes */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-normal);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

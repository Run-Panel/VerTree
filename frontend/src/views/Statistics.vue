<template>
  <div class="statistics">
    <div class="page-header">
      <h1>{{ $t('stats.title') }}</h1>
      <el-select v-model="selectedPeriod" @change="fetchStats">
        <el-option :label="$t('statistics.period.1d')" value="1d" />
        <el-option :label="$t('statistics.period.7d')" value="7d" />
        <el-option :label="$t('statistics.period.30d')" value="30d" />
        <el-option :label="$t('statistics.period.90d')" value="90d" />
      </el-select>
    </div>

    <!-- Key Metrics -->
    <el-row :gutter="20" class="metrics-row">
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon">
              <el-icon size="32" color="#409EFF"><User /></el-icon>
            </div>
            <div class="metric-info">
              <h3>{{ stats.total_users || 0 }}</h3>
              <p>{{ $t('stats.totalUsers') }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon">
              <el-icon size="32" color="#67C23A"><Download /></el-icon>
            </div>
            <div class="metric-info">
              <h3>{{ stats.total_downloads || 0 }}</h3>
              <p>{{ $t('stats.totalDownloads') }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon">
              <el-icon size="32" color="#E6A23C"><TrendCharts /></el-icon>
            </div>
            <div class="metric-info">
              <h3>{{ (stats.success_rate || 0).toFixed(1) }}%</h3>
              <p>{{ $t('stats.successRate') }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="metric-card">
          <div class="metric-content">
            <div class="metric-icon">
              <el-icon size="32" color="#F56C6C"><PieChart /></el-icon>
            </div>
            <div class="metric-info">
              <h3>{{ Object.keys(stats.version_distribution || {}).length }}</h3>
              <p>{{ $t('statistics.activeVersions') }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts -->
    <el-row :gutter="20">
      <!-- Daily Trend Chart -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>{{ $t('stats.updateTrend') }}</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="dailyTrendOption" v-loading="loading" />
          </div>
        </el-card>
      </el-col>

      <!-- Version Distribution Chart -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>{{ $t('stats.versionDistribution') }}</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="versionDistributionOption" v-loading="loading" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Region Distribution -->
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>{{ $t('stats.regionDistribution') }}</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="regionDistributionOption" v-loading="loading" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart, { THEME_KEY } from 'vue-echarts'
import { ElMessage } from 'element-plus'
import { getStats } from '@/api/stats'

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const { t } = useI18n()
const loading = ref(false)
const selectedPeriod = ref('7d')
const stats = ref({})

const fetchStats = async () => {
  loading.value = true
  try {
    const response = await getStats({
      period: selectedPeriod.value,
      action: 'all'
    })
    stats.value = response.data || {}
  } catch (error) {
    ElMessage.error(t('statistics.fetchError'))
  } finally {
    loading.value = false
  }
}

// Daily Trend Chart Options
const dailyTrendOption = computed(() => {
  const dailyStats = stats.value.daily_stats || []
  const dates = dailyStats.map(item => item.date)
  const downloads = dailyStats.map(item => item.downloads)
  const installs = dailyStats.map(item => item.installs)
  const failures = dailyStats.map(item => item.failures)

  return {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: [t('statistics.downloads'), t('statistics.installSuccess'), t('statistics.installFailed')]
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: t('statistics.downloads'),
        type: 'line',
        data: downloads,
        smooth: true,
        lineStyle: { color: '#409EFF' },
        itemStyle: { color: '#409EFF' }
      },
      {
        name: t('statistics.installSuccess'),
        type: 'line',
        data: installs,
        smooth: true,
        lineStyle: { color: '#67C23A' },
        itemStyle: { color: '#67C23A' }
      },
      {
        name: t('statistics.installFailed'),
        type: 'line',
        data: failures,
        smooth: true,
        lineStyle: { color: '#F56C6C' },
        itemStyle: { color: '#F56C6C' }
      }
    ]
  }
})

// Version Distribution Chart Options
const versionDistributionOption = computed(() => {
  const versionDist = stats.value.version_distribution || {}
  const data = Object.entries(versionDist).map(([version, count]) => ({
    name: version,
    value: count
  }))

  return {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: t('statistics.versionDistribution'),
        type: 'pie',
        radius: '50%',
        data: data,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
})

// Region Distribution Chart Options
const regionDistributionOption = computed(() => {
  const regionDist = stats.value.region_distribution || {}
  const regions = Object.keys(regionDist)
  const counts = Object.values(regionDist)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    xAxis: {
      type: 'category',
      data: regions
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: t('statistics.userCount'),
        type: 'bar',
        data: counts,
        itemStyle: {
          color: '#409EFF'
        }
      }
    ]
  }
})

onMounted(() => {
  fetchStats()
})
</script>

<style scoped>
.statistics {
  max-width: 1200px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.metrics-row {
  margin-bottom: 20px;
}

.metric-card {
  height: 120px;
}

.metric-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.metric-icon {
  margin-right: 20px;
}

.metric-info h3 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: bold;
}

.metric-info p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.chart-card {
  margin-bottom: 20px;
}

.chart-container {
  height: 400px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

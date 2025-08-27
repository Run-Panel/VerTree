<template>
  <div class="version-management">
    <div class="page-header">
      <h1>{{ $t('version.title') }}</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        {{ $t('version.createVersion') }}
      </el-button>
    </div>

    <!-- Filters -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item :label="$t('version.channel')">
          <el-select v-model="filters.channel" :placeholder="$t('versions.allChannels')" clearable>
            <el-option :label="$t('versions.allChannels')" value="" />
            <el-option :label="$t('versions.stable')" value="stable" />
            <el-option :label="$t('versions.beta')" value="beta" />
            <el-option :label="$t('versions.alpha')" value="alpha" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchVersions">
            <el-icon><Search /></el-icon>
            {{ $t('common.search') }}
          </el-button>
          <el-button @click="resetFilters">
            {{ $t('common.refresh') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Version List -->
    <el-card class="table-card">
      <el-table :data="versions" v-loading="loading" @sort-change="handleSortChange">
        <el-table-column prop="version" :label="$t('version.versionNumber')" width="120" sortable="custom" />
        <el-table-column prop="channel" :label="$t('version.channel')" width="100">
          <template #default="scope">
            <el-tag :type="getChannelTagType(scope.row.channel)">
              {{ scope.row.channel }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" :label="$t('version.versionTitle')" />
        <el-table-column prop="publish_time" :label="$t('dashboard.publishTime')" width="180" sortable="custom">
          <template #default="scope">
            {{ formatDate(scope.row.publish_time) }}
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.status')" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_published ? 'success' : 'warning'">
              {{ scope.row.is_published ? $t('version.status.published') : $t('version.status.draft') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('version.forceUpdate')" width="100">
          <template #default="scope">
            <el-icon v-if="scope.row.is_forced" color="#F56C6C"><Warning /></el-icon>
          </template>
        </el-table-column>
        <el-table-column :label="$t('channels.actions')" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="editVersion(scope.row)">
              {{ $t('common.edit') }}
            </el-button>
            <el-button 
              v-if="!scope.row.is_published"
              size="small" 
              type="success" 
              @click="publishVersion(scope.row)"
            >
              {{ $t('common.publish') }}
            </el-button>
            <el-button 
              v-else
              size="small" 
              type="warning" 
              @click="unpublishVersion(scope.row)"
            >
              {{ $t('common.unpublish') }}
            </el-button>
            <el-popconfirm
              :title="`${$t('versions.confirmDelete')} ${scope.row.version}?`"
              @confirm="deleteVersion(scope.row)"
            >
              <template #reference>
                <el-button 
                  size="small" 
                  type="danger"
                  :disabled="scope.row.is_published"
                >
                  {{ $t('common.delete') }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.limit"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        class="pagination"
      />
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingVersion ? $t('common.edit') + $t('nav.version') : $t('version.createVersion')"
      width="800px"
      @closed="resetForm"
    >
      <version-form 
        :version="editingVersion" 
        @submit="handleSubmit"
        @cancel="showCreateDialog = false"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getVersions, deleteVersion as deleteVersionAPI, publishVersion as publishVersionAPI, unpublishVersion as unpublishVersionAPI } from '@/api/version'
import VersionForm from '@/components/VersionForm.vue'

const { t } = useI18n()
const loading = ref(false)
const showCreateDialog = ref(false)
const editingVersion = ref(null)
const versions = ref([])

const filters = reactive({
  channel: ''
})

const pagination = reactive({
  page: 1,
  limit: 20,
  total: 0
})

const fetchVersions = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      limit: pagination.limit,
      channel: filters.channel || undefined
    }
    
    const response = await getVersions(params)
    versions.value = response.data || []
    pagination.total = response.pagination?.total || 0
  } catch (error) {
    ElMessage.error(t('versions.fetchError'))
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.channel = ''
  pagination.page = 1
  fetchVersions()
}

const editVersion = (version) => {
  editingVersion.value = { ...version }
  showCreateDialog.value = true
}

const publishVersion = async (version) => {
  try {
    await publishVersionAPI(version.id)
    ElMessage.success(t('versions.publishSuccess'))
    fetchVersions()
  } catch (error) {
    ElMessage.error(t('versions.publishFailed'))
  }
}

const unpublishVersion = async (version) => {
  try {
    await ElMessageBox.confirm(t('versions.unpublishConfirm'), t('versions.confirmAction'))
    await unpublishVersionAPI(version.id)
    ElMessage.success(t('versions.unpublishSuccess'))
    fetchVersions()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('versions.unpublishFailed'))
    }
  }
}

const deleteVersion = async (version) => {
  try {
    await deleteVersionAPI(version.id)
    ElMessage.success(t('versions.deleteSuccess'))
    fetchVersions()
  } catch (error) {
    ElMessage.error(t('versions.deleteFailed'))
  }
}

const handleSubmit = () => {
  showCreateDialog.value = false
  editingVersion.value = null
  fetchVersions()
}

const resetForm = () => {
  editingVersion.value = null
}

const handleSizeChange = (val) => {
  pagination.limit = val
  pagination.page = 1
  fetchVersions()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchVersions()
}

const handleSortChange = ({ prop, order }) => {
  // Handle sorting if needed
  fetchVersions()
}

const getChannelTagType = (channel) => {
  const typeMap = {
    'stable': 'success',
    'beta': 'warning',
    'alpha': 'info'
  }
  return typeMap[channel] || 'info'
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString()
}

onMounted(() => {
  fetchVersions()
})
</script>

<style scoped>
.version-management {
  max-width: 1200px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-form {
  margin: 0;
}

.table-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}
</style>

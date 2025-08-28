<template>
  <div class="applications-container">
    <div class="page-header">
      <h1>应用管理</h1>
      <p>管理您的应用程序和API密钥</p>
    </div>

    <!-- Applications list -->
    <div class="content-section">
      <div class="section-header">
        <h2>应用列表</h2>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          创建应用
        </el-button>
      </div>

      <el-table
        :data="applications"
        v-loading="loading"
        class="applications-table"
        @row-click="handleRowClick"
      >
        <el-table-column prop="name" label="应用名称" min-width="200">
          <template #default="scope">
            <div class="app-name-cell">
              <img v-if="scope.row.icon_url" :src="scope.row.icon_url" class="app-icon" />
              <div>
                <div class="app-name">{{ scope.row.name }}</div>
                <div class="app-id">{{ scope.row.app_id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="description" label="描述" min-width="300" />
        
        <el-table-column prop="keys_count" label="API密钥数量" width="120">
          <template #default="scope">
            <el-tag>{{ scope.row.keys_count || 0 }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="is_active" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'danger'">
              {{ scope.row.is_active ? '活跃' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click.stop="editApplication(scope.row)">编辑</el-button>
            <el-button size="small" @click.stop="viewKeys(scope.row)">密钥</el-button>
            <el-button
              size="small"
              type="danger"
              @click.stop="deleteApplication(scope.row)"
            >
              删除
            </el-button>
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

    <!-- Create/Edit Application Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingApp ? '编辑应用' : '创建应用'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="appFormRef"
        :model="appForm"
        :rules="appFormRules"
        label-width="100px"
      >
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="appForm.name" placeholder="请输入应用名称" />
        </el-form-item>
        
        <el-form-item label="应用描述" prop="description">
          <el-input
            v-model="appForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入应用描述"
          />
        </el-form-item>
        
        <el-form-item label="图标URL" prop="icon_url">
          <el-input v-model="appForm.icon_url" placeholder="请输入图标URL（可选）" />
        </el-form-item>
        
        <el-form-item label="状态">
          <el-switch
            v-model="appForm.is_active"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ editingApp ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- API Keys Dialog -->
    <el-dialog
      v-model="showKeysDialog"
      title="API密钥管理"
      width="800px"
      @close="resetKeysDialog"
    >
      <div class="keys-header">
        <h3>{{ selectedApp?.name }} - API密钥</h3>
        <el-button type="primary" size="small" @click="showCreateKeyDialog = true">
          <el-icon><Plus /></el-icon>
          创建密钥
        </el-button>
      </div>

      <el-table :data="apiKeys" v-loading="keysLoading">
        <el-table-column prop="name" label="密钥名称" />
        <el-table-column prop="key_id" label="密钥ID" />
        <el-table-column prop="permissions" label="权限">
          <template #default="scope">
            <el-tag
              v-for="permission in scope.row.permissions"
              :key="permission"
              size="small"
              class="permission-tag"
            >
              {{ permission }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_active" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'danger'" size="small">
              {{ scope.row.is_active ? '活跃' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_used" label="最后使用" width="150">
          <template #default="scope">
            {{ scope.row.last_used ? formatDate(scope.row.last_used) : '未使用' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button size="small" @click="editKey(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteKey(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Create/Edit API Key Dialog -->
    <el-dialog
      v-model="showCreateKeyDialog"
      :title="editingKey ? '编辑密钥' : '创建密钥'"
      width="500px"
      @close="resetKeyForm"
    >
      <el-form
        ref="keyFormRef"
        :model="keyForm"
        :rules="keyFormRules"
        label-width="100px"
      >
        <el-form-item label="密钥名称" prop="name">
          <el-input v-model="keyForm.name" placeholder="请输入密钥名称" />
        </el-form-item>
        
        <el-form-item label="权限" prop="permissions">
          <el-checkbox-group v-model="keyForm.permissions">
            <el-checkbox label="check_update">检查更新</el-checkbox>
            <el-checkbox label="download">下载记录</el-checkbox>
            <el-checkbox label="install">安装记录</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        
        <el-form-item label="状态">
          <el-switch
            v-model="keyForm.is_active"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateKeyDialog = false">取消</el-button>
          <el-button type="primary" @click="submitKeyForm" :loading="keySubmitting">
            {{ editingKey ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- API Key Created Success Dialog -->
    <el-dialog
      v-model="showKeyCreatedDialog"
      title="密钥创建成功"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="key-created-content">
        <el-alert
          title="重要提醒"
          type="warning"
          description="API密钥只会显示一次，请立即保存到安全的地方。"
          show-icon
          :closable="false"
        />
        
        <div class="key-info">
          <h4>应用ID:</h4>
          <el-input
            :value="selectedApp?.app_id"
            readonly
            class="key-display"
          >
            <template #append>
              <el-button @click="copyToClipboard(selectedApp?.app_id)">复制</el-button>
            </template>
          </el-input>
          
          <h4>API密钥:</h4>
          <el-input
            :value="createdKeySecret"
            readonly
            class="key-display"
          >
            <template #append>
              <el-button @click="copyToClipboard(createdKeySecret)">复制</el-button>
            </template>
          </el-input>
          
          <h4>Authorization Header:</h4>
          <el-input
            :value="`Bearer ${selectedApp?.app_id}:${createdKeySecret}`"
            readonly
            class="key-display"
          >
            <template #append>
              <el-button @click="copyToClipboard(`Bearer ${selectedApp?.app_id}:${createdKeySecret}`)">复制</el-button>
            </template>
          </el-input>
        </div>
      </div>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button type="primary" @click="showKeyCreatedDialog = false">我已保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getApplications,
  createApplication,
  updateApplication,
  deleteApplication as deleteApp,
  getApplicationKeys,
  createApplicationKey,
  updateApplicationKey,
  deleteApplicationKey
} from '@/api/application'

// Reactive data
const loading = ref(false)
const applications = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// Application form
const showCreateDialog = ref(false)
const submitting = ref(false)
const editingApp = ref(null)
const appFormRef = ref()
const appForm = reactive({
  name: '',
  description: '',
  icon_url: '',
  is_active: true
})

const appFormRules = {
  name: [
    { required: true, message: '请输入应用名称', trigger: 'blur' },
    { min: 2, max: 100, message: '应用名称长度在 2 到 100 个字符', trigger: 'blur' }
  ]
}

// API Keys management
const showKeysDialog = ref(false)
const showCreateKeyDialog = ref(false)
const showKeyCreatedDialog = ref(false)
const keysLoading = ref(false)
const keySubmitting = ref(false)
const selectedApp = ref(null)
const apiKeys = ref([])
const editingKey = ref(null)
const createdKeySecret = ref('')
const keyFormRef = ref()
const keyForm = reactive({
  name: '',
  permissions: ['check_update'],
  is_active: true
})

const keyFormRules = {
  name: [
    { required: true, message: '请输入密钥名称', trigger: 'blur' },
    { min: 2, max: 100, message: '密钥名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  permissions: [
    { required: true, message: '请选择至少一个权限', trigger: 'change' }
  ]
}

// Methods
const fetchApplications = async () => {
  loading.value = true
  try {
    const response = await getApplications({
      page: currentPage.value,
      limit: pageSize.value
    })
    
    if (response.data) {
      // API返回格式: { code: 200, message: "success", data: [...], pagination: {...} }
      applications.value = response.data || []
      total.value = response.pagination?.total || 0
    }
  } catch (error) {
    ElMessage.error('获取应用列表失败')
    console.error('Failed to fetch applications:', error)
  } finally {
    loading.value = false
  }
}

const handlePageSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  fetchApplications()
}

const handleCurrentPageChange = (page) => {
  currentPage.value = page
  fetchApplications()
}

const handleRowClick = (row) => {
  viewKeys(row)
}

const editApplication = (app) => {
  editingApp.value = app
  appForm.name = app.name
  appForm.description = app.description || ''
  appForm.icon_url = app.icon_url || ''
  appForm.is_active = app.is_active
  showCreateDialog.value = true
}

const resetForm = () => {
  editingApp.value = null
  appForm.name = ''
  appForm.description = ''
  appForm.icon_url = ''
  appForm.is_active = true
  appFormRef.value?.clearValidate()
}

const submitForm = async () => {
  if (!appFormRef.value) return
  
  const valid = await appFormRef.value.validate().catch(() => false)
  if (!valid) return
  
  submitting.value = true
  try {
    const formData = { ...appForm }
    
    if (editingApp.value) {
      await updateApplication(editingApp.value.app_id, formData)
      ElMessage.success('应用更新成功')
    } else {
      await createApplication(formData)
      ElMessage.success('应用创建成功')
    }
    
    showCreateDialog.value = false
    fetchApplications()
  } catch (error) {
    ElMessage.error(editingApp.value ? '应用更新失败' : '应用创建失败')
    console.error('Failed to save application:', error)
  } finally {
    submitting.value = false
  }
}

const deleteApplication = async (app) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除应用 "${app.name}" 吗？此操作将同时删除该应用下的所有版本、渠道和API密钥。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteApp(app.app_id)
    ElMessage.success('应用删除成功')
    fetchApplications()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('应用删除失败')
      console.error('Failed to delete application:', error)
    }
  }
}

// API Keys methods
const viewKeys = async (app) => {
  selectedApp.value = app
  showKeysDialog.value = true
  await fetchApiKeys()
}

const fetchApiKeys = async () => {
  if (!selectedApp.value) return
  
  keysLoading.value = true
  try {
    const response = await getApplicationKeys(selectedApp.value.app_id)
    // API返回格式: { code: 200, message: "success", data: [...] | null }
    apiKeys.value = response.data || []
  } catch (error) {
    ElMessage.error('获取API密钥失败')
    console.error('Failed to fetch API keys:', error)
  } finally {
    keysLoading.value = false
  }
}

const editKey = (key) => {
  editingKey.value = key
  keyForm.name = key.name
  keyForm.permissions = [...key.permissions]
  keyForm.is_active = key.is_active
  showCreateKeyDialog.value = true
}

const resetKeyForm = () => {
  editingKey.value = null
  keyForm.name = ''
  keyForm.permissions = ['check_update']
  keyForm.is_active = true
  keyFormRef.value?.clearValidate()
}

const resetKeysDialog = () => {
  selectedApp.value = null
  apiKeys.value = []
  resetKeyForm()
}

const submitKeyForm = async () => {
  if (!keyFormRef.value) return
  
  const valid = await keyFormRef.value.validate().catch(() => false)
  if (!valid) return
  
  keySubmitting.value = true
  try {
    const formData = { ...keyForm }
    
    if (editingKey.value) {
      await updateApplicationKey(selectedApp.value.app_id, editingKey.value.key_id, formData)
      ElMessage.success('密钥更新成功')
      showCreateKeyDialog.value = false
      fetchApiKeys()
    } else {
      const response = await createApplicationKey(selectedApp.value.app_id, formData)
      createdKeySecret.value = response.data.key_secret
      showCreateKeyDialog.value = false
      showKeyCreatedDialog.value = true
      fetchApiKeys()
    }
  } catch (error) {
    ElMessage.error(editingKey.value ? '密钥更新失败' : '密钥创建失败')
    console.error('Failed to save API key:', error)
  } finally {
    keySubmitting.value = false
  }
}

const deleteKey = async (key) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除密钥 "${key.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteApplicationKey(selectedApp.value.app_id, key.key_id)
    ElMessage.success('密钥删除成功')
    fetchApiKeys()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('密钥删除失败')
      console.error('Failed to delete API key:', error)
    }
  }
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
    console.error('Failed to copy to clipboard:', error)
  }
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleString('zh-CN')
}

// Lifecycle
onMounted(() => {
  fetchApplications()
})
</script>

<style scoped>
.applications-container {
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

.applications-table {
  width: 100%;
  margin-bottom: 20px;
}

.app-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-icon {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
}

.app-name {
  font-weight: 600;
  color: #303133;
}

.app-id {
  font-size: 12px;
  color: #909399;
  font-family: monospace;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.keys-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.keys-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.permission-tag {
  margin-right: 4px;
}

.key-created-content {
  padding: 20px 0;
}

.key-info {
  margin-top: 20px;
}

.key-info h4 {
  margin: 16px 0 8px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.key-display {
  margin-bottom: 12px;
}

.key-display :deep(.el-input__inner) {
  font-family: monospace;
  font-size: 12px;
}
</style>

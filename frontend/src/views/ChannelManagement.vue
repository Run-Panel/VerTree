<template>
  <div class="channel-management">
    <div class="page-header">
      <h1>{{ $t('channel.title') }}</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        {{ $t('channels.create') }}
      </el-button>
    </div>

    <!-- Channel List -->
    <el-card class="table-card">
      <el-table :data="channels" v-loading="loading">
        <el-table-column prop="name" :label="$t('channel.name')" width="120" />
        <el-table-column prop="display_name" :label="$t('channel.displayName')" width="150" />
        <el-table-column prop="description" :label="$t('channel.description')" />
        <el-table-column :label="$t('channel.isActive')" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'danger'">
              {{ scope.row.is_active ? $t('channels.active') : $t('channels.inactive') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('channel.autoPublish')" width="100">
          <template #default="scope">
            <el-icon v-if="scope.row.auto_publish" color="#67C23A"><Check /></el-icon>
            <el-icon v-else color="#F56C6C"><Close /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="rollout_percentage" :label="$t('channel.rolloutPercentage')" width="120">
          <template #default="scope">
            {{ scope.row.rollout_percentage }}%
          </template>
        </el-table-column>
        <el-table-column :label="$t('channels.actions')" width="180" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="editChannel(scope.row)">
              {{ $t('common.edit') }}
            </el-button>
            <el-popconfirm
              :title="`${$t('channels.confirmDelete')} ${scope.row.name}?`"
              @confirm="deleteChannel(scope.row)"
            >
              <template #reference>
                <el-button size="small" type="danger">
                  {{ $t('common.delete') }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingChannel ? $t('channels.editTitle') : $t('channels.createTitle')"
      width="600px"
      @closed="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item :label="$t('channel.name')" prop="name">
          <el-input v-model="form.name" :disabled="!!editingChannel" />
        </el-form-item>
        
        <el-form-item :label="$t('channel.displayName')" prop="display_name">
          <el-input v-model="form.display_name" />
        </el-form-item>
        
        <el-form-item :label="$t('channel.description')" prop="description">
          <el-input type="textarea" v-model="form.description" :rows="3" />
        </el-form-item>
        
        <el-form-item :label="$t('channel.isActive')">
          <el-switch v-model="form.is_active" />
        </el-form-item>
        
        <el-form-item :label="$t('channel.autoPublish')">
          <el-switch v-model="form.auto_publish" />
          <span class="form-help">{{ $t('channels.autoPublishHelp') }}</span>
        </el-form-item>
        
        <el-form-item :label="$t('channel.rolloutPercentage')" prop="rollout_percentage">
          <el-slider v-model="form.rollout_percentage" :min="0" :max="100" show-input />
          <span class="form-help">{{ form.rollout_percentage }}%{{ $t('channels.rolloutHelp') }}</span>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="submitForm">
            {{ $t('common.confirm') }}
          </el-button>
          <el-button @click="showCreateDialog = false">
            {{ $t('common.cancel') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getChannels, createChannel, updateChannel, deleteChannel as deleteChannelAPI } from '@/api/channel'

const { t } = useI18n()
const loading = ref(false)
const showCreateDialog = ref(false)
const editingChannel = ref(null)
const channels = ref([])
const formRef = ref()

const form = reactive({
  name: '',
  display_name: '',
  description: '',
  is_active: true,
  auto_publish: false,
  rollout_percentage: 100
})

const rules = {
  name: [
    { required: true, message: t('channels.nameRequired'), trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_-]*$/, message: t('channels.namePattern'), trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: t('channels.displayNameRequired'), trigger: 'blur' }
  ],
  rollout_percentage: [
    { required: true, message: t('channels.rolloutRequired'), trigger: 'blur' },
    { type: 'number', min: 0, max: 100, message: t('channels.rolloutRange'), trigger: 'blur' }
  ]
}

const fetchChannels = async () => {
  loading.value = true
  try {
    const response = await getChannels()
    channels.value = response.data || []
  } catch (error) {
    ElMessage.error(t('channels.fetchError'))
  } finally {
    loading.value = false
  }
}

const editChannel = (channel) => {
  editingChannel.value = channel
  Object.keys(form).forEach(key => {
    if (channel[key] !== undefined) {
      form[key] = channel[key]
    }
  })
  showCreateDialog.value = true
}

const deleteChannel = async (channel) => {
  try {
    await deleteChannelAPI(channel.id)
    ElMessage.success(t('channels.deleteSuccess'))
    fetchChannels()
  } catch (error) {
    ElMessage.error(t('channels.deleteFailed'))
  }
}

const submitForm = async () => {
  try {
    await formRef.value.validate()
    
    if (editingChannel.value) {
      await updateChannel(editingChannel.value.id, form)
      ElMessage.success(t('channels.updateSuccess'))
    } else {
      await createChannel(form)
      ElMessage.success(t('channels.createSuccess'))
    }
    
    showCreateDialog.value = false
    fetchChannels()
  } catch (error) {
    if (error.errors) {
      return
    }
    ElMessage.error(editingChannel.value ? t('channels.updateFailed') : t('channels.createFailed'))
  }
}

const resetForm = () => {
  editingChannel.value = null
  Object.keys(form).forEach(key => {
    if (key === 'is_active') {
      form[key] = true
    } else if (key === 'auto_publish') {
      form[key] = false
    } else if (key === 'rollout_percentage') {
      form[key] = 100
    } else {
      form[key] = ''
    }
  })
}

onMounted(() => {
  fetchChannels()
})
</script>

<style scoped>
.channel-management {
  max-width: 1200px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.form-help {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}
</style>

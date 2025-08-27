<template>
  <el-form :model="form" :rules="rules" ref="formRef" label-width="140px">
    <el-form-item :label="$t('version.versionNumber')" prop="version">
      <el-input v-model="form.version" placeholder="v1.2.3" />
    </el-form-item>
    
    <el-form-item :label="$t('version.channel')" prop="channel">
      <el-select v-model="form.channel" placeholder="选择发布通道">
        <el-option label="稳定版" value="stable" />
        <el-option label="测试版" value="beta" />
        <el-option label="预览版" value="alpha" />
      </el-select>
    </el-form-item>
    
    <el-form-item :label="$t('version.versionTitle')" prop="title">
      <el-input v-model="form.title" />
    </el-form-item>
    
    <el-form-item :label="$t('version.description')" prop="description">
      <el-input type="textarea" v-model="form.description" :rows="3" />
    </el-form-item>
    
    <el-form-item :label="$t('version.releaseNotes')" prop="release_notes">
      <el-input type="textarea" v-model="form.release_notes" :rows="6" placeholder="支持Markdown格式" />
    </el-form-item>
    
    <el-form-item :label="$t('version.breakingChanges')">
      <el-input type="textarea" v-model="form.breaking_changes" :rows="2" />
    </el-form-item>
    
    <el-form-item :label="$t('version.minUpgradeVersion')">
      <el-input v-model="form.min_upgrade_version" placeholder="v1.0.0" />
    </el-form-item>
    
    <el-form-item label="文件URL" prop="file_url">
      <el-input v-model="form.file_url" placeholder="https://releases.example.com/v1.2.3/app" />
    </el-form-item>
    
    <el-form-item label="文件大小 (字节)" prop="file_size">
      <el-input-number v-model="form.file_size" :min="1" controls-position="right" class="full-width" />
    </el-form-item>
    
    <el-form-item label="文件校验和" prop="file_checksum">
      <el-input v-model="form.file_checksum" placeholder="sha256:abc123..." />
    </el-form-item>
    
    <el-form-item :label="$t('version.forceUpdate')">
      <el-switch v-model="form.is_forced" />
      <span class="form-help">开启后，客户端将被强制更新</span>
    </el-form-item>
    
    <el-form-item>
      <el-button type="primary" @click="submitForm('draft')">
        {{ $t('version.actions.saveDraft') }}
      </el-button>
      <el-button type="success" @click="submitForm('publish')">
        {{ $t('version.actions.publishNow') }}
      </el-button>
      <el-button @click="$emit('cancel')">
        {{ $t('common.cancel') }}
      </el-button>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createVersion, updateVersion, publishVersion as publishVersionAPI } from '@/api/version'

const props = defineProps({
  version: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['submit', 'cancel'])

const formRef = ref()
const form = reactive({
  version: '',
  channel: 'stable',
  title: '',
  description: '',
  release_notes: '',
  breaking_changes: '',
  min_upgrade_version: '',
  file_url: '',
  file_size: null,
  file_checksum: '',
  is_forced: false
})

const rules = {
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' }
  ],
  channel: [
    { required: true, message: '请选择发布通道', trigger: 'change' }
  ],
  title: [
    { required: true, message: '请输入版本标题', trigger: 'blur' }
  ],
  file_url: [
    { required: true, message: '请输入文件URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ],
  file_size: [
    { required: true, message: '请输入文件大小', trigger: 'blur' },
    { type: 'number', min: 1, message: '文件大小必须大于0', trigger: 'blur' }
  ],
  file_checksum: [
    { required: true, message: '请输入文件校验和', trigger: 'blur' }
  ]
}

// Watch for version prop changes
watch(() => props.version, (newVersion) => {
  if (newVersion) {
    Object.keys(form).forEach(key => {
      if (newVersion[key] !== undefined) {
        form[key] = newVersion[key]
      }
    })
  } else {
    resetForm()
  }
}, { immediate: true })

const resetForm = () => {
  Object.keys(form).forEach(key => {
    if (key === 'channel') {
      form[key] = 'stable'
    } else if (key === 'is_forced') {
      form[key] = false
    } else if (key === 'file_size') {
      form[key] = null
    } else {
      form[key] = ''
    }
  })
}

const submitForm = async (action) => {
  try {
    await formRef.value.validate()
    
    const formData = { ...form }
    
    let response
    if (props.version) {
      // Update existing version
      response = await updateVersion(props.version.id, formData)
    } else {
      // Create new version
      response = await createVersion(formData)
    }
    
    // If action is publish, publish the version
    if (action === 'publish') {
      await publishVersionAPI(response.data.id)
      ElMessage.success(props.version ? '版本更新并发布成功' : '版本创建并发布成功')
    } else {
      ElMessage.success(props.version ? '版本更新成功' : '版本创建成功')
    }
    
    emit('submit')
  } catch (error) {
    if (error.errors) {
      // Validation errors
      return
    }
    ElMessage.error(props.version ? '版本更新失败' : '版本创建失败')
  }
}
</script>

<style scoped>
.full-width {
  width: 100%;
}

.form-help {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}
</style>

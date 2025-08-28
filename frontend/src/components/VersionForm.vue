<template>
  <el-form :model="form" :rules="rules" ref="formRef" label-width="140px">
    <el-form-item :label="$t('version.application')" prop="app_id">
      <el-select v-model="form.app_id" placeholder="选择应用" filterable>
        <el-option 
          v-for="app in applications" 
          :key="app.app_id" 
          :label="app.name" 
          :value="app.app_id" 
        />
      </el-select>
    </el-form-item>
    
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
    
    <el-form-item label="文件来源" prop="file_source">
      <el-radio-group v-model="fileSourceType" @change="handleFileSourceChange">
        <el-radio value="url">URL链接</el-radio>
        <el-radio value="upload">本地上传</el-radio>
      </el-radio-group>
    </el-form-item>
    
    <!-- URL 输入模式 -->
    <el-form-item v-if="fileSourceType === 'url'" label="文件URL" prop="file_url">
      <el-input 
        v-model="form.file_url" 
        placeholder="https://releases.example.com/v1.2.3/app"
        @blur="handleUrlBlur"
      />
      <div class="form-help">支持直接输入文件下载链接</div>
    </el-form-item>
    
    <!-- 文件上传模式 -->
    <el-form-item v-if="fileSourceType === 'upload'" label="上传文件" prop="file_upload">
      <el-upload
        ref="uploadRef"
        class="upload-demo"
        drag
        :auto-upload="false"
        :on-change="handleFileChange"
        :on-remove="handleFileRemove"
        :limit="1"
        :show-file-list="true"
        accept=".zip,.exe,.dmg,.pkg,.deb,.rpm,.tar.gz,.msi"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持 .zip, .exe, .dmg, .pkg, .deb, .rpm, .tar.gz, .msi 等格式，文件大小不超过 500MB
          </div>
        </template>
      </el-upload>
      <div class="form-help">选择要发布的版本文件，系统将自动计算文件大小和校验和</div>
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
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { createVersion, updateVersion, publishVersion as publishVersionAPI, createVersionWithFile, updateVersionWithFile } from '@/api/version'
import { getApplications } from '@/api/application'
import CryptoJS from 'crypto-js'

const props = defineProps({
  version: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['submit', 'cancel'])

// 定义响应式变量
const formRef = ref()
const uploadRef = ref()
const fileSourceType = ref('url') // 'url' or 'upload'
const uploadedFile = ref(null)
const applications = ref([])

// 定义表单数据
const form = reactive({
  app_id: '',
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
  app_id: [
    { required: true, message: '请选择应用', trigger: 'change' }
  ],
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
    { 
      validator: (rule, value, callback) => {
        if (fileSourceType.value === 'url') {
          if (!value) {
            callback(new Error('请输入文件URL'))
          } else if (!/^https?:\/\/.+/.test(value)) {
            callback(new Error('请输入有效的URL'))
          } else {
            callback()
          }
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ],
  file_upload: [
    { 
      validator: (rule, value, callback) => {
        if (fileSourceType.value === 'upload' && !uploadedFile.value) {
          callback(new Error('请选择要上传的文件'))
        } else {
          callback()
        }
      }, 
      trigger: 'change' 
    }
  ],
  file_size: [
    { required: true, message: '请输入文件大小', trigger: 'blur' },
    { type: 'number', min: 1, message: '文件大小必须大于0', trigger: 'blur' }
  ],
  file_checksum: [
    { required: true, message: '请输入文件校验和', trigger: 'blur' }
  ]
}

// 重置表单函数
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
  fileSourceType.value = 'url'
  uploadedFile.value = null
  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }
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

// 文件来源切换处理
const handleFileSourceChange = (type) => {
  if (type === 'upload') {
    form.file_url = ''
  } else {
    uploadedFile.value = null
    if (uploadRef.value) {
      uploadRef.value.clearFiles()
    }
  }
}

// URL输入失焦处理
const handleUrlBlur = () => {
  // 可以在这里添加URL验证逻辑
}

// 文件上传变化处理
const handleFileChange = (file) => {
  uploadedFile.value = file
  
  // 自动填充文件大小
  if (file.raw) {
    form.file_size = file.raw.size
    
    // 计算文件校验和
    calculateFileChecksum(file.raw)
    
    // 自动生成文件URL（基于文件名）
    const fileName = file.name
    form.file_url = `/uploads/versions/${fileName}`
  }
}

// 文件移除处理
const handleFileRemove = () => {
  uploadedFile.value = null
  form.file_size = null
  form.file_checksum = ''
  form.file_url = ''
}

// 计算文件校验和
const calculateFileChecksum = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    const wordArray = CryptoJS.lib.WordArray.create(e.target.result)
    const hash = CryptoJS.SHA256(wordArray).toString()
    form.file_checksum = `sha256:${hash}`
  }
  reader.readAsArrayBuffer(file)
}

// 获取应用列表
const fetchApplications = async () => {
  try {
    console.log('Fetching applications...')
    const response = await getApplications()
    console.log('Applications response:', response)
    
    // 处理不同的响应格式
    let appData = []
    if (response && response.data) {
      if (Array.isArray(response.data)) {
        appData = response.data
      } else if (response.data.data && Array.isArray(response.data.data)) {
        appData = response.data.data
      }
    } else if (Array.isArray(response)) {
      appData = response
    }
    
    applications.value = appData
    console.log('Applications loaded:', applications.value)
    
    if (appData.length === 0) {
      ElMessage.warning('暂无应用数据，请先创建应用')
    }
  } catch (error) {
    console.error('Failed to fetch applications:', error)
    applications.value = []
    ElMessage.error('获取应用列表失败: ' + (error.message || '网络错误'))
  }
}

// 组件挂载时获取应用列表
onMounted(async () => {
  await fetchApplications()
})

const submitForm = async (action) => {
  try {
    await formRef.value.validate()
    
    let response
    
    if (fileSourceType.value === 'upload' && uploadedFile.value) {
      // 文件上传模式
      const formData = new FormData()
      
      // 添加文件
      formData.append('file', uploadedFile.value.raw)
      
      // 添加版本信息
      Object.keys(form).forEach(key => {
        if (form[key] !== null && form[key] !== '') {
          formData.append(key, form[key])
        }
      })
      
      // 标记为发布模式
      if (action === 'publish') {
        formData.append('publish', 'true')
      }
      
      if (props.version) {
        // 更新现有版本（包含文件）
        response = await updateVersionWithFile(props.version.id, formData)
      } else {
        // 创建新版本（包含文件）
        response = await createVersionWithFile(formData)
      }
    } else {
      // URL模式
      const formData = { ...form }
      
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
      }
    }
    
    ElMessage.success(
      action === 'publish' 
        ? (props.version ? '版本更新并发布成功' : '版本创建并发布成功')
        : (props.version ? '版本更新成功' : '版本创建成功')
    )
    
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

<template>
  <el-dialog
    v-model="visible"
    title="导入邮箱文件"
    width="600px"
    :before-close="handleClose"
  >
    <div class="file-import-content">
      <!-- 格式说明 -->
      <el-alert
        title="文件格式说明"
        type="info"
        :closable="false"
        show-icon
      >
        <p>支持导入 <strong>.txt</strong> 或 <strong>.csv</strong> 格式的文件，每行一个邮箱：</p>
        <p><strong>格式1：</strong> 邮箱----密码----客户端ID----刷新令牌----备注</p>
        <p><strong>格式2：</strong> 邮箱,密码,客户端ID,刷新令牌,备注</p>
        <p class="note">注：备注为可选项，其他字段必填</p>
      </el-alert>

      <!-- 数量限制说明 -->
      <el-alert
        title="数量限制"
        type="warning"
        :closable="false"
        show-icon
        style="margin-top: 12px;"
      >
        <p>为了确保系统稳定性和验证效果，一次性最多只能导入30个邮箱账号。</p>
        <p>如需导入更多账号，请分批次进行操作。</p>
      </el-alert>

      <!-- 文件上传区域 -->
      <div
        class="upload-area"
        :class="{ 'drag-over': isDragOver }"
        style="margin-top: 20px;"
        @dragenter="isDragOver = true"
        @dragleave="isDragOver = false"
        @drop="isDragOver = false"
      >
        <el-upload
          ref="uploadRef"
          class="upload-dragger"
          drag
          :auto-upload="false"
          :show-file-list="true"
          :limit="1"
          accept=".txt,.csv"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :on-exceed="handleExceed"
        >
          <div class="upload-content">
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <div class="el-upload__tip">
              只能上传 .txt/.csv 文件，且不超过 2MB
            </div>
          </div>
        </el-upload>
      </div>

      <!-- 文件预览 -->
      <div v-if="fileContent" class="file-preview" style="margin-top: 20px;">
        <h4>文件内容预览：</h4>
        <el-input
          v-model="fileContent"
          type="textarea"
          :rows="8"
          readonly
          style="font-family: monospace; font-size: 12px;"
        />
        <div class="preview-info">
          <el-tag type="success">检测到 {{ parsedEmails.length }} 个邮箱</el-tag>
          <el-tag v-if="parseErrors.length > 0" type="danger" style="margin-left: 8px;">
            {{ parseErrors.length }} 个错误
          </el-tag>
        </div>
      </div>

      <!-- 错误信息 -->
      <div v-if="parseErrors.length > 0" class="error-list" style="margin-top: 12px;">
        <el-alert
          title="解析错误"
          type="error"
          :closable="false"
          show-icon
        >
          <ul style="margin: 0; padding-left: 20px;">
            <li v-for="(error, index) in parseErrors.slice(0, 5)" :key="index">
              {{ error }}
            </li>
            <li v-if="parseErrors.length > 5">
              ... 还有 {{ parseErrors.length - 5 }} 个错误
            </li>
          </ul>
        </el-alert>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          :loading="loading"
          :disabled="!selectedFile || parsedEmails.length === 0"
          @click="handleImport"
        >
          导入邮箱 ({{ parsedEmails.length }})
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import type { UploadFile, UploadFiles, UploadInstance } from 'element-plus'
import { emailAPI, type AddEmailRequest } from '@/api'

interface Props {
  modelValue: boolean
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const visible = ref(false)
const loading = ref(false)
const uploadRef = ref<UploadInstance>()
const selectedFile = ref<File | null>(null)
const fileContent = ref('')
const parsedEmails = ref<AddEmailRequest[]>([])
const parseErrors = ref<string[]>([])
const isDragOver = ref(false)

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (!val) {
    resetForm()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const resetForm = () => {
  selectedFile.value = null
  fileContent.value = ''
  parsedEmails.value = []
  parseErrors.value = []
  uploadRef.value?.clearFiles()
}

const handleFileChange = (file: UploadFile, files: UploadFiles) => {
  if (file.raw) {
    selectedFile.value = file.raw
    readFile(file.raw)
  }
}

const handleFileRemove = () => {
  selectedFile.value = null
  fileContent.value = ''
  parsedEmails.value = []
  parseErrors.value = []
}

const handleExceed = () => {
  ElMessage.warning('只能选择一个文件')
}

const readFile = (file: File) => {
  // 检查文件大小（2MB限制）
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 2MB')
    uploadRef.value?.clearFiles()
    return
  }

  // 检查文件类型
  const allowedTypes = ['.txt', '.csv']
  const fileName = file.name.toLowerCase()
  const isValidType = allowedTypes.some(type => fileName.endsWith(type))
  
  if (!isValidType) {
    ElMessage.error('只支持 .txt 和 .csv 格式的文件')
    uploadRef.value?.clearFiles()
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    const content = e.target?.result as string
    fileContent.value = content
    parseFileContent(content)
  }
  reader.onerror = () => {
    ElMessage.error('文件读取失败')
  }
  reader.readAsText(file, 'UTF-8')
}

const parseFileContent = (content: string) => {
  parsedEmails.value = []
  parseErrors.value = []

  const lines = content.split('\n').filter(line => line.trim())
  
  if (lines.length === 0) {
    parseErrors.value.push('文件内容为空')
    return
  }

  lines.forEach((line, index) => {
    const trimmedLine = line.trim()
    if (!trimmedLine) return

    // 支持两种分隔符：---- 和 ,
    let parts: string[]
    if (trimmedLine.includes('----')) {
      parts = trimmedLine.split('----').map(part => part.trim())
    } else {
      parts = trimmedLine.split(',').map(part => part.trim())
    }

    if (parts.length < 4) {
      parseErrors.value.push(`第${index + 1}行格式错误，字段不足: ${trimmedLine}`)
      return
    }

    const email: AddEmailRequest = {
      email_address: parts[0],
      password: parts[1],
      client_id: parts[2],
      refresh_token: parts[3],
      remark: parts[4] || ''
    }

    // 基本验证
    if (!email.email_address || !email.password || !email.client_id || !email.refresh_token) {
      parseErrors.value.push(`第${index + 1}行必填字段为空: ${trimmedLine}`)
      return
    }

    // 邮箱格式验证
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(email.email_address)) {
      parseErrors.value.push(`第${index + 1}行邮箱格式错误: ${email.email_address}`)
      return
    }

    parsedEmails.value.push(email)
  })

  // 检查数量限制
  if (parsedEmails.value.length > 30) {
    ElMessage.error(`解析到 ${parsedEmails.value.length} 个邮箱，超过最大限制30个，请减少邮箱数量`)
    parsedEmails.value = parsedEmails.value.slice(0, 30) // 只保留前30个
    ElMessage.warning('已自动截取前30个邮箱')
  }

  if (parsedEmails.value.length > 0) {
    ElMessage.success(`成功解析 ${parsedEmails.value.length} 个邮箱`)
  }

  if (parseErrors.value.length > 0) {
    ElMessage.warning(`发现 ${parseErrors.value.length} 个错误`)
  }
}

const handleImport = async () => {
  if (parsedEmails.value.length === 0) {
    ElMessage.warning('没有有效的邮箱数据')
    return
  }

  try {
    loading.value = true
    const response = await emailAPI.batchAddEmails({
      emails: parsedEmails.value
    })

    if (response.data.success) {
      ElMessage.success('文件导入成功')
      emit('success')
      handleClose()
    } else {
      ElMessage.error(response.data.message || '文件导入失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '文件导入失败')
  } finally {
    loading.value = false
  }
}

const handleClose = () => {
  visible.value = false
}
</script>

<style scoped>
.file-import-content {
  max-height: 70vh;
  overflow-y: auto;
  padding: 4px;
}

.note {
  color: #909399;
  font-size: 12px;
  margin: 4px 0 0 0;
  font-style: italic;
}

/* 添加进入动画 */
.file-import-content > * {
  animation: fadeInUp 0.3s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 文件拖拽状态 */
.upload-area.drag-over {
  border-color: #409eff;
  background: linear-gradient(135deg, #e6f7ff 0%, #f0f9ff 100%);
  transform: scale(1.02);
}

.upload-area.drag-over .el-icon--upload {
  color: #409eff;
  transform: scale(1.2);
}

.upload-area {
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  background-color: #fafcff;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.upload-area:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.upload-area:hover .el-icon--upload {
  color: #409eff;
  transform: scale(1.1);
}

:deep(.el-upload-dragger) {
  border: none;
  background: transparent;
  width: 100%;
  height: auto;
  padding: 50px 30px;
  border-radius: 8px;
}

:deep(.el-upload-dragger:hover) {
  background: transparent;
}

.upload-content {
  text-align: center;
}

.el-icon--upload {
  font-size: 72px;
  color: #c0c4cc;
  margin-bottom: 20px;
  transition: all 0.3s ease;
}

.el-upload__text {
  color: #606266;
  font-size: 16px;
  margin-bottom: 12px;
  font-weight: 500;
}

.el-upload__text em {
  color: #409eff;
  font-style: normal;
  font-weight: 600;
}

.el-upload__tip {
  color: #909399;
  font-size: 13px;
  line-height: 1.4;
}

.file-preview {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 16px;
}

.file-preview h4 {
  margin: 0 0 12px 0;
  color: #303133;
  font-size: 15px;
  font-weight: 600;
  display: flex;
  align-items: center;
}

.file-preview h4::before {
  content: '📄';
  margin-right: 8px;
  font-size: 16px;
}

.preview-info {
  margin-top: 12px;
  text-align: right;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.error-list {
  background: #fef0f0;
  border: 1px solid #fbc4c4;
  border-radius: 6px;
  padding: 12px;
}

.error-list ul {
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  color: #f56c6c;
}

.dialog-footer {
  text-align: right;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
  margin-top: 20px;
}

:deep(.el-upload-list) {
  margin-top: 16px;
  background: #f8f9fa;
  border-radius: 6px;
  padding: 8px;
}

:deep(.el-upload-list__item) {
  margin-top: 0;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 8px 12px;
}

:deep(.el-upload-list__item:hover) {
  background: #f0f9ff;
  border-color: #409eff;
}

/* 优化Alert组件样式 */
:deep(.el-alert) {
  border-radius: 8px;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

:deep(.el-alert--info) {
  background: linear-gradient(135deg, #e6f7ff 0%, #f0f9ff 100%);
  color: #1890ff;
}

:deep(.el-alert--warning) {
  background: linear-gradient(135deg, #fff7e6 0%, #fffbf0 100%);
  color: #fa8c16;
}

:deep(.el-alert--error) {
  background: linear-gradient(135deg, #fff2f0 0%, #fff1f0 100%);
  color: #f5222d;
}
</style>

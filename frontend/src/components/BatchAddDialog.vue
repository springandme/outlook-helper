<template>
  <el-dialog
    v-model="visible"
    title="批量添加邮箱"
    width="600px"
    :before-close="handleClose"
  >
    <div class="batch-add-content">
      <el-alert
        title="格式说明"
        type="info"
        :closable="false"
        show-icon
      >
        <p>请按以下格式输入邮箱信息，每行一个邮箱：</p>
        <p><strong>格式1：</strong> 邮箱----密码----客户端ID----刷新令牌----备注</p>
        <p><strong>格式2：</strong> 邮箱,密码,客户端ID,刷新令牌,备注</p>
        <p class="note">注：备注为可选项，其他字段必填</p>
      </el-alert>

      <el-alert
        title="数量限制"
        type="warning"
        :closable="false"
        show-icon
        style="margin-top: 12px;"
      >
        <p>为了确保系统稳定性和验证效果，一次性最多只能批量添加30个邮箱账号。</p>
        <p>如需添加更多账号，请分批次进行操作。</p>
      </el-alert>
      
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
        style="margin-top: 20px;"
      >
        <el-form-item label="邮箱数据" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="12"
            placeholder="请输入邮箱数据，每行一个邮箱"
            style="font-family: monospace;"
          />
        </el-form-item>
      </el-form>
      
      <!-- 解析结果预览 -->
      <div v-if="parsedEmails.length > 0" class="preview-section">
        <h4>解析结果预览 ({{ parsedEmails.length }} 个邮箱)</h4>
        <el-table
          :data="parsedEmails.slice(0, 5)"
          style="width: 100%"
          size="small"
          max-height="200"
        >
          <el-table-column prop="email_address" label="邮箱地址" width="200" />
          <el-table-column prop="client_id" label="客户端ID" width="150" />
          <el-table-column prop="remark" label="备注" />
        </el-table>
        <p v-if="parsedEmails.length > 5" class="more-info">
          还有 {{ parsedEmails.length - 5 }} 个邮箱...
        </p>
      </div>
      
      <!-- 解析错误 -->
      <div v-if="parseErrors.length > 0" class="error-section">
        <h4>解析错误 ({{ parseErrors.length }} 个)</h4>
        <el-alert
          v-for="(error, index) in parseErrors.slice(0, 3)"
          :key="index"
          :title="error"
          type="error"
          :closable="false"
          style="margin-bottom: 8px;"
        />
        <p v-if="parseErrors.length > 3" class="more-info">
          还有 {{ parseErrors.length - 3 }} 个错误...
        </p>
      </div>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button @click="parseContent">解析数据</el-button>
        <el-button 
          type="primary" 
          :loading="loading" 
          :disabled="parsedEmails.length === 0"
          @click="handleSubmit"
        >
          {{ loading ? '添加中...' : `添加 ${parsedEmails.length} 个邮箱` }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { emailAPI, type AddEmailRequest, type BatchAddEmailRequest } from '@/api'

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
const formRef = ref<FormInstance>()

const form = reactive({
  content: ''
})

const parsedEmails = ref<AddEmailRequest[]>([])
const parseErrors = ref<string[]>([])

const rules: FormRules = {
  content: [
    { required: true, message: '请输入邮箱数据', trigger: 'blur' }
  ]
}

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    resetForm()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  form.content = ''
  parsedEmails.value = []
  parseErrors.value = []
}

const parseContent = () => {
  parsedEmails.value = []
  parseErrors.value = []
  
  if (!form.content.trim()) {
    ElMessage.warning('请先输入邮箱数据')
    return
  }
  
  const lines = form.content.split('\n')
  
  lines.forEach((line, index) => {
    const trimmedLine = line.trim()
    if (!trimmedLine) return
    
    let parts: string[] = []
    
    // 支持两种分隔符
    if (trimmedLine.includes('----')) {
      parts = trimmedLine.split('----')
    } else if (trimmedLine.includes(',')) {
      parts = trimmedLine.split(',')
    } else {
      parseErrors.value.push(`第${index + 1}行格式错误: ${trimmedLine}`)
      return
    }
    
    if (parts.length < 4) {
      parseErrors.value.push(`第${index + 1}行数据不完整: ${trimmedLine}`)
      return
    }
    
    const email: AddEmailRequest = {
      email_address: parts[0]?.trim() || '',
      password: parts[1]?.trim() || '',
      client_id: parts[2]?.trim() || '',
      refresh_token: parts[3]?.trim() || '',
      remark: parts[4]?.trim() || ''
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

const handleClose = () => {
  visible.value = false
}

const handleSubmit = async () => {
  if (parsedEmails.value.length === 0) {
    ElMessage.warning('请先解析邮箱数据')
    return
  }
  
  try {
    loading.value = true
    const request: BatchAddEmailRequest = {
      emails: parsedEmails.value
    }
    
    const response = await emailAPI.batchAddEmails(request)
    
    if (response.data.success) {
      ElMessage.success('批量添加成功')
      emit('success')
      handleClose()
    } else {
      ElMessage.error(response.data.message || '批量添加失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '批量添加失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.batch-add-content {
  max-height: 600px;
  overflow-y: auto;
}

.note {
  margin: 8px 0 0 0;
  font-size: 12px;
  color: #909399;
}

.preview-section,
.error-section {
  margin-top: 20px;
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 6px;
}

.preview-section h4,
.error-section h4 {
  margin: 0 0 12px 0;
  color: #2c3e50;
  font-size: 14px;
}

.more-info {
  margin: 8px 0 0 0;
  font-size: 12px;
  color: #909399;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-textarea__inner) {
  resize: vertical;
  line-height: 1.4;
}

:deep(.el-alert__content) {
  font-size: 13px;
}
</style>

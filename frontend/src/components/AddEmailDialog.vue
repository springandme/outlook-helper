<template>
  <el-dialog
    v-model="visible"
    title="添加邮箱"
    width="500px"
    :before-close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="邮箱地址" prop="email_address">
        <el-input
          v-model="form.email_address"
          placeholder="请输入邮箱地址"
          clearable
        />
      </el-form-item>
      
      <el-form-item label="密码" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          placeholder="请输入邮箱密码"
          show-password
          clearable
        />
      </el-form-item>
      
      <el-form-item label="客户端ID" prop="client_id">
        <el-input
          v-model="form.client_id"
          placeholder="请输入客户端ID"
          clearable
        />
      </el-form-item>
      
      <el-form-item label="刷新令牌" prop="refresh_token">
        <el-input
          v-model="form.refresh_token"
          type="textarea"
          :rows="3"
          placeholder="请输入刷新令牌"
        />
      </el-form-item>
      
      <el-form-item label="备注">
        <el-input
          v-model="form.remark"
          placeholder="请输入备注信息（可选）"
          clearable
        />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          {{ loading ? '添加中...' : '确定' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
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
const formRef = ref<FormInstance>()

const form = reactive<AddEmailRequest>({
  email_address: '',
  password: '',
  client_id: '',
  refresh_token: '',
  remark: ''
})

const rules: FormRules = {
  email_address: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  client_id: [
    { required: true, message: '请输入客户端ID', trigger: 'blur' }
  ],
  refresh_token: [
    { required: true, message: '请输入刷新令牌', trigger: 'blur' }
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
  Object.assign(form, {
    email_address: '',
    password: '',
    client_id: '',
    refresh_token: '',
    remark: ''
  })
}

const handleClose = () => {
  visible.value = false
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    const valid = await formRef.value.validate()
    if (!valid) return
    
    loading.value = true
    const response = await emailAPI.addEmail(form)
    
    if (response.data.success) {
      ElMessage.success('邮箱添加成功')
      emit('success')
      handleClose()
    } else {
      ElMessage.error(response.data.message || '添加失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '添加失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.dialog-footer {
  text-align: right;
}

:deep(.el-textarea__inner) {
  resize: vertical;
}
</style>

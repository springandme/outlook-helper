<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑标签' : '创建标签'"
    width="400px"
    :before-close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="80px"
    >
      <el-form-item label="标签名称" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入标签名称"
          clearable
        />
      </el-form-item>
      
      <el-form-item label="描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入标签描述（可选）"
        />
      </el-form-item>
      
      <el-form-item label="颜色" prop="color">
        <div class="color-picker-container">
          <el-color-picker
            v-model="form.color"
            :predefine="predefineColors"
            show-alpha
          />
          <el-input
            v-model="form.color"
            placeholder="#000000"
            style="margin-left: 12px; flex: 1;"
          />
        </div>
      </el-form-item>
      
      <el-form-item label="预览">
        <div class="tag-preview">
          <el-tag :color="form.color" class="preview-tag">
            {{ form.name || '标签预览' }}
          </el-tag>
        </div>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          {{ loading ? (isEdit ? '更新中...' : '创建中...') : (isEdit ? '更新' : '创建') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { tagAPI, type Tag, type CreateTagRequest, type UpdateTagRequest } from '@/api'

interface Props {
  modelValue: boolean
  tagData?: Tag | null
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
  name: '',
  description: '',
  color: '#007bff'
})

const isEdit = computed(() => !!props.tagData)

const predefineColors = [
  '#007bff',
  '#28a745',
  '#ffc107',
  '#dc3545',
  '#6c757d',
  '#17a2b8',
  '#e83e8c',
  '#fd7e14',
  '#20c997',
  '#6f42c1'
]

const rules: FormRules = {
  name: [
    { required: true, message: '请输入标签名称', trigger: 'blur' },
    { min: 1, max: 20, message: '标签名称长度在 1 到 20 个字符', trigger: 'blur' }
  ],
  description: [
    { max: 100, message: '描述长度不能超过 100 个字符', trigger: 'blur' }
  ],
  color: [
    { required: true, message: '请选择颜色', trigger: 'change' },
    { pattern: /^#[0-9A-Fa-f]{6}$/, message: '请输入正确的颜色格式', trigger: 'blur' }
  ]
}

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    resetForm()
    if (props.tagData) {
      // 编辑模式，填充数据
      form.name = props.tagData.name
      form.description = props.tagData.description
      form.color = props.tagData.color
    }
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
    name: '',
    description: '',
    color: '#007bff'
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
    
    if (isEdit.value && props.tagData) {
      // 编辑模式
      const updateData: UpdateTagRequest = {
        name: form.name,
        description: form.description,
        color: form.color
      }
      
      const response = await tagAPI.updateTag(props.tagData.id, updateData)
      if (response.data.success) {
        ElMessage.success('标记更新成功')
        emit('success')
        handleClose()
      } else {
        ElMessage.error(response.data.message || '更新失败')
      }
    } else {
      // 创建模式
      const createData: CreateTagRequest = {
        name: form.name,
        description: form.description,
        color: form.color
      }
      
      const response = await tagAPI.createTag(createData)
      if (response.data.success) {
        ElMessage.success('标记创建成功')
        emit('success')
        handleClose()
      } else {
        ElMessage.error(response.data.message || '创建失败')
      }
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.color-picker-container {
  display: flex;
  align-items: center;
  width: 100%;
}

.tag-preview {
  padding: 8px 0;
}

.preview-tag {
  color: white;
  border: none;
  font-size: 14px;
  padding: 6px 12px;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-color-picker__trigger) {
  border-radius: 6px;
}

:deep(.el-textarea__inner) {
  resize: vertical;
}
</style>

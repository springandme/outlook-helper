<template>
  <el-dialog
    v-model="visible"
    title="添加邮箱标签"
    width="400px"
    :before-close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="80px"
    >
      <el-form-item label="选择标签" prop="tag_id">
        <el-select
          v-model="form.tag_id"
          placeholder="请选择标签"
          style="width: 100%"
          filterable
        >
          <el-option
            v-for="tag in tags"
            :key="tag.id"
            :label="tag.name"
            :value="tag.id"
          >
            <div class="tag-option">
              <div class="tag-color" :style="{ backgroundColor: tag.color }"></div>
              <span>{{ tag.name }}</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      
      <el-form-item label="操作类型">
        <el-radio-group v-model="form.action">
          <el-radio label="add">添加标签</el-radio>
          <el-radio label="remove">移除标签</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    
    <div class="info-text">
      将对 {{ emailIds.length }} 个邮箱执行{{ form.action === 'add' ? '添加' : '移除' }}标记操作
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          {{ loading ? '处理中...' : '确定' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { tagAPI, type Tag } from '@/api'

interface Props {
  modelValue: boolean
  emailIds: number[]
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
const tags = ref<Tag[]>([])

const form = reactive({
  tag_id: null as number | null,
  action: 'add' as 'add' | 'remove'
})

const rules: FormRules = {
  tag_id: [
    { required: true, message: '请选择标记', trigger: 'change' }
  ]
}

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    resetForm()
    loadTags()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  form.tag_id = null
  form.action = 'add'
}

const loadTags = async () => {
  try {
    const response = await tagAPI.getTags()
    if (response.data.success) {
      tags.value = response.data.data || []
    }
  } catch (error) {
    console.error('Failed to load tags:', error)
  }
}

const handleClose = () => {
  visible.value = false
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    const valid = await formRef.value.validate()
    if (!valid) return
    
    if (props.emailIds.length === 0) {
      ElMessage.warning('没有选择邮箱')
      return
    }
    
    loading.value = true
    
    const requestData = {
      email_ids: props.emailIds,
      tag_id: form.tag_id!
    }
    
    let response
    if (form.action === 'add') {
      response = await tagAPI.batchTagEmails(requestData)
    } else {
      response = await tagAPI.batchUntagEmails(requestData)
    }
    
    if (response.data.success) {
      ElMessage.success(`批量${form.action === 'add' ? '添加' : '移除'}标记成功`)
      emit('success')
      handleClose()
    } else {
      ElMessage.error(response.data.message || '操作失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.tag-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.info-text {
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 6px;
  color: #606266;
  font-size: 14px;
  margin-top: 16px;
}

.dialog-footer {
  text-align: right;
}
</style>

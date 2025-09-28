<template>
  <el-dialog
    v-model="visible"
    title="导出邮箱数据"
    width="580px"
    :before-close="handleClose"
  >
    <el-form :model="exportForm" label-width="100px" class="export-form">
      <!-- 导出范围 -->
      <el-form-item label="导出范围">
        <el-radio-group v-model="exportForm.range">
          <el-radio label="all">全部邮箱</el-radio>
          <el-radio label="selected" :disabled="selectedCount === 0">
            已选择邮箱 ({{ selectedCount }}个)
          </el-radio>
        </el-radio-group>
      </el-form-item>

      <!-- 导出格式 -->
      <el-form-item label="导出格式">
        <el-radio-group v-model="exportForm.format">
          <el-radio label="txt">TXT格式 (----分隔)</el-radio>
          <el-radio label="csv">CSV格式 (逗号分隔)</el-radio>
        </el-radio-group>
      </el-form-item>

      <!-- 排序字段 -->
      <el-form-item label="排序字段">
        <el-select v-model="exportForm.sortField" style="width: 200px">
          <el-option label="邮箱地址" value="email_address" />
          <el-option label="密码" value="password" />
          <el-option label="RefreshToken" value="refresh_token" />
          <el-option label="ClientID" value="client_id" />
          <el-option label="添加时间" value="created_at" />
        </el-select>
        <el-select v-model="exportForm.sortDirection" style="width: 120px; margin-left: 10px">
          <el-option label="升序" value="asc" />
          <el-option label="降序" value="desc" />
        </el-select>
      </el-form-item>

      <!-- 预览格式 -->
      <el-form-item label="格式预览">
        <el-input
          type="textarea"
          :rows="3"
          :value="formatPreview"
          readonly
          class="preview-area"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="export-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button
          type="primary"
          @click="handleExport"
          :loading="exporting"
        >
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'

// Props
interface Props {
  modelValue: boolean
  selectedCount: number
  selectedEmailIds?: number[]
}

const props = withDefaults(defineProps<Props>(), {
  selectedCount: 0,
  selectedEmailIds: () => []
})

// Emits
interface Emits {
  (event: 'update:modelValue', value: boolean): void
  (event: 'export', params: ExportParams): void
}

const emit = defineEmits<Emits>()

// 导出参数接口
interface ExportParams {
  range: 'all' | 'selected'
  format: 'txt' | 'csv'
  sortField: string
  sortDirection: 'asc' | 'desc'
  emailIds?: number[]
}

// 响应式数据
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const exporting = ref(false)

const exportForm = reactive<ExportParams>({
  range: 'all',
  format: 'txt',
  sortField: 'email_address',
  sortDirection: 'asc'
})

// 格式预览
const formatPreview = computed(() => {
  const sample = {
    email: 'example@outlook.com',
    password: 'password123',
    refreshToken: 'refresh_token_here',
    clientId: 'client_id_here'
  }

  if (exportForm.format === 'txt') {
    return `${sample.email}----${sample.password}----${sample.refreshToken}----${sample.clientId}`
  } else {
    return `"${sample.email}","${sample.password}","${sample.refreshToken}","${sample.clientId}"`
  }
})

// 监听选中数量变化，自动调整导出范围
watch(() => props.selectedCount, (newCount) => {
  if (newCount === 0 && exportForm.range === 'selected') {
    exportForm.range = 'all'
  }
})

// 方法
const handleClose = () => {
  visible.value = false
}

const handleExport = async () => {
  try {
    exporting.value = true

    const params: ExportParams = {
      range: exportForm.range,
      format: exportForm.format,
      sortField: exportForm.sortField,
      sortDirection: exportForm.sortDirection
    }

    // 如果导出选中的邮箱，需要传递邮箱ID列表
    if (exportForm.range === 'selected') {
      params.emailIds = props.selectedEmailIds
    }

    emit('export', params)

    // 导出成功后关闭对话框
    visible.value = false

  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败，请重试')
  } finally {
    exporting.value = false
  }
}

// 重置表单
const resetForm = () => {
  exportForm.range = 'all'
  exportForm.format = 'txt'
  exportForm.sortField = 'email_address'
  exportForm.sortDirection = 'asc'
}

// 监听对话框打开，重置表单
watch(visible, (newVisible) => {
  if (newVisible) {
    resetForm()
    // 如果有选中的邮箱，默认选择导出选中的
    if (props.selectedCount > 0) {
      exportForm.range = 'selected'
    }
  }
})
</script>

<style scoped>
.export-form {
  padding: 0 20px;
}

.export-form .el-form-item {
  margin-bottom: 20px;
}

.preview-area {
  background-color: #f5f7fa;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.preview-area :deep(.el-textarea__inner) {
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  color: #606266;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.export-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.el-radio-group .el-radio {
  margin-right: 30px;
}

.el-form-item__label {
  font-weight: 500;
  color: #303133;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .export-form {
    padding: 0 10px;
  }

  .el-select {
    width: 100% !important;
    margin-left: 0 !important;
    margin-top: 10px;
  }
}
</style>
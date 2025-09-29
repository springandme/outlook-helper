<template>
  <el-dialog
    v-model="visible"
    title="导出邮箱数据"
    width="650px"
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

      <!-- 字段顺序 -->
      <el-form-item label="字段顺序">
        <div class="field-order-container">
          <div class="field-order-hint">拖拽调整字段顺序：</div>
          <VueDraggable
            v-model="exportForm.fieldOrder"
            class="field-list"
            item-key="key"
            :animation="150"
            tag="div"
          >
            <template #item="{ element, index }">
              <div class="field-item">
                <el-icon class="drag-handle"><Rank /></el-icon>
                <span class="field-label">{{ element.label }}</span>
                <span class="field-position">{{ index + 1 }}</span>
              </div>
            </template>
          </VueDraggable>
        </div>
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
import { Download, Rank } from '@element-plus/icons-vue'
import VueDraggable from 'vuedraggable'

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

// 字段定义接口
interface FieldOption {
  key: string
  label: string
  value: string
}

// 导出参数接口
interface ExportParams {
  range: 'all' | 'selected'
  format: 'txt' | 'csv'
  fieldOrder: FieldOption[]
  emailIds?: number[]
}

// 响应式数据
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const exporting = ref(false)

// 默认字段顺序
const defaultFieldOrder: FieldOption[] = [
  { key: 'email_address', label: '邮箱地址', value: 'email_address' },
  { key: 'password', label: '密码', value: 'password' },
  { key: 'refresh_token', label: 'RefreshToken', value: 'refresh_token' },
  { key: 'client_id', label: 'ClientID', value: 'client_id' }
]

const exportForm = reactive<ExportParams>({
  range: 'all',
  format: 'txt',
  fieldOrder: [...defaultFieldOrder]
})

// 格式预览
const formatPreview = computed(() => {
  const sampleData: Record<string, string> = {
    email_address: 'example@outlook.com',
    password: 'password123',
    refresh_token: 'refresh_token_here',
    client_id: 'client_id_here'
  }

  const orderedValues = exportForm.fieldOrder.map(field => sampleData[field.key])

  if (exportForm.format === 'txt') {
    return orderedValues.join('----')
  } else {
    return orderedValues.map(val => `"${val}"`).join(',')
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
      fieldOrder: exportForm.fieldOrder
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
  exportForm.fieldOrder = [...defaultFieldOrder]
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

/* 字段顺序相关样式 */
.field-order-container {
  width: 100%;
}

.field-order-hint {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.field-list {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #fafafa;
  padding: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.field-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  margin: 4px 0;
  background-color: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  cursor: move;
  transition: all 0.2s;
}

.field-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.12);
}

.field-item.sortable-chosen {
  border-color: #409eff;
  background-color: #ecf5ff;
}

.field-item.sortable-ghost {
  opacity: 0.5;
  background-color: #f0f9ff;
}

.drag-handle {
  margin-right: 8px;
  color: #909399;
  cursor: grab;
}

.drag-handle:active {
  cursor: grabbing;
}

.field-label {
  flex: 1;
  font-size: 14px;
  color: #303133;
}

.field-position {
  background-color: #409eff;
  color: #fff;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 500;
  min-width: 20px;
  text-align: center;
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
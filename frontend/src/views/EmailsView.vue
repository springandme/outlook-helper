<template>
  <div class="emails-page">
    <!-- 操作栏 -->
    <el-card class="operation-card">
      <div class="operation-bar">
        <div class="operation-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索邮箱地址或备注"
            style="width: 300px;"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <div class="operation-right">
          <el-button
            type="danger"
            :disabled="selectedEmails.length === 0 || operationLoading.batchClearInbox"
            :loading="operationLoading.batchClearInbox"
            @click="batchClearInbox"
          >
            <el-icon><Delete /></el-icon>
            清空收件箱
          </el-button>
          <el-button
            type="danger"
            plain
            :disabled="selectedEmails.length === 0 || operationLoading.batchDelete"
            :loading="operationLoading.batchDelete"
            @click="batchDelete"
          >
            <el-icon><Delete /></el-icon>
            批量删除
          </el-button>
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            添加邮箱
          </el-button>
          <el-button @click="showBatchDialog = true">
            <el-icon><Upload /></el-icon>
            批量添加
          </el-button>
          <el-button @click="showImportDialog = true">
            <el-icon><Document /></el-icon>
            导入文件
          </el-button>
          <el-button @click="loadEmails">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 邮箱列表 -->
    <el-card class="table-card" v-loading="pageLoading" :element-loading-text="pageLoadingText">
      <el-table
        v-loading="loading"
        :data="emails"
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="email_address" label="邮箱地址" width="320">
          <template #default="{ row }">
            <div class="email-address-cell">
              <span
                class="email-address-link"
                @click="handleEmailClick(row)"
                title="点击标记邮箱"
              >
                {{ row.email_address }}
              </span>
              <el-button
                size="small"
                type="text"
                class="copy-button"
                @click.stop="copyEmailAddress(row.email_address)"
                title="复制邮箱地址"
              >
                <el-icon size="12"><CopyDocument /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="130">
          <template #default="{ row }">
            <div v-if="row.tags && row.tags.length > 0" class="tags-container">
              <el-tag
                v-for="tag in row.tags"
                :key="tag.id"
                :color="tag.color"
                size="small"
                class="tag-item"
              >
                {{ tag.name }}
              </el-tag>
            </div>
            <span v-else class="no-tags">-</span>
            </template>
          </el-table-column>
        <el-table-column prop="remark" label="备注" width="100">
          <template #default="{ row }">
            <el-tooltip
              :content="row.remark || '无备注'"
              placement="top"
              :disabled="!row.remark || row.remark.length <= 10"
            >
              <span class="remark-text">{{ row.remark || '-' }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
          <el-table-column prop="created_at" label="添加时间" width="160">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="last_operation_at" label="最后操作" width="160">
            <template #default="{ row }">
              {{ row.last_operation_at ? formatTime(row.last_operation_at) : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <div class="operation-buttons">
                <el-button
                  link
                  class="icon-button icon-button-primary"
                  :loading="operationLoading.getLatestMail[row.id]"
                  :disabled="operationLoading.getLatestMail[row.id]"
                  @click="getLatestMail(row)"
                >
                  <el-icon :size="18"><Bell /></el-icon>
                </el-button>

                <el-button
                  link
                  class="icon-button icon-button-info"
                  :loading="operationLoading.getAllMails[row.id]"
                  :disabled="operationLoading.getAllMails[row.id]"
                  @click="getAllMails(row)"
                >
                  <el-icon :size="18"><Grid /></el-icon>
                </el-button>

                <el-button
                  link
                  class="icon-button icon-button-danger"
                  :loading="operationLoading.deleteEmail[row.id]"
                  :disabled="operationLoading.deleteEmail[row.id]"
                  @click="deleteEmail(row)"
                >
                  <el-icon :size="18"><Close /></el-icon>
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </el-card>




    <!-- 添加邮箱对话框 -->
    <AddEmailDialog
      v-model="showAddDialog"
      @success="handleAddSuccess"
    />

    <!-- 批量添加对话框 -->
    <BatchAddDialog
      v-model="showBatchDialog"
      @success="handleBatchSuccess"
    />

    <!-- 批量标签对话框 -->
    <BatchTagDialog
      v-model="showBatchTagDialog"
      :email-ids="selectedEmails.map(e => e.id)"
      @success="handleBatchTagSuccess"
    />

    <!-- 邮件查看对话框 -->
    <MailViewDialog
      v-model="showMailDialog"
      :mail-data="currentMail"
      :mail-list="currentMailList"
    />

    <!-- 文件导入对话框 -->
    <FileImportDialog
      v-model="showImportDialog"
      @success="handleImportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Upload,
  Document,
  Search,
  Refresh,
  Message,
  Folder,
  Files,
  Delete,
  DeleteFilled,
  ArrowDown,
  Collection,
  CopyDocument,
  Promotion,
  FolderOpened,
  Bell,
  Grid,
  Close
} from '@element-plus/icons-vue'
import AddEmailDialog from '@/components/AddEmailDialog.vue'
import BatchAddDialog from '@/components/BatchAddDialog.vue'
import BatchTagDialog from '@/components/BatchTagDialog.vue'
import MailViewDialog from '@/components/MailViewDialog.vue'
import FileImportDialog from '@/components/FileImportDialog.vue'
import { emailAPI, type Email, type OutlookMail } from '@/api'

// 响应式数据
const loading = ref(false)
const emails = ref<Email[]>([])
const selectedEmails = ref<Email[]>([])
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 操作loading状态
const operationLoading = ref({
  getLatestMail: {} as Record<number, boolean>,
  getAllMails: {} as Record<number, boolean>,
  deleteEmail: {} as Record<number, boolean>,
  batchClearInbox: false,
  batchDelete: false
})

// 对话框状态
const showAddDialog = ref(false)
const showBatchDialog = ref(false)
const showBatchTagDialog = ref(false)
const showMailDialog = ref(false)
const showImportDialog = ref(false)

// 全页面加载遮罩
const pageLoading = ref(false)
const pageLoadingText = ref('加载中...')

// 邮件相关
const currentMail = ref<OutlookMail | null>(null)
const currentMailList = ref<OutlookMail[]>([])

// 方法
const loadEmails = async () => {
  loading.value = true
  try {
    const params = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      keyword: searchKeyword.value || undefined
    }
    
    const response = await emailAPI.getEmails(params)
    if (response.data.success) {
      const data = response.data.data as any
      emails.value = data.list || []
      total.value = data.total || 0
    } else {
      ElMessage.error(response.data.message || '获取邮箱列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '获取邮箱列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadEmails()
}

const handleSelectionChange = (selection: Email[]) => {
  selectedEmails.value = selection
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadEmails()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  loadEmails()
}

const handleImportSuccess = () => {
  loadEmails()
}

const copyEmailAddress = async (emailAddress: string) => {
  try {
    await navigator.clipboard.writeText(emailAddress)
    ElMessage.success('邮箱地址已复制到剪贴板')
  } catch (error) {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = emailAddress
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    ElMessage.success('邮箱地址已复制到剪贴板')
  }
}

const getLatestMail = async (email: Email) => {
  operationLoading.value.getLatestMail[email.id] = true
  pageLoading.value = true
  pageLoadingText.value = '正在获取最新邮件...'
  try {
    const response = await emailAPI.getLatestMail(email.id)
    if (response.data.success && response.data.data) {
      currentMail.value = response.data.data
      currentMailList.value = [response.data.data]
      showMailDialog.value = true
    } else {
      // 检查是否是"Nothing to fetch"错误
      if (response.data.error && response.data.error.includes('Nothing to fetch')) {
        ElMessage.info('当前邮箱暂无邮件')
      } else {
        ElMessage.error(response.data.message || '获取邮件失败')
      }
    }
  } catch (error: any) {
    // 检查错误响应中是否包含"Nothing to fetch"
    const errorMessage = error.response?.data?.error || error.response?.data?.message || ''
    if (errorMessage.includes('Nothing to fetch')) {
      ElMessage.info('当前邮箱暂无邮件')
    } else {
      ElMessage.error(error.response?.data?.message || '获取邮件失败')
    }
  } finally {
    operationLoading.value.getLatestMail[email.id] = false
    pageLoading.value = false
  }
}

const getAllMails = async (email: Email) => {
  operationLoading.value.getAllMails[email.id] = true
  pageLoading.value = true
  pageLoadingText.value = '正在获取全部邮件...'
  try {
    const response = await emailAPI.getAllMails(email.id)
    if (response.data.success && response.data.data) {
      currentMailList.value = response.data.data
      currentMail.value = response.data.data[0] || null
      showMailDialog.value = true
    } else {
      // 检查是否是"Nothing to fetch"错误
      if (response.data.error && response.data.error.includes('Nothing to fetch')) {
        ElMessage.info('当前邮箱暂无邮件')
      } else {
        ElMessage.error(response.data.message || '获取邮件失败')
      }
    }
  } catch (error: any) {
    // 检查错误响应中是否包含"Nothing to fetch"
    const errorMessage = error.response?.data?.error || error.response?.data?.message || ''
    if (errorMessage.includes('Nothing to fetch')) {
      ElMessage.info('当前邮箱暂无邮件')
    } else {
      ElMessage.error(error.response?.data?.message || '获取邮件失败')
    }
  } finally {
    operationLoading.value.getAllMails[email.id] = false
    pageLoading.value = false
  }
}

const clearInbox = async (email: Email) => {
  try {
    await ElMessageBox.confirm(
      `确定要清空邮箱 ${email.email_address} 的收件箱吗？此操作不可恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await emailAPI.clearInbox(email.id)
    if (response.data.success) {
      ElMessage.success('收件箱清空成功')
      loadEmails()
    } else {
      ElMessage.error(response.data.message || '清空收件箱失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '清空收件箱失败')
    }
  }
}

const batchClearInbox = async () => {
  if (selectedEmails.value.length === 0) {
    ElMessage.warning('请先选择要清空收件箱的邮箱')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要清空选中的 ${selectedEmails.value.length} 个邮箱的收件箱吗？此操作不可恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operationLoading.value.batchClearInbox = true
    const emailIds = selectedEmails.value.map(email => email.id)
    const response = await emailAPI.batchClearInbox(emailIds)

    if (response.data.success) {
      const data = response.data.data as any
      if (data.error_count > 0) {
        ElMessage.warning(`批量清空收件箱完成，成功: ${data.success_count}, 失败: ${data.error_count}`)
      } else {
        ElMessage.success(`成功清空 ${data.success_count} 个邮箱的收件箱`)
      }
      selectedEmails.value = []
      loadEmails()
    } else {
      ElMessage.error(response.data.message || '批量清空收件箱失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '批量清空收件箱失败')
    }
  } finally {
    operationLoading.value.batchClearInbox = false
  }
}

const handleDropdownCommand = (command: string, email: Email) => {
  if (command === 'tag') {
    selectedEmails.value = [email]
    showBatchTagDialog.value = true
  } else if (command === 'delete') {
    deleteEmail(email)
  }
}

const deleteEmail = async (email: Email) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除邮箱 ${email.email_address} 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operationLoading.value.deleteEmail[email.id] = true
    pageLoading.value = true
    pageLoadingText.value = '正在删除邮箱...'
    const response = await emailAPI.deleteEmail(email.id)
    if (response.data.success) {
      ElMessage.success('删除成功')
      loadEmails()
    } else {
      ElMessage.error(response.data.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  } finally {
    operationLoading.value.deleteEmail[email.id] = false
    pageLoading.value = false
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedEmails.value.length} 个邮箱吗？`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    operationLoading.value.batchDelete = true
    // 批量删除
    const emailIds = selectedEmails.value.map(email => email.id)
    await emailAPI.batchDeleteEmails(emailIds)

    ElMessage.success('批量删除成功')
    selectedEmails.value = []
    loadEmails()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '批量删除失败')
    }
  } finally {
    operationLoading.value.batchDelete = false
  }
}

const handleAddSuccess = () => {
  loadEmails()
}

const handleBatchSuccess = () => {
  loadEmails()
}

const handleBatchTagSuccess = () => {
  selectedEmails.value = []
  loadEmails()
}

const handleEmailClick = (email: Email) => {
  // 设置选中的邮箱为当前点击的邮箱
  selectedEmails.value = [email]
  // 打开批量标签对话框
  showBatchTagDialog.value = true
}

const formatTime = (timeStr: string): string => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadEmails()
})
</script>

<style scoped>
.emails-page {
  width: 100%;
  max-width: 100%;
  margin: 0 auto;
  padding: 0 16px;
}

@media (min-width: 1200px) {
  .emails-page {
    max-width: 1800px;
  }
}

.operation-card {
  margin-bottom: 20px;
}

.operation-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.operation-left {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.operation-right {
  display: flex;
  gap: 12px;
  align-items: center;
}

.table-card {
  margin-bottom: 20px;
}

.tags-container {
  display: flex;
  flex-wrap: nowrap;
  gap: 2px;
  overflow: hidden;
  align-items: center;
}

.tag-item {
  color: white;
  border: none;
  height: 22px;
  line-height: 20px;
  padding: 0 7px;
  font-size: 12px;
}

.tags-container .el-tag {
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  height: 22px;
  line-height: 20px;
  padding: 0 7px;
  font-size: 12px;
}

.no-tags {
  color: #999;
}



:deep(.el-table .cell) {
  padding: 4px 8px;
  font-size: 14px;
}

:deep(.el-table td) {
  height: 36px;
}

:deep(.el-table th) {
  height: 40px;
  padding: 8px 0;
}

:deep(.el-table .el-table__row) {
  height: 36px;
}

:deep(.el-button-group .el-button) {
  margin: 0;
}

.operation-buttons {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: nowrap;
  height: 28px;
}

.operation-buttons .el-button {
  margin: 0;
}

.operation-buttons .icon-button {
  padding: 3px;
  border-radius: 4px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  background: transparent;
  position: relative;
  border: 1px solid transparent;
  height: 26px;
  width: 26px;
}

.operation-buttons .icon-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.operation-buttons .icon-button:active {
  transform: translateY(0);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

/* 主要操作按钮 - 最新邮件（紫色） */
.operation-buttons .icon-button-primary .el-icon {
  color: #9c27b0;
}

.operation-buttons .icon-button-primary:hover {
  background: rgba(156, 39, 176, 0.1);
  border-color: rgba(156, 39, 176, 0.2);
}

.operation-buttons .icon-button-primary:hover .el-icon {
  color: #7b1fa2;
}

/* 信息操作按钮 - 全部邮件（黑色） */
.operation-buttons .icon-button-info .el-icon {
  color: #303133;
}

.operation-buttons .icon-button-info:hover {
  background: rgba(48, 49, 51, 0.08);
  border-color: rgba(48, 49, 51, 0.15);
}

.operation-buttons .icon-button-info:hover .el-icon {
  color: #000000;
}

/* 危险操作按钮 - 删除 */
.operation-buttons .icon-button-danger .el-icon {
  color: #f56c6c;
}

.operation-buttons .icon-button-danger:hover {
  background: rgba(245, 108, 108, 0.1);
  border-color: rgba(245, 108, 108, 0.2);
}

.operation-buttons .icon-button-danger:hover .el-icon {
  color: #f23c3c;
}

/* 禁用状态 */
.operation-buttons .icon-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.operation-buttons .icon-button:disabled:hover {
  background: transparent;
  transform: none;
  box-shadow: none;
  border-color: transparent;
}

.email-address-cell {
  display: flex;
  align-items: center;
  width: 100%;
  white-space: nowrap;
  overflow: hidden;
  height: 28px;
}

.email-address-link {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
  color: #409eff;
  transition: all 0.2s ease;
  padding: 2px 4px;
  border-radius: 2px;
  font-size: 14px;
  line-height: 20px;
}

.email-address-link:hover {
  background-color: #f0f9ff;
  color: #66b1ff;
  text-decoration: underline;
}

.email-address-link:active {
  background-color: #e6f7ff;
  color: #409eff;
}

.copy-button {
  margin-left: 4px;
  padding: 1px;
  opacity: 0.6;
  transition: all 0.2s ease;
  border-radius: 2px;
  background: transparent;
  min-height: auto;
  height: 16px;
  width: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  vertical-align: middle;
  flex-shrink: 0;
}

.copy-button:hover {
  opacity: 1;
  background: #f0f9ff;
  color: #409eff;
  transform: scale(1.05);
}

/* 确保复制按钮不会影响表格行的点击 */
.copy-button:focus {
  outline: none;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.remark-text {
  display: inline-block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 自定义加载遮罩样式 - 仅限于卡片区域 */
.table-card :deep(.el-loading-mask) {
  background-color: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(3px);
  transition: opacity 0.3s ease;
  border-radius: 4px;
}

.table-card :deep(.el-loading-spinner) {
  top: 50%;
  transform: translateY(-50%);
}

.table-card :deep(.el-loading-spinner .el-loading-text) {
  color: #409eff;
  font-size: 14px;
  margin-top: 10px;
  font-weight: 500;
}

.table-card :deep(.el-loading-spinner .circular) {
  width: 42px;
  height: 42px;
  animation: loading-rotate 2s linear infinite;
}

@keyframes loading-rotate {
  100% {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .operation-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .operation-right {
    justify-content: space-between;
  }

  .operation-right .el-input {
    flex: 1;
    margin-right: 12px;
  }
}
</style>

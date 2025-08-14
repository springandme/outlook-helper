<template>
  <div class="dashboard">
    <div class="dashboard-content">
        <!-- 统计卡片 -->
        <el-row :gutter="20" class="stats-cards">
          <el-col :xs="24" :sm="12" :md="6">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon email">
                  <el-icon size="24"><Message /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ dashboardData?.total_emails || 0 }}</div>
                  <div class="stat-label">邮箱总数</div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="6">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon tag">
                  <el-icon size="24"><Collection /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ dashboardData?.total_tags || 0 }}</div>
                  <div class="stat-label">标签总数</div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="6">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon operation">
                  <el-icon size="24"><Operation /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ recentOperationsCount }}</div>
                  <div class="stat-label">今日操作</div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <el-col :xs="24" :sm="12" :md="6">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon active">
                  <el-icon size="24"><CircleCheck /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ activeEmailsCount }}</div>
                  <div class="stat-label">活跃邮箱</div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        
        <!-- 图表和列表 -->
        <el-row :gutter="20" class="charts-section">
          <!-- 邮箱分布图表 -->
          <el-col :xs="24" :lg="12">
            <el-card class="chart-card">
              <template #header>
                <div class="card-header">
                  <span>邮箱标签分布</span>
                  <el-button type="text" @click="refreshData">
                    <el-icon><Refresh /></el-icon>
                  </el-button>
                </div>
              </template>
              <div class="chart-content">
                <div v-if="emailsByTagList.length === 0" class="empty-chart">
                  <el-empty description="暂无数据" />
                </div>
                <div v-else class="tag-stats">
                  <div
                    v-for="item in emailsByTagList"
                    :key="item.name"
                    class="tag-stat-item"
                  >
                    <div class="tag-info">
                      <div class="tag-color" :style="{ backgroundColor: item.color }"></div>
                      <span class="tag-name">{{ item.name }}</span>
                    </div>
                    <div class="tag-count">{{ item.count }}</div>
                  </div>
                </div>
              </div>
            </el-card>
          </el-col>
          
          <!-- 操作统计图表 -->
          <el-col :xs="24" :lg="12">
            <el-card class="chart-card">
              <template #header>
                <div class="card-header">
                  <span>操作类型统计</span>
                </div>
              </template>
              <div class="chart-content">
                <div v-if="operationsList.length === 0" class="empty-chart">
                  <el-empty description="暂无数据" />
                </div>
                <div v-else class="operation-stats">
                  <div
                    v-for="item in operationsList"
                    :key="item.type"
                    class="operation-stat-item"
                  >
                    <div class="operation-info">
                      <span class="operation-name">{{ item.type }}</span>
                      <span class="operation-count">{{ item.count }}</span>
                    </div>
                    <div class="operation-bar">
                      <div
                        class="operation-progress"
                        :style="{ width: `${(item.count / maxOperationCount) * 100}%` }"
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        
        <!-- 最近操作记录 -->
        <el-row>
          <el-col :span="24">
            <el-card class="operations-card">
              <template #header>
                <div class="card-header">
                  <span>最近操作记录</span>
                  <div class="header-actions">
                    <el-button
                      type="danger"
                      size="small"
                      :icon="Delete"
                      @click="clearAllLogs"
                      :loading="logsLoading"
                    >
                      清空
                    </el-button>
                    <el-button
                      type="primary"
                      size="small"
                      :icon="Refresh"
                      @click="loadOperationLogs"
                      :loading="logsLoading"
                    >
                      刷新
                    </el-button>
                  </div>
                </div>
              </template>
              <div class="operations-content">
                <el-table
                  :data="operationLogs"
                  style="width: 100%"
                  :show-header="true"
                  empty-text="暂无操作记录"
                  v-loading="logsLoading"
                >
                  <el-table-column prop="operation_type" label="操作类型" width="120">
                    <template #default="{ row }">
                      <el-tag :type="getOperationTagType(row.operation_type)" size="small">
                        {{ row.operation_name || getOperationName(row.operation_type) }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="description" label="操作描述" />
                  <el-table-column prop="created_at" label="操作时间" width="180">
                    <template #default="{ row }">
                      {{ formatTime(row.created_at) }}
                    </template>
                  </el-table-column>
                </el-table>

                <!-- 分页组件 -->
                <div class="pagination-wrapper" v-if="total > 0">
                  <el-pagination
                    v-model:current-page="currentPage"
                    :page-size="pageSize"
                    :total="total"
                    layout="prev, pager, next, jumper"
                    @current-change="handlePageChange"
                  />
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Message,
  Collection,
  Operation,
  CircleCheck,
  Refresh,
  Delete
} from '@element-plus/icons-vue'
import { dashboardAPI, logsAPI, type DashboardStats, type OperationLog } from '@/api'

const loading = ref(false)
const dashboardData = ref<DashboardStats | null>(null)

// 操作日志分页数据
const logsLoading = ref(false)
const operationLogs = ref<OperationLog[]>([])
const currentPage = ref(1)
const pageSize = ref(5)
const total = ref(0)
const totalPages = ref(0)

// 计算属性
const recentOperationsCount = computed(() => {
  if (!dashboardData.value?.recent_operations) return 0
  const today = new Date().toDateString()
  return dashboardData.value.recent_operations.filter(op => 
    new Date(op.created_at).toDateString() === today
  ).length
})

const activeEmailsCount = computed(() => {
  // 这里可以根据实际需求计算活跃邮箱数量
  // 暂时使用邮箱总数
  return dashboardData.value?.total_emails || 0
})

const emailsByTagList = computed(() => {
  if (!dashboardData.value?.emails_by_tag) return []
  return Object.entries(dashboardData.value.emails_by_tag).map(([name, count]) => ({
    name,
    count,
    color: getTagColor(name)
  }))
})

const operationsList = computed(() => {
  if (!dashboardData.value?.operations_by_type) return []
  return Object.entries(dashboardData.value.operations_by_type)
    .map(([type, count]) => ({ type, count }))
    .sort((a, b) => b.count - a.count)
})

const maxOperationCount = computed(() => {
  return Math.max(...operationsList.value.map(item => item.count), 1)
})

// 方法
const loadDashboardData = async () => {
  loading.value = true
  try {
    const response = await dashboardAPI.getDashboard()
    if (response.data.success) {
      dashboardData.value = response.data.data
    } else {
      ElMessage.error(response.data.message || '获取仪表盘数据失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '获取仪表盘数据失败')
  } finally {
    loading.value = false
  }
}

// 加载操作日志
const loadOperationLogs = async () => {
  logsLoading.value = true
  try {
    const response = await logsAPI.getLogs(currentPage.value, pageSize.value)
    if (response.data.success) {
      const data = response.data.data
      operationLogs.value = data.logs
      total.value = data.total
      totalPages.value = data.total_pages
    } else {
      ElMessage.error(response.data.message || '获取操作日志失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '获取操作日志失败')
  } finally {
    logsLoading.value = false
  }
}

// 处理分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  loadOperationLogs()
}

// 清空所有操作日志
const clearAllLogs = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有操作记录吗？此操作不可恢复。',
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    logsLoading.value = true
    const response = await logsAPI.clearLogs()
    if (response.data.success) {
      ElMessage.success('操作记录已清空')
      // 重置分页并重新加载
      currentPage.value = 1
      await loadOperationLogs()
      // 同时刷新仪表盘数据
      await loadDashboardData()
    } else {
      ElMessage.error(response.data.message || '清空操作记录失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '清空操作记录失败')
    }
  } finally {
    logsLoading.value = false
  }
}

const refreshData = () => {
  loadDashboardData()
  loadOperationLogs()
}



const getOperationName = (type: string): string => {
  const operationNames: Record<string, string> = {
    // 认证相关
    'login_success': '登录成功',
    'login_failed': '登录失败',
    'logout': '退出登录',
    'clear_all_logs': '清空操作日志',

    // 邮箱相关
    'email_added': '添加邮箱',
    'email_deleted': '删除邮箱',
    'email_updated': '更新邮箱',
    'email_validation_failed': '邮箱验证失败',
    'batch_add_emails': '批量添加邮箱',
    'batch_delete_emails': '批量删除邮箱',

    // 邮件操作相关
    'get_latest_mail': '获取最新邮件',
    'get_latest_mail_failed': '获取邮件失败',
    'get_all_mails': '获取全部邮件',
    'get_all_mails_failed': '获取邮件失败',
    'clear_inbox': '清空收件箱',
    'clear_inbox_failed': '清空收件箱失败',
    'clear_junk': '清空垃圾箱',
    'clear_junk_failed': '清空垃圾箱失败',

    // 标签相关
    'tag_created': '创建标签',
    'tag_updated': '更新标签',
    'tag_deleted': '删除标签',
    'email_tagged': '添加标签',
    'email_untagged': '移除标签',
    'batch_tag_emails': '批量添加标签',
    'batch_untag_emails': '批量移除标签'
  }
  return operationNames[type] || type
}

const getOperationTagType = (type: string): string => {
  const tagTypes: Record<string, string> = {
    // 认证相关
    'login_success': 'success',
    'login_failed': 'danger',
    'logout': 'info',
    'clear_all_logs': 'warning',

    // 邮箱相关
    'email_added': 'success',
    'email_deleted': 'danger',
    'email_updated': 'warning',
    'email_validation_failed': 'danger',
    'batch_add_emails': 'success',
    'batch_delete_emails': 'danger',

    // 邮件操作相关
    'get_latest_mail': 'primary',
    'get_latest_mail_failed': 'danger',
    'get_all_mails': 'primary',
    'get_all_mails_failed': 'danger',
    'clear_inbox': 'warning',
    'clear_inbox_failed': 'danger',
    'clear_junk': 'warning',
    'clear_junk_failed': 'danger',

    // 标签相关
    'tag_created': 'success',
    'tag_updated': 'warning',
    'tag_deleted': 'danger',
    'email_tagged': 'success',
    'email_untagged': 'warning',
    'batch_tag_emails': 'success',
    'batch_untag_emails': 'warning'
  }
  return tagTypes[type] || 'info'
}

const getTagColor = (tagName: string): string => {
  const colors = ['#007bff', '#28a745', '#ffc107', '#dc3545', '#6c757d', '#17a2b8']
  const index = tagName.length % colors.length
  return colors[index]
}

// 格式化时间
const formatTime = (timeStr: string): string => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadDashboardData()
  loadOperationLogs()
})
</script>

<style scoped>
.dashboard-content {
  width: 100%;
  max-width: 100%;
  margin: 0 auto;
  padding: 0 16px;
}

@media (min-width: 1200px) {
  .dashboard-content {
    max-width: 1800px;
  }
}

.stats-cards {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
}

.stat-item {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
}

.stat-icon.email {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.tag {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.operation {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.active {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #7f8c8d;
  margin-top: 4px;
}

.charts-section {
  margin-bottom: 20px;
}

.chart-card,
.operations-card {
  height: 400px;
}

.chart-card .el-card__body,
.operations-card .el-card__body {
  height: calc(100% - 60px); /* 减去header高度 */
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.chart-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow-y: auto;
}

.empty-chart {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tag-stats,
.operation-stats {
  width: 100%;
  padding: 20px 0;
  max-height: 100%;
  overflow-y: auto;
}

.tag-stat-item,
.operation-stat-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.tag-stat-item:last-child,
.operation-stat-item:last-child {
  border-bottom: none;
}

.tag-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.tag-name,
.operation-name {
  font-size: 14px;
  color: #2c3e50;
}

.tag-count,
.operation-count {
  font-weight: 600;
  color: #2c3e50;
}

.operation-stat-item {
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
}

.operation-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.operation-bar {
  height: 6px;
  background-color: #f0f0f0;
  border-radius: 3px;
  overflow: hidden;
}

.operation-progress {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s ease;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table .cell) {
  padding: 0 8px;
}

.operations-content {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

:deep(.pagination-wrapper .el-pagination) {
  font-size: 14px;
}

:deep(.pagination-wrapper .el-pager li) {
  font-size: 14px;
  min-width: 32px;
  height: 32px;
  line-height: 30px;
}

:deep(.pagination-wrapper .el-pagination button) {
  font-size: 14px;
  height: 32px;
  line-height: 30px;
}
</style>

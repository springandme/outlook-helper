<template>
  <div class="tags-page">
    <!-- 操作栏 -->
    <el-card class="operation-card">
      <div class="operation-bar">
        <div class="operation-left">
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            创建标签
          </el-button>
        </div>
        <div class="operation-right">
          <el-button @click="loadTags">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 标签列表 -->
    <el-row :gutter="20">
      <el-col
        v-for="tag in tags"
        :key="tag.id"
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
        :xl="4"
        style="margin-bottom: 20px;"
      >
        <el-card class="tag-card" :body-style="{ padding: '20px' }">
          <div class="tag-content">
            <div class="tag-header">
              <div class="tag-color" :style="{ backgroundColor: tag.color }"></div>
              <h3 class="tag-name">{{ tag.name }}</h3>
            </div>

            <p class="tag-description">{{ tag.description || '暂无描述' }}</p>

            <div class="tag-stats">
              <div class="stat-item">
                <span class="stat-label">关联邮箱：</span>
                <span class="stat-value">{{ tag.email_count || 0 }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">创建时间：</span>
                <span class="stat-value">{{ formatDate(tag.created_at) }}</span>
              </div>
            </div>

            <div class="tag-actions">
              <el-button size="small" @click="editTag(tag)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                size="small"
                type="danger"
                :disabled="(tag.email_count || 0) > 0"
                @click="deleteTag(tag)"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 空状态 -->
      <el-col v-if="tags.length === 0 && !loading" :span="24">
        <el-card class="empty-card">
          <el-empty description="暂无标签">
            <el-button type="primary" @click="showCreateDialog = true">
              创建第一个标签
            </el-button>
          </el-empty>
      </el-card>
      </el-col>
    </el-row>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 创建/编辑标签对话框 -->
    <TagFormDialog
      v-model="showCreateDialog"
      :tag-data="editingTag"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Refresh,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import TagFormDialog from '@/components/TagFormDialog.vue'
import { tagAPI, type Tag } from '@/api'

// 响应式数据
const loading = ref(false)
const tags = ref<Tag[]>([])
const showCreateDialog = ref(false)
const editingTag = ref<Tag | null>(null)

// 方法
const loadTags = async () => {
  loading.value = true
  try {
    const response = await tagAPI.getTags()
    if (response.data.success) {
      tags.value = response.data.data || []
    } else {
      ElMessage.error(response.data.message || '获取标签列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '获取标签列表失败')
  } finally {
    loading.value = false
  }
}

const editTag = (tag: Tag) => {
  editingTag.value = tag
  showCreateDialog.value = true
}

const deleteTag = async (tag: Tag) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除标签 "${tag.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await tagAPI.deleteTag(tag.id)
    if (response.data.success) {
      ElMessage.success('删除成功')
      loadTags()
    } else {
      ElMessage.error(response.data.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

const handleFormSuccess = () => {
  editingTag.value = null
  loadTags()
}

const formatDate = (dateStr: string): string => {
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

onMounted(() => {
  loadTags()
})
</script>

<style scoped>
.tags-page {
  width: 100%;
  max-width: 100%;
  margin: 0 auto;
  padding: 0 16px;
}

@media (min-width: 1200px) {
  .tags-page {
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
}

.operation-left {
  display: flex;
  gap: 12px;
}

.operation-right {
  display: flex;
  gap: 12px;
}

.tag-card {
  height: 100%;
  transition: all 0.3s ease;
  cursor: pointer;
}

.tag-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.tag-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tag-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.tag-color {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tag-name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tag-description {
  color: #7f8c8d;
  font-size: 14px;
  line-height: 1.4;
  margin: 0 0 16px 0;
  flex: 1;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.tag-stats {
  margin-bottom: 16px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  font-size: 12px;
}

.stat-item:last-child {
  margin-bottom: 0;
}

.stat-label {
  color: #909399;
}

.stat-value {
  color: #2c3e50;
  font-weight: 500;
}

.tag-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.empty-card {
  text-align: center;
  padding: 40px 0;
}

.loading-container {
  padding: 20px;
}

@media (max-width: 768px) {
  .operation-bar {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .operation-right {
    justify-content: center;
  }
}
</style>

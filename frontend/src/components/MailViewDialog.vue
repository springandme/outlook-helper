<template>
  <el-dialog
    v-model="visible"
    title="邮件查看"
    width="800px"
    :before-close="handleClose"
  >
    <div v-if="mailData" class="mail-content">
      <!-- 邮件列表 -->
      <div v-if="mailList.length > 1" class="mail-list">
        <h4>邮件列表 ({{ mailList.length }} 封)</h4>
        <el-table
          :data="mailList"
          style="width: 100%"
          size="small"
          @row-click="selectMail"
          highlight-current-row
        >
          <el-table-column prop="subject" label="主题" min-width="200" />
          <el-table-column prop="from" label="发件人" width="150" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_read ? 'info' : 'success'" size="small">
                {{ row.is_read ? '已读' : '未读' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="received_at" label="接收时间" width="150">
            <template #default="{ row }">
              {{ formatTime(row.received_at) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <!-- 邮件详情 -->
      <div class="mail-detail">
        <h4>邮件详情</h4>
        <div class="mail-header">
          <div class="header-item">
            <label>主题：</label>
            <span>{{ mailData.subject || '无主题' }}</span>
          </div>
          <div class="header-item">
            <label>发件人：</label>
            <span>{{ mailData.from }}</span>
          </div>
          <div class="header-item">
            <label>收件人：</label>
            <span>{{ mailData.to }}</span>
          </div>
          <div class="header-item">
            <label>接收时间：</label>
            <span>{{ formatTime(mailData.received_at) }}</span>
          </div>
          <div class="header-item">
            <label>状态：</label>
            <el-tag :type="mailData.is_read ? 'info' : 'success'" size="small">
              {{ mailData.is_read ? '已读' : '未读' }}
            </el-tag>
          </div>
          <div v-if="mailData.verify_code" class="header-item verify-code">
            <label>验证码：</label>
            <el-tag type="warning" size="large" class="code-tag">
              {{ mailData.verify_code }}
            </el-tag>
            <el-button
              size="small"
              type="primary"
              @click="copyVerifyCode"
            >
              复制
            </el-button>
          </div>
        </div>
        
        <div class="mail-body">
          <h5>邮件内容：</h5>
          <div class="body-content" v-html="formatMailBody(mailData.body)"></div>
        </div>
      </div>
    </div>
    
    <div v-else class="empty-content">
      <el-empty description="暂无邮件数据" />
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button v-if="mailData?.verify_code" type="primary" @click="copyVerifyCode">
          复制验证码
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { OutlookMail } from '@/api'

interface Props {
  modelValue: boolean
  mailData: OutlookMail | null
  mailList: OutlookMail[]
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const visible = ref(false)
const currentMail = ref<OutlookMail | null>(null)

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val && props.mailData) {
    currentMail.value = props.mailData
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const selectMail = (mail: OutlookMail) => {
  currentMail.value = mail
}

const handleClose = () => {
  visible.value = false
}

const formatTime = (timeStr: string): string => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

const formatMailBody = (body: string): string => {
  if (!body) return '无内容'

  // 检查是否为HTML内容
  const isHTML = /<[^>]+>/.test(body)

  if (isHTML) {
    // 如果是HTML内容，进行安全处理但保留样式
    let htmlContent = body

    // 清理多余的空白和换行
    htmlContent = htmlContent
      // 移除HTML标签之间的多余空白
      .replace(/>\s+</g, '><')
      // 移除开头和结尾的空白
      .trim()
      // 移除连续的空行（保留在HTML标签内的换行）
      .replace(/\n\s*\n\s*\n/g, '\n\n')
      // 移除HTML标签前后的多余空白，但保留标签内的内容
      .replace(/\s*(<[^>]+>)\s*/g, '$1')
      // 处理相对路径的图片和链接
      .replace(/src="(?!http|data:)/g, 'src="https://')
      .replace(/href="(?!http|mailto:|#)/g, 'href="https://')

    // 确保HTML内容在一个容器中，并设置基本样式
    if (!body.includes('<html') && !body.includes('<body')) {
      htmlContent = `<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 100%; word-wrap: break-word;">${htmlContent}</div>`
    }

    return htmlContent
  } else {
    // 纯文本处理
    let formatted = body
      // 移除开头和结尾的空白
      .trim()
      // 移除连续的空行，最多保留一个空行
      .replace(/\n\s*\n\s*\n/g, '\n\n')
      // 移除行首行尾的空白
      .replace(/^\s+|\s+$/gm, '')
      // HTML转义
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      // 转换换行和制表符
      .replace(/\n/g, '<br>')
      .replace(/\t/g, '&nbsp;&nbsp;&nbsp;&nbsp;')
      .replace(/  /g, '&nbsp;&nbsp;')

    return `<div style="white-space: pre-wrap; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; word-wrap: break-word;">${formatted}</div>`
  }
}

const copyVerifyCode = async () => {
  if (!currentMail.value?.verify_code) return
  
  try {
    await navigator.clipboard.writeText(currentMail.value.verify_code)
    ElMessage.success('验证码已复制到剪贴板')
  } catch (error) {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = currentMail.value.verify_code
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    ElMessage.success('验证码已复制到剪贴板')
  }
}

// 计算属性，获取当前显示的邮件数据
const mailData = computed(() => currentMail.value || props.mailData)
</script>

<style scoped>
.mail-content {
  max-height: 600px;
  overflow-y: auto;
}

.mail-list {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
}

.mail-list h4 {
  margin: 0 0 12px 0;
  color: #2c3e50;
  font-size: 16px;
}

.mail-detail h4 {
  margin: 0 0 16px 0;
  color: #2c3e50;
  font-size: 16px;
}

.mail-header {
  background-color: #f5f7fa;
  padding: 16px;
  border-radius: 6px;
  margin-bottom: 16px;
}

.header-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.header-item:last-child {
  margin-bottom: 0;
}

.header-item label {
  font-weight: 600;
  color: #606266;
  min-width: 80px;
  margin-right: 8px;
}

.header-item span {
  color: #2c3e50;
  word-break: break-all;
}

.verify-code {
  align-items: center;
  gap: 8px;
}

.code-tag {
  font-family: 'Courier New', monospace;
  font-size: 16px;
  font-weight: bold;
  letter-spacing: 2px;
}

.mail-body {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
}

.mail-body h5 {
  margin: 0;
  padding: 12px 16px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  color: #606266;
  font-size: 14px;
  font-weight: 600;
}

.body-content {
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
  line-height: 1.6;
  color: #2c3e50;
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 6px;
  font-size: 14px;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.empty-content {
  text-align: center;
  padding: 40px 0;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-table__row) {
  cursor: pointer;
}

:deep(.el-table__row:hover) {
  background-color: #f5f7fa;
}

:deep(.body-content pre) {
  margin: 0;
  font-size: 14px;
}

:deep(.body-content img) {
  max-width: 100%;
  height: auto;
}

:deep(.body-content a) {
  color: #409eff;
  text-decoration: none;
}

:deep(.body-content a:hover) {
  text-decoration: underline;
}

/* HTML邮件内容样式优化 */
:deep(.body-content table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
}

:deep(.body-content td),
:deep(.body-content th) {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

:deep(.body-content th) {
  background-color: #f2f2f2;
  font-weight: bold;
}

:deep(.body-content p) {
  margin: 8px 0;
}

:deep(.body-content h1),
:deep(.body-content h2),
:deep(.body-content h3),
:deep(.body-content h4),
:deep(.body-content h5),
:deep(.body-content h6) {
  margin: 16px 0 8px 0;
  color: #2c3e50;
}

:deep(.body-content ul),
:deep(.body-content ol) {
  margin: 8px 0;
  padding-left: 20px;
}

:deep(.body-content blockquote) {
  margin: 8px 0;
  padding: 8px 16px;
  border-left: 4px solid #409eff;
  background-color: #f0f9ff;
}

:deep(.body-content code) {
  background-color: #f1f2f6;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}
</style>

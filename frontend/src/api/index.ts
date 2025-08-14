import axios from 'axios'
import type { AxiosResponse } from 'axios'

// API基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加认证token
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      // 清除token并跳转到登录页
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API响应类型
export interface APIResponse<T = any> {
  success: boolean
  message: string
  data?: T
  error?: string
}

// 用户相关类型
export interface User {
  id: number
  username: string
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  auth_token: string
}

export interface LoginResponse {
  token: string
  expires_at: number
  user: User
}

// 邮箱相关类型
export interface Email {
  id: number
  user_id: number
  email_address: string
  remark: string
  last_operation_at?: string
  created_at: string
  updated_at: string
  tags?: Tag[]
}

export interface AddEmailRequest {
  email_address: string
  password: string
  client_id: string
  refresh_token: string
  remark?: string
}

export interface BatchAddEmailRequest {
  emails: AddEmailRequest[]
}

// 标记相关类型
export interface Tag {
  id: number
  name: string
  description: string
  color: string
  created_at: string
  updated_at: string
  email_count?: number
}

export interface CreateTagRequest {
  name: string
  description: string
  color: string
}

export interface UpdateTagRequest {
  name?: string
  description?: string
  color?: string
}

export interface TagEmailRequest {
  email_ids: number[]
  tag_id: number
}

// 邮件相关类型
export interface OutlookMail {
  id: string
  subject: string
  from: string
  to: string
  body: string
  is_read: boolean
  received_at: string
  verify_code?: string
}

// 仪表盘相关类型
export interface DashboardStats {
  total_emails: number
  total_tags: number
  recent_operations: OperationLog[]
  emails_by_tag: Record<string, number>
  operations_by_type: Record<string, number>
}

export interface OperationLog {
  id: number
  user_id: number
  operation_type: string
  target_type: string
  target_id?: number
  description: string
  ip_address: string
  user_agent: string
  created_at: string
}

// 认证API
export const authAPI = {
  // 登录
  login: (data: LoginRequest): Promise<AxiosResponse<APIResponse<LoginResponse>>> =>
    api.post('/auth/login', data),
  
  // 登出
  logout: (): Promise<AxiosResponse<APIResponse>> =>
    api.post('/auth/logout')
}

// 邮箱API
export const emailAPI = {
  // 获取邮箱列表
  getEmails: (params?: { limit?: number; offset?: number; keyword?: string }): Promise<AxiosResponse<APIResponse<Email[]>>> =>
    api.get('/emails', { params }),
  
  // 添加邮箱
  addEmail: (data: AddEmailRequest): Promise<AxiosResponse<APIResponse<Email>>> =>
    api.post('/emails', data),
  
  // 批量添加邮箱
  batchAddEmails: (data: BatchAddEmailRequest): Promise<AxiosResponse<APIResponse>> =>
    api.post('/emails/batch', data),
  
  // 导入邮箱
  importEmails: (file: File): Promise<AxiosResponse<APIResponse>> => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/emails/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },
  
  // 删除邮箱
  deleteEmail: (id: number): Promise<AxiosResponse<APIResponse>> =>
    api.delete(`/emails/${id}`),

  // 批量删除邮箱
  batchDeleteEmails: (emailIds: number[]): Promise<AxiosResponse<APIResponse>> =>
    api.delete('/emails/batch', { data: { email_ids: emailIds } }),

  // 批量清空收件箱
  batchClearInbox: (emailIds: number[]): Promise<AxiosResponse<APIResponse>> =>
    api.post('/emails/batch-clear-inbox', { email_ids: emailIds }),

  // 标记邮箱
  tagEmail: (id: number, data: TagEmailRequest): Promise<AxiosResponse<APIResponse>> =>
    api.put(`/emails/${id}/tags`, data),
  
  // 获取最新邮件
  getLatestMail: (id: number, mailbox?: string): Promise<AxiosResponse<APIResponse<OutlookMail>>> =>
    api.get(`/emails/${id}/latest`, { params: { mailbox } }),
  
  // 获取全部邮件
  getAllMails: (id: number, mailbox?: string): Promise<AxiosResponse<APIResponse<OutlookMail[]>>> =>
    api.get(`/emails/${id}/all`, { params: { mailbox } }),
  
  // 清空收件箱
  clearInbox: (id: number): Promise<AxiosResponse<APIResponse>> =>
    api.delete(`/emails/${id}/inbox`)
}

// 标记API
export const tagAPI = {
  // 获取标记列表
  getTags: (): Promise<AxiosResponse<APIResponse<Tag[]>>> =>
    api.get('/tags'),
  
  // 创建标记
  createTag: (data: CreateTagRequest): Promise<AxiosResponse<APIResponse<Tag>>> =>
    api.post('/tags', data),
  
  // 更新标记
  updateTag: (id: number, data: UpdateTagRequest): Promise<AxiosResponse<APIResponse<Tag>>> =>
    api.put(`/tags/${id}`, data),
  
  // 删除标记
  deleteTag: (id: number): Promise<AxiosResponse<APIResponse>> =>
    api.delete(`/tags/${id}`),
  
  // 批量标记邮箱
  batchTagEmails: (data: TagEmailRequest): Promise<AxiosResponse<APIResponse>> =>
    api.post('/tags/batch-tag', data),
  
  // 批量取消标记邮箱
  batchUntagEmails: (data: TagEmailRequest): Promise<AxiosResponse<APIResponse>> =>
    api.post('/tags/batch-untag', data)
}

// 仪表盘API
export const dashboardAPI = {
  // 获取仪表盘数据
  getDashboard: (): Promise<AxiosResponse<APIResponse<DashboardStats>>> =>
    api.get('/dashboard'),

  // 获取详细统计
  getStats: (type?: string): Promise<AxiosResponse<APIResponse>> =>
    api.get('/dashboard/stats', { params: { type } })
}

// 操作日志API
export const logsAPI = {
  // 获取操作日志（分页）
  getLogs: (page: number = 1, pageSize: number = 5): Promise<AxiosResponse<APIResponse<{
    logs: OperationLog[]
    total: number
    page: number
    page_size: number
    total_pages: number
  }>>> =>
    api.get('/logs', { params: { page, page_size: pageSize } }),

  // 清空所有操作日志
  clearLogs: (): Promise<AxiosResponse<APIResponse>> =>
    api.delete('/logs')
}

export default api

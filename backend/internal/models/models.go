package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID           int        `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Email 邮箱模型
type Email struct {
	ID              int        `json:"id" db:"id"`
	UserID          int        `json:"user_id" db:"user_id"`
	EmailAddress    string     `json:"email_address" db:"email_address"`
	Password        string     `json:"-" db:"password"`
	ClientID        string     `json:"-" db:"client_id"`
	RefreshToken    string     `json:"-" db:"refresh_token"`
	Remark          string     `json:"remark" db:"remark"`
	LastOperationAt *time.Time `json:"last_operation_at" db:"last_operation_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	Tags            []Tag      `json:"tags,omitempty"`
}

// Tag 标记模型
type Tag struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	EmailCount  int       `json:"email_count,omitempty"`
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	OperationType string    `json:"operation_type" db:"operation_type"`
	OperationName string    `json:"operation_name,omitempty"` // 操作类型中文名称
	TargetType    string    `json:"target_type" db:"target_type"`
	TargetID      *int      `json:"target_id" db:"target_id"`
	Description   string    `json:"description" db:"description"`
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	UserAgent     string    `json:"user_agent" db:"user_agent"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// OutlookMail Outlook邮件模型
type OutlookMail struct {
	ID         string    `json:"id"`
	Subject    string    `json:"subject"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Body       string    `json:"body"`
	IsRead     bool      `json:"is_read"`
	ReceivedAt time.Time `json:"received_at"`
	VerifyCode string    `json:"verify_code,omitempty"`
}

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	TotalEmails      int            `json:"total_emails"`
	TotalTags        int            `json:"total_tags"`
	RecentOperations []OperationLog `json:"recent_operations"`
	EmailsByTag      map[string]int `json:"emails_by_tag"`
	OperationsByType map[string]int `json:"operations_by_type"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	AuthToken string `json:"auth_token" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      User   `json:"user"`
}

// AddEmailRequest 添加邮箱请求
type AddEmailRequest struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
	Remark       string `json:"remark"`
}

// BatchAddEmailRequest 批量添加邮箱请求
type BatchAddEmailRequest struct {
	Emails []AddEmailRequest `json:"emails" binding:"required,dive"`
}

// CreateTagRequest 创建标记请求
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"required"`
}

// UpdateTagRequest 更新标记请求
type UpdateTagRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// TagEmailRequest 标记邮箱请求
type TagEmailRequest struct {
	EmailIDs []int `json:"email_ids" binding:"required"`
	TagID    int   `json:"tag_id" binding:"required"`
}

// FieldOption 字段选项
type FieldOption struct {
	Key   string `json:"key" binding:"required"`
	Label string `json:"label" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// ExportEmailRequest 导出邮箱请求
type ExportEmailRequest struct {
	Range      string        `json:"range" binding:"required,oneof=all selected"`
	Format     string        `json:"format" binding:"required,oneof=txt csv"`
	FieldOrder []FieldOption `json:"field_order" binding:"required,dive"`
	EmailIDs   []int         `json:"email_ids,omitempty"`
}

// ExportEmailResponse 导出邮箱响应
type ExportEmailResponse struct {
	Content string `json:"content"`
	Count   int    `json:"count"`
	Format  string `json:"format"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

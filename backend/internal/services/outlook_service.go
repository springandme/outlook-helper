package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"outlook-helper/backend/internal/models"
)

// OutlookService Outlook API服务
type OutlookService struct {
	baseURL    string
	httpClient *http.Client
}

// NewOutlookService 创建Outlook服务
func NewOutlookService(baseURL string) *OutlookService {
	return &OutlookService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OutlookAPIResponse Outlook API响应结构
type OutlookAPIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// MailData 邮件数据结构
type MailData struct {
	ID          string    `json:"id"`
	Subject     string    `json:"subject"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Body        string    `json:"body"`
	BodyPreview string    `json:"bodyPreview"`
	IsRead      bool      `json:"isRead"`
	ReceivedAt  time.Time `json:"receivedDateTime"`
	VerifyCode  string    `json:"verifyCode,omitempty"`
}

// GetLatestMail 获取最新邮件
func (s *OutlookService) GetLatestMail(email *models.Email, mailbox string, responseType string) (*models.OutlookMail, error) {
	// 构建请求体
	requestData := map[string]string{
		"refresh_token": email.RefreshToken,
		"client_id":     email.ClientID,
		"email":         email.EmailAddress,
		"mailbox":       mailbox,
	}
	if responseType != "" {
		requestData["response_type"] = responseType
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	requestURL := fmt.Sprintf("%s/api/mail-new", s.baseURL)

	// 创建POST请求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 尝试解析邮件数据，支持两种格式：单个对象或数组
	var mailData map[string]interface{}

	// 首先尝试解析为单个对象
	if err := json.Unmarshal(body, &mailData); err != nil {
		// 如果失败，尝试解析为数组格式
		var mailsArray []map[string]interface{}
		if arrayErr := json.Unmarshal(body, &mailsArray); arrayErr != nil {
			// 两种格式都解析失败，记录完整的响应内容
			return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
		}

		// 如果是数组格式，取第一个元素作为最新邮件
		if len(mailsArray) == 0 {
			return nil, fmt.Errorf("API返回空邮件数组")
		}
		mailData = mailsArray[0]
	}

	// 解析时间字段
	var receivedAt time.Time
	if dateStr := getStringFromMap(mailData, "date"); dateStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
			receivedAt = parsedTime
		}
	}

	// 优先获取HTML格式的邮件内容，如果没有则使用text字段
	mailBody := getStringFromMap(mailData, "html")
	if mailBody == "" {
		mailBody = getStringFromMap(mailData, "text")
	}

	mail := &models.OutlookMail{
		ID:         getStringFromMap(mailData, "id"),
		Subject:    getStringFromMap(mailData, "subject"),
		From:       getStringFromMap(mailData, "send"), // API返回的字段名是"send"
		To:         getStringFromMap(mailData, "to"),
		Body:       mailBody, // 优先使用HTML格式，否则使用text字段
		IsRead:     getBoolFromMap(mailData, "isRead"),
		VerifyCode: getStringFromMap(mailData, "verifyCode"),
		ReceivedAt: receivedAt,
	}

	// 解析时间
	if receivedAtStr := getStringFromMap(mailData, "receivedDateTime"); receivedAtStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, receivedAtStr); err == nil {
			mail.ReceivedAt = parsedTime
		}
	}

	return mail, nil
}

// GetAllMails 获取全部邮件
func (s *OutlookService) GetAllMails(email *models.Email, mailbox string) ([]models.OutlookMail, error) {
	// 构建请求体
	requestData := map[string]string{
		"refresh_token": email.RefreshToken,
		"client_id":     email.ClientID,
		"email":         email.EmailAddress,
		"mailbox":       mailbox,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	requestURL := fmt.Sprintf("%s/api/mail-all", s.baseURL)

	// 创建POST请求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 直接解析邮件数组（API直接返回邮件对象数组，不包装在通用响应中）
	var mailsData []map[string]interface{}
	if err := json.Unmarshal(body, &mailsData); err != nil {
		// 记录完整的响应内容以便调试
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	var mails []models.OutlookMail
	for _, mailData := range mailsData {
		// 解析时间字段
		var receivedAt time.Time
		if dateStr := getStringFromMap(mailData, "date"); dateStr != "" {
			if parsedTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
				receivedAt = parsedTime
			}
		}

		// 优先获取HTML格式的邮件内容，如果没有则使用text字段
		mailBody := getStringFromMap(mailData, "html")
		if mailBody == "" {
			mailBody = getStringFromMap(mailData, "text")
		}

		mail := models.OutlookMail{
			ID:         getStringFromMap(mailData, "id"),
			Subject:    getStringFromMap(mailData, "subject"),
			From:       getStringFromMap(mailData, "send"), // API返回的字段名是"send"
			To:         getStringFromMap(mailData, "to"),
			Body:       mailBody, // 优先使用HTML格式，否则使用text字段
			IsRead:     getBoolFromMap(mailData, "isRead"),
			VerifyCode: getStringFromMap(mailData, "verifyCode"),
			ReceivedAt: receivedAt,
		}

		mails = append(mails, mail)
	}

	return mails, nil
}

// ClearInbox 清空收件箱
func (s *OutlookService) ClearInbox(email *models.Email) error {
	// 构建请求体
	requestData := map[string]string{
		"refresh_token": email.RefreshToken,
		"client_id":     email.ClientID,
		"email":         email.EmailAddress,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	requestURL := fmt.Sprintf("%s/api/process-inbox", s.baseURL)

	// 创建POST请求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应（API直接返回消息对象）
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		// 记录完整的响应内容以便调试
		return fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	// 检查是否有错误消息
	if errorMsg := getStringFromMap(response, "error"); errorMsg != "" {
		return fmt.Errorf("API返回错误: %s", errorMsg)
	}

	return nil
}

// ClearJunk 清空垃圾箱
func (s *OutlookService) ClearJunk(email *models.Email) error {
	// 构建请求体
	requestData := map[string]string{
		"refresh_token": email.RefreshToken,
		"client_id":     email.ClientID,
		"email":         email.EmailAddress,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	requestURL := fmt.Sprintf("%s/api/process-junk", s.baseURL)

	// 创建POST请求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应（API直接返回消息对象）
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		// 记录完整的响应内容以便调试
		return fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	// 检查是否有错误消息
	if errorMsg := getStringFromMap(response, "error"); errorMsg != "" {
		return fmt.Errorf("API返回错误: %s", errorMsg)
	}

	return nil
}

// ValidateEmailCredentials 验证邮箱凭据
func (s *OutlookService) ValidateEmailCredentials(email *models.Email) error {
	// 尝试获取最新邮件来验证凭据
	_, err := s.GetLatestMail(email, "INBOX", "json")
	if err != nil {
		// 为验证失败提供更详细的错误信息
		return fmt.Errorf("验证邮箱 %s 凭据失败: %v", email.EmailAddress, err)
	}
	return nil
}

// 辅助函数
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

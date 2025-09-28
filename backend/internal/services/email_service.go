package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"outlook-helper/backend/internal/config"
	"outlook-helper/backend/internal/database"
	"outlook-helper/backend/internal/models"
)

// EmailService 邮件服务
type EmailService struct {
	emailRepo      *database.EmailRepository
	logRepo        *database.LogRepository
	outlookService *OutlookService
	config         *config.Config
}

// NewEmailService 创建邮件服务
func NewEmailService(db *database.DB, outlookService *OutlookService, cfg *config.Config) *EmailService {
	return &EmailService{
		emailRepo:      db.Email,
		logRepo:        db.Log,
		outlookService: outlookService,
		config:         cfg,
	}
}

// AddEmail 添加邮箱
func (s *EmailService) AddEmail(userID int, req *models.AddEmailRequest, ipAddress, userAgent string) (*models.Email, error) {
	// 检查邮箱是否已存在
	exists, err := s.emailRepo.EmailExists(userID, req.EmailAddress)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("邮箱已存在")
	}

	// 创建邮箱对象
	email := &models.Email{
		UserID:       userID,
		EmailAddress: req.EmailAddress,
		Password:     req.Password,
		ClientID:     req.ClientID,
		RefreshToken: req.RefreshToken,
		Remark:       req.Remark,
	}

	// 验证邮箱凭据（如果配置允许跳过验证则跳过）
	if !s.config.SkipEmailValidation {
		if err := s.outlookService.ValidateEmailCredentials(email); err != nil {
			// 记录验证失败日志，包含详细错误信息
			s.logRepo.LogEmail(userID, "email_validation_failed", 0,
				fmt.Sprintf("邮箱 %s 凭据验证失败: %v", req.EmailAddress, err),
				ipAddress, userAgent)
			return nil, err // 直接返回详细的错误信息
		}
	} else {
		// 记录跳过验证的日志
		s.logRepo.LogEmail(userID, "email_validation_skipped", 0,
			fmt.Sprintf("邮箱 %s 跳过凭据验证（调试模式）", req.EmailAddress),
			ipAddress, userAgent)
	}

	// 保存到数据库
	savedEmail, err := s.emailRepo.CreateEmail(email)
	if err != nil {
		return nil, err
	}

	// 记录添加成功日志
	s.logRepo.LogEmail(userID, "email_added", savedEmail.ID,
		fmt.Sprintf("添加邮箱: %s", req.EmailAddress),
		ipAddress, userAgent)

	return savedEmail, nil
}

// BatchAddEmails 批量添加邮箱 - 并发验证和批量处理
func (s *EmailService) BatchAddEmails(userID int, req *models.BatchAddEmailRequest, ipAddress, userAgent string) ([]models.Email, []string, error) {
	if len(req.Emails) == 0 {
		return []models.Email{}, []string{}, nil
	}

	// 限制最大数量为30
	if len(req.Emails) > 30 {
		return nil, nil, errors.New("批量添加邮箱数量不能超过30个")
	}

	// 使用worker pool控制并发度，避免对API造成过大压力
	maxWorkers := s.config.EmailValidationWorkers
	if maxWorkers <= 0 {
		maxWorkers = 5 // 默认并发数
	}

	type emailTask struct {
		index int
		req   models.AddEmailRequest
	}

	type emailResult struct {
		email *models.Email
		error string
		index int
	}

	taskChan := make(chan emailTask, len(req.Emails))
	resultChan := make(chan emailResult, len(req.Emails))

	// 启动worker goroutines
	for w := 0; w < maxWorkers; w++ {
		go func() {
			for task := range taskChan {
				result := emailResult{index: task.index}

				// 检查邮箱是否已存在
				exists, err := s.emailRepo.EmailExists(userID, task.req.EmailAddress)
				if err != nil {
					result.error = fmt.Sprintf("邮箱 %s: 检查邮箱存在性失败: %v", task.req.EmailAddress, err)
					resultChan <- result
					continue
				}
				if exists {
					result.error = fmt.Sprintf("邮箱 %s: 邮箱已存在", task.req.EmailAddress)
					resultChan <- result
					continue
				}

				// 创建邮箱对象
				email := &models.Email{
					UserID:       userID,
					EmailAddress: task.req.EmailAddress,
					Password:     task.req.Password,
					ClientID:     task.req.ClientID,
					RefreshToken: task.req.RefreshToken,
					Remark:       task.req.Remark,
				}

				// 验证邮箱凭据（如果配置允许跳过验证则跳过）
				if !s.config.SkipEmailValidation {
					if err := s.outlookService.ValidateEmailCredentials(email); err != nil {
						// 提供更详细的错误信息，包含具体的API响应
						result.error = fmt.Sprintf("邮箱 %s: %v", task.req.EmailAddress, err)
						resultChan <- result
						continue
					}
				}

				result.email = email
				resultChan <- result
			}
		}()
	}

	// 发送任务到worker pool
	for i, emailReq := range req.Emails {
		taskChan <- emailTask{index: i, req: emailReq}
	}
	close(taskChan)

	// 收集验证结果
	var validEmails []*models.Email
	var errors []string
	results := make([]emailResult, len(req.Emails))

	for i := 0; i < len(req.Emails); i++ {
		result := <-resultChan
		results[result.index] = result
	}

	// 按原始顺序处理结果
	for _, result := range results {
		if result.error != "" {
			errors = append(errors, result.error)
		} else if result.email != nil {
			validEmails = append(validEmails, result.email)
		}
	}

	// 批量保存验证成功的邮箱
	var successEmails []models.Email
	if len(validEmails) > 0 {
		savedEmails, err := s.emailRepo.BatchCreateEmails(validEmails)
		if err != nil {
			return nil, nil, fmt.Errorf("批量保存邮箱失败: %v", err)
		}
		successEmails = savedEmails

		// 为每个成功添加的邮箱记录日志
		for _, email := range savedEmails {
			s.logRepo.LogEmail(userID, "email_added", email.ID,
				fmt.Sprintf("添加邮箱: %s", email.EmailAddress),
				ipAddress, userAgent)
		}
	}

	// 记录批量添加日志
	s.logRepo.LogEmail(userID, "batch_add_emails", 0,
		fmt.Sprintf("批量添加邮箱，成功: %d, 失败: %d", len(successEmails), len(errors)),
		ipAddress, userAgent)

	return successEmails, errors, nil
}

// GetUserEmails 获取用户邮箱列表
func (s *EmailService) GetUserEmails(userID int, limit, offset int) ([]models.Email, error) {
	return s.emailRepo.GetEmailsByUserID(userID, limit, offset)
}

// SearchEmails 搜索邮箱
func (s *EmailService) SearchEmails(userID int, keyword string, limit, offset int) ([]models.Email, error) {
	return s.emailRepo.SearchEmails(userID, keyword, limit, offset)
}

// GetEmailByID 根据ID获取邮箱
func (s *EmailService) GetEmailByID(userID, emailID int) (*models.Email, error) {
	email, err := s.emailRepo.GetEmailByID(emailID)
	if err != nil {
		return nil, err
	}

	// 检查邮箱是否属于当前用户
	if email.UserID != userID {
		return nil, errors.New("无权访问此邮箱")
	}

	return email, nil
}

// UpdateEmail 更新邮箱
func (s *EmailService) UpdateEmail(userID int, emailID int, req *models.AddEmailRequest, ipAddress, userAgent string) (*models.Email, error) {
	// 获取现有邮箱
	email, err := s.GetEmailByID(userID, emailID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	email.EmailAddress = req.EmailAddress
	email.Password = req.Password
	email.ClientID = req.ClientID
	email.RefreshToken = req.RefreshToken
	email.Remark = req.Remark

	// 验证新的凭据
	if err := s.outlookService.ValidateEmailCredentials(email); err != nil {
		return nil, fmt.Errorf("邮箱凭据验证失败: %v", err)
	}

	// 更新数据库
	if err := s.emailRepo.UpdateEmail(email); err != nil {
		return nil, err
	}

	// 记录更新日志
	s.logRepo.LogEmail(userID, "email_updated", emailID,
		fmt.Sprintf("更新邮箱: %s", req.EmailAddress),
		ipAddress, userAgent)

	return email, nil
}

// DeleteEmail 删除邮箱
func (s *EmailService) DeleteEmail(userID, emailID int, ipAddress, userAgent string) error {
	// 检查邮箱是否存在且属于当前用户
	email, err := s.GetEmailByID(userID, emailID)
	if err != nil {
		return err
	}

	// 删除邮箱
	if err := s.emailRepo.DeleteEmail(emailID); err != nil {
		return err
	}

	// 记录删除日志
	s.logRepo.LogEmail(userID, "email_deleted", emailID,
		fmt.Sprintf("删除邮箱: %s", email.EmailAddress),
		ipAddress, userAgent)

	return nil
}

// BatchDeleteEmails 批量删除邮箱
func (s *EmailService) BatchDeleteEmails(userID int, emailIDs []int, ipAddress, userAgent string) error {
	if len(emailIDs) == 0 {
		return nil
	}

	// 验证所有邮箱都属于当前用户
	for _, emailID := range emailIDs {
		_, err := s.GetEmailByID(userID, emailID)
		if err != nil {
			return fmt.Errorf("邮箱ID %d 不存在或不属于当前用户", emailID)
		}
	}

	// 批量删除
	if err := s.emailRepo.BatchDeleteEmails(emailIDs); err != nil {
		return err
	}

	// 记录批量删除日志
	s.logRepo.LogEmail(userID, "batch_delete_emails", 0,
		fmt.Sprintf("批量删除邮箱，数量: %d", len(emailIDs)),
		ipAddress, userAgent)

	return nil
}

// GetLatestMail 获取最新邮件
func (s *EmailService) GetLatestMail(userID, emailID int, mailbox string, ipAddress, userAgent string) (*models.OutlookMail, error) {
	// 获取邮箱信息
	email, err := s.GetEmailByID(userID, emailID)
	if err != nil {
		return nil, err
	}

	// 调用Outlook API
	mail, err := s.outlookService.GetLatestMail(email, mailbox, "json")
	if err != nil {
		// 记录操作失败日志
		s.logRepo.LogEmail(userID, "get_latest_mail_failed", emailID,
			fmt.Sprintf("获取最新邮件失败: %v", err),
			ipAddress, userAgent)
		return nil, err
	}

	// 更新最后操作时间
	s.emailRepo.UpdateLastOperation(emailID)

	// 记录操作成功日志
	s.logRepo.LogEmail(userID, "get_latest_mail", emailID,
		fmt.Sprintf("获取最新邮件成功，邮箱: %s", email.EmailAddress),
		ipAddress, userAgent)

	return mail, nil
}

// GetAllMails 获取全部邮件
func (s *EmailService) GetAllMails(userID, emailID int, mailbox string, ipAddress, userAgent string) ([]models.OutlookMail, error) {
	// 获取邮箱信息
	email, err := s.GetEmailByID(userID, emailID)
	if err != nil {
		return nil, err
	}

	// 调用Outlook API
	mails, err := s.outlookService.GetAllMails(email, mailbox)
	if err != nil {
		// 记录操作失败日志
		s.logRepo.LogEmail(userID, "get_all_mails_failed", emailID,
			fmt.Sprintf("获取全部邮件失败: %v", err),
			ipAddress, userAgent)
		return nil, err
	}

	// 更新最后操作时间
	s.emailRepo.UpdateLastOperation(emailID)

	// 记录操作成功日志
	s.logRepo.LogEmail(userID, "get_all_mails", emailID,
		fmt.Sprintf("获取全部邮件成功，邮箱: %s，邮件数量: %d", email.EmailAddress, len(mails)),
		ipAddress, userAgent)

	return mails, nil
}

// ClearInbox 清空收件箱
func (s *EmailService) ClearInbox(userID, emailID int, ipAddress, userAgent string) error {
	// 获取邮箱信息
	email, err := s.GetEmailByID(userID, emailID)
	if err != nil {
		return err
	}

	// 调用Outlook API
	if err := s.outlookService.ClearInbox(email); err != nil {
		// 记录操作失败日志
		s.logRepo.LogEmail(userID, "clear_inbox_failed", emailID,
			fmt.Sprintf("清空收件箱失败: %v", err),
			ipAddress, userAgent)
		return err
	}

	// 更新最后操作时间
	s.emailRepo.UpdateLastOperation(emailID)

	// 记录操作成功日志
	s.logRepo.LogEmail(userID, "clear_inbox", emailID,
		fmt.Sprintf("清空收件箱成功，邮箱: %s", email.EmailAddress),
		ipAddress, userAgent)

	return nil
}

// BatchClearInbox 批量清空收件箱
func (s *EmailService) BatchClearInbox(userID int, emailIDs []int, ipAddress, userAgent string) (int, []string, error) {
	if len(emailIDs) == 0 {
		return 0, []string{}, nil
	}

	var successCount int
	var errors []string

	// 逐个清空收件箱
	for _, emailID := range emailIDs {
		// 验证邮箱是否属于当前用户
		email, err := s.GetEmailByID(userID, emailID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("邮箱ID %d: %v", emailID, err))
			continue
		}

		// 调用Outlook API清空收件箱
		if err := s.outlookService.ClearInbox(email); err != nil {
			// 记录操作失败日志
			s.logRepo.LogEmail(userID, "clear_inbox_failed", emailID,
				fmt.Sprintf("批量清空收件箱失败: %v", err),
				ipAddress, userAgent)
			errors = append(errors, fmt.Sprintf("邮箱 %s: %v", email.EmailAddress, err))
			continue
		}

		// 更新最后操作时间
		s.emailRepo.UpdateLastOperation(emailID)

		// 记录操作成功日志
		s.logRepo.LogEmail(userID, "clear_inbox", emailID,
			fmt.Sprintf("批量清空收件箱成功，邮箱: %s", email.EmailAddress),
			ipAddress, userAgent)

		successCount++
	}

	// 记录批量操作日志
	s.logRepo.LogEmail(userID, "batch_clear_inbox", 0,
		fmt.Sprintf("批量清空收件箱，成功: %d, 失败: %d", successCount, len(errors)),
		ipAddress, userAgent)

	return successCount, errors, nil
}

// CountUserEmails 统计用户邮箱数量
func (s *EmailService) CountUserEmails(userID int) (int, error) {
	return s.emailRepo.CountEmailsByUserID(userID)
}

// CountSearchEmails 统计搜索结果数量
func (s *EmailService) CountSearchEmails(userID int, keyword string) (int, error) {
	return s.emailRepo.CountSearchEmails(userID, keyword)
}

// ExportEmails 导出邮箱数据
func (s *EmailService) ExportEmails(userID int, req *models.ExportEmailRequest, ipAddress, userAgent string) (*models.ExportEmailResponse, error) {
	var emails []models.Email
	var err error

	// 根据导出范围获取邮箱数据
	if req.Range == "selected" {
		if len(req.EmailIDs) == 0 {
			return nil, errors.New("选择导出时必须提供邮箱ID列表")
		}

		// 验证所有邮箱都属于当前用户并获取数据
		emails, err = s.emailRepo.GetEmailsByIDs(userID, req.EmailIDs)
		if err != nil {
			return nil, err
		}

		// 验证数量匹配，确保所有请求的邮箱都存在且属于当前用户
		if len(emails) != len(req.EmailIDs) {
			return nil, errors.New("部分邮箱不存在或无权访问")
		}
	} else {
		// 导出全部邮箱
		emails, err = s.emailRepo.GetEmailsForExport(userID, req.SortField, req.SortDirection)
		if err != nil {
			return nil, err
		}
	}

	// 如果是选中的邮箱，需要根据排序字段和方向重新排序
	if req.Range == "selected" && len(emails) > 1 {
		emails = s.sortEmails(emails, req.SortField, req.SortDirection)
	}

	// 生成导出内容
	content := s.generateExportContent(emails, req.Format)

	// 记录导出日志
	s.logRepo.LogEmail(userID, "export_emails", 0,
		fmt.Sprintf("导出邮箱数据，范围: %s，格式: %s，数量: %d", req.Range, req.Format, len(emails)),
		ipAddress, userAgent)

	response := &models.ExportEmailResponse{
		Content: content,
		Count:   len(emails),
		Format:  req.Format,
	}

	return response, nil
}

// generateExportContent 生成导出内容
func (s *EmailService) generateExportContent(emails []models.Email, format string) string {
	if len(emails) == 0 {
		return ""
	}

	var content strings.Builder

	if format == "csv" {
		// CSV格式：添加头部
		content.WriteString("邮箱地址,密码,RefreshToken,ClientID,备注,添加时间\n")

		for _, email := range emails {
			// CSV格式需要处理特殊字符和引号
			content.WriteString(fmt.Sprintf(`"%s","%s","%s","%s","%s","%s"`,
				s.escapeCSV(email.EmailAddress),
				s.escapeCSV(email.Password),
				s.escapeCSV(email.RefreshToken),
				s.escapeCSV(email.ClientID),
				s.escapeCSV(email.Remark),
				email.CreatedAt.Format("2006-01-02 15:04:05"),
			))
			content.WriteString("\n")
		}
	} else {
		// TXT格式：每行一个邮箱，使用----分隔
		for _, email := range emails {
			content.WriteString(fmt.Sprintf("%s----%s----%s----%s",
				email.EmailAddress,
				email.Password,
				email.RefreshToken,
				email.ClientID,
			))
			content.WriteString("\n")
		}
	}

	return content.String()
}

// escapeCSV CSV字段转义处理
func (s *EmailService) escapeCSV(field string) string {
	// 如果字段包含引号，需要双引号转义
	if strings.Contains(field, `"`) {
		field = strings.ReplaceAll(field, `"`, `""`)
	}
	return field
}

// sortEmails 对邮箱列表进行排序
func (s *EmailService) sortEmails(emails []models.Email, sortField, sortDirection string) []models.Email {
	sort.Slice(emails, func(i, j int) bool {
		var less bool

		switch sortField {
		case "email_address":
			less = emails[i].EmailAddress < emails[j].EmailAddress
		case "password":
			less = emails[i].Password < emails[j].Password
		case "refresh_token":
			less = emails[i].RefreshToken < emails[j].RefreshToken
		case "client_id":
			less = emails[i].ClientID < emails[j].ClientID
		case "created_at":
			less = emails[i].CreatedAt.Before(emails[j].CreatedAt)
		default:
			// 默认按邮箱地址排序
			less = emails[i].EmailAddress < emails[j].EmailAddress
		}

		if sortDirection == "desc" {
			return !less
		}
		return less
	})

	return emails
}

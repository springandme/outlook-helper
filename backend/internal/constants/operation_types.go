package constants

// 操作类型常量
const (
	// 认证相关
	OpLoginSuccess = "login_success"
	OpLoginFailed  = "login_failed"
	OpLogout       = "logout"
	OpClearAllLogs = "clear_all_logs"

	// 邮箱相关
	OpEmailAdded            = "email_added"
	OpEmailDeleted          = "email_deleted"
	OpEmailUpdated          = "email_updated"
	OpEmailValidationFailed = "email_validation_failed"
	OpBatchAddEmails        = "batch_add_emails"
	OpBatchDeleteEmails     = "batch_delete_emails"

	// 邮件操作相关
	OpGetLatestMail       = "get_latest_mail"
	OpGetLatestMailFailed = "get_latest_mail_failed"
	OpGetAllMails         = "get_all_mails"
	OpGetAllMailsFailed   = "get_all_mails_failed"
	OpClearInbox          = "clear_inbox"
	OpClearInboxFailed    = "clear_inbox_failed"
	OpClearJunk           = "clear_junk"
	OpClearJunkFailed     = "clear_junk_failed"

	// 标签相关
	OpTagCreated       = "tag_created"
	OpTagUpdated       = "tag_updated"
	OpTagDeleted       = "tag_deleted"
	OpEmailTagged      = "email_tagged"
	OpEmailUntagged    = "email_untagged"
	OpBatchTagEmails   = "batch_tag_emails"
	OpBatchUntagEmails = "batch_untag_emails"
)

// 操作类型中文映射
var OperationTypeNames = map[string]string{
	// 认证相关
	OpLoginSuccess: "登录成功",
	OpLoginFailed:  "登录失败",
	OpLogout:       "退出登录",
	OpClearAllLogs: "清空操作日志",

	// 邮箱相关
	OpEmailAdded:            "添加邮箱",
	OpEmailDeleted:          "删除邮箱",
	OpEmailUpdated:          "更新邮箱",
	OpEmailValidationFailed: "邮箱验证失败",
	OpBatchAddEmails:        "批量添加邮箱",
	OpBatchDeleteEmails:     "批量删除邮箱",

	// 邮件操作相关
	OpGetLatestMail:       "获取最新邮件",
	OpGetLatestMailFailed: "获取邮件失败",
	OpGetAllMails:         "获取全部邮件",
	OpGetAllMailsFailed:   "获取邮件失败",
	OpClearInbox:          "清空收件箱",
	OpClearInboxFailed:    "清空收件箱失败",
	OpClearJunk:           "清空垃圾箱",
	OpClearJunkFailed:     "清空垃圾箱失败",

	// 标签相关
	OpTagCreated:       "创建标签",
	OpTagUpdated:       "更新标签",
	OpTagDeleted:       "删除标签",
	OpEmailTagged:      "添加标签",
	OpEmailUntagged:    "移除标签",
	OpBatchTagEmails:   "批量添加标签",
	OpBatchUntagEmails: "批量移除标签",
}

// GetOperationTypeName 获取操作类型的中文名称
func GetOperationTypeName(operationType string) string {
	if name, exists := OperationTypeNames[operationType]; exists {
		return name
	}
	return operationType
}

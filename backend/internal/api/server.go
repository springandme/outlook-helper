package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"outlook-helper/backend/internal/auth"
	"outlook-helper/backend/internal/config"
	"outlook-helper/backend/internal/database"
	"outlook-helper/backend/internal/models"
	"outlook-helper/backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server API服务器
type Server struct {
	config       *config.Config
	db           *database.DB
	router       *gin.Engine
	authService  *auth.Service
	emailService *services.EmailService
}

// NewServer 创建新的API服务器
func NewServer(cfg *config.Config, db *database.DB) *Server {
	// 创建认证服务
	authService := auth.NewService(db, cfg.JWTSecret, cfg.JWTExpire, cfg)

	// 创建Outlook服务
	outlookService := services.NewOutlookService(cfg.OutlookAPI)

	// 创建邮件服务
	emailService := services.NewEmailService(db, outlookService, cfg)

	server := &Server{
		config:       cfg,
		db:           db,
		authService:  authService,
		emailService: emailService,
	}

	server.setupRouter()
	return server
}

// setupRouter 设置路由
func (s *Server) setupRouter() {
	// 设置Gin模式
	if s.config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	s.router = gin.Default()

	// 配置CORS
	corsConfig := cors.DefaultConfig()
	if s.config.CORSOrigins == "" {
		// 如果没有配置CORS来源，则允许所有来源
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = strings.Split(s.config.CORSOrigins, ",")
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	s.router.Use(cors.New(corsConfig))

	// 静态文件服务 - 为assets目录设置正确的MIME类型
	s.router.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		fullPath := "./backend/static/assets" + filepath

		// 根据文件扩展名设置正确的MIME类型
		if strings.HasSuffix(filepath, ".js") {
			c.Header("Content-Type", "application/javascript; charset=utf-8")
			c.Header("Cache-Control", "public, max-age=31536000") // 1年缓存
		} else if strings.HasSuffix(filepath, ".css") {
			c.Header("Content-Type", "text/css; charset=utf-8")
			c.Header("Cache-Control", "public, max-age=31536000") // 1年缓存
		} else if strings.HasSuffix(filepath, ".map") {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}

		c.File(fullPath)
	})

	// 静态资源文件服务（favicon.ico, outlook.svg等）
	s.router.GET("/favicon.ico", func(c *gin.Context) {
		c.Header("Content-Type", "image/x-icon")
		c.File("./backend/static/favicon.ico")
	})

	s.router.GET("/outlook.svg", func(c *gin.Context) {
		c.Header("Content-Type", "image/svg+xml")
		c.File("./backend/static/outlook.svg")
	})

	// 其他静态文件
	s.router.Static("/static", "./backend/static")

	// 根路径返回index.html
	s.router.StaticFile("/", "./backend/static/index.html")

	// API路由组
	api := s.router.Group("/api")
	{
		// 健康检查接口（无需认证）
		api.GET("/health", s.handleHealth)

		// 认证相关
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", s.handleLogin)
		}

		// 需要认证的认证相关接口
		authProtected := api.Group("/auth")
		authProtected.Use(auth.AuthMiddleware(s.authService))
		{
			authProtected.POST("/logout", s.handleLogout)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(auth.AuthMiddleware(s.authService))
		{
			// 仪表盘
			protected.GET("/dashboard", s.handleDashboard)
			protected.GET("/dashboard/stats", s.handleDashboardStats)

			// 邮箱管理
			emails := protected.Group("/emails")
			{
				emails.GET("", s.handleGetEmails)
				emails.POST("", s.handleAddEmail)
				emails.POST("/batch", s.handleBatchAddEmails)
				emails.POST("/import", s.handleImportEmails)
				emails.DELETE("/batch", s.handleBatchDeleteEmails)
				emails.POST("/batch-clear-inbox", s.handleBatchClearInbox)
				emails.GET("/:id/latest", s.handleGetLatestMail)
				emails.GET("/:id/all", s.handleGetAllMails)
				emails.DELETE("/:id/inbox", s.handleClearInbox)
				emails.PUT("/:id/tags", s.handleTagEmail)
				emails.DELETE("/:id", s.handleDeleteEmail)
			}

			// 标记管理
			tags := protected.Group("/tags")
			{
				tags.GET("", s.handleGetTags)
				tags.POST("", s.handleCreateTag)
				tags.PUT("/:id", s.handleUpdateTag)
				tags.DELETE("/:id", s.handleDeleteTag)
				tags.POST("/batch-tag", s.handleBatchTagEmails)
				tags.POST("/batch-untag", s.handleBatchUntagEmails)
			}

			// 操作日志管理
			logs := protected.Group("/logs")
			{
				logs.GET("", s.handleGetLogs)
				logs.DELETE("", s.handleClearLogs)
			}
		}
	}

	// 404处理 - 只对HTML页面请求返回index.html，其他返回404
	s.router.NoRoute(func(c *gin.Context) {
		// 检查请求的Accept头部，如果是HTML请求则返回index.html
		accept := c.GetHeader("Accept")
		if strings.Contains(accept, "text/html") {
			c.File("./backend/static/index.html")
		} else {
			c.JSON(404, gin.H{"error": "Not Found"})
		}
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Port)
}

// handleHealth 健康检查接口
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Outlook取件助手服务运行正常",
		"version": "1.0.0",
	})
}

// handleLogin 处理用户登录
func (s *Server) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 执行登录
	response, err := s.authService.Login(req.AuthToken, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "登录失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登录成功",
		Data:    response,
	})
}

// handleLogout 处理用户登出
func (s *Server) handleLogout(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 执行登出
	if err := s.authService.Logout(userID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "登出失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "登出成功",
	})
}

// handleDashboard 获取仪表盘数据
func (s *Server) handleDashboard(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取邮箱总数
	totalEmails, err := s.emailService.CountUserEmails(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取邮箱统计失败",
			Error:   err.Error(),
		})
		return
	}

	// 获取标记总数和邮箱分布（只取最新5个用于展示）
	allTags, err := s.db.Tag.GetTagsWithEmailCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取标记统计失败",
			Error:   err.Error(),
		})
		return
	}
	totalTags := len(allTags)

	// 只取最新创建的5个标签用于分布展示
	tags := allTags
	if len(allTags) > 5 {
		tags = allTags[:5]
	}

	// 获取最近操作记录（仪表盘只显示最新5条）
	recentOperations, err := s.db.Log.GetRecentLogs(userID, 5, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取操作记录失败",
			Error:   err.Error(),
		})
		return
	}

	// 获取按标记分组的邮箱统计
	emailsByTag := make(map[string]int)
	for _, tag := range tags {
		emailsByTag[tag.Name] = tag.EmailCount
	}

	// 获取操作类型统计
	operationStats, err := s.db.Log.GetOperationStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取操作统计失败",
			Error:   err.Error(),
		})
		return
	}

	// 构建仪表盘数据
	dashboardStats := models.DashboardStats{
		TotalEmails:      totalEmails,
		TotalTags:        totalTags,
		RecentOperations: recentOperations,
		EmailsByTag:      emailsByTag,
		OperationsByType: operationStats,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取仪表盘数据成功",
		Data:    dashboardStats,
	})
}

// handleDashboardStats 获取详细统计数据
func (s *Server) handleDashboardStats(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取统计类型参数
	statsType := c.DefaultQuery("type", "all")

	var data interface{}
	var message string

	switch statsType {
	case "emails":
		// 邮箱相关统计
		totalEmails, err := s.emailService.CountUserEmails(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取邮箱统计失败",
				Error:   err.Error(),
			})
			return
		}

		// 获取最近添加的邮箱
		recentEmails, err := s.emailService.GetUserEmails(userID, 5, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取最近邮箱失败",
				Error:   err.Error(),
			})
			return
		}

		data = map[string]interface{}{
			"total_emails":  totalEmails,
			"recent_emails": recentEmails,
		}
		message = "获取邮箱统计成功"

	case "tags":
		// 标记相关统计
		tags, err := s.db.Tag.GetAllTags()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取标记统计失败",
				Error:   err.Error(),
			})
			return
		}

		tagStats := make([]map[string]interface{}, 0, len(tags))
		for _, tag := range tags {
			tagStats = append(tagStats, map[string]interface{}{
				"id":          tag.ID,
				"name":        tag.Name,
				"color":       tag.Color,
				"email_count": tag.EmailCount,
			})
		}

		data = map[string]interface{}{
			"total_tags": len(tags),
			"tag_stats":  tagStats,
		}
		message = "获取标记统计成功"

	case "operations":
		// 操作相关统计
		recentOperations, err := s.db.Log.GetRecentLogs(userID, 20, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取操作统计失败",
				Error:   err.Error(),
			})
			return
		}

		operationStats, err := s.db.Log.GetOperationStats(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取操作类型统计失败",
				Error:   err.Error(),
			})
			return
		}

		data = map[string]interface{}{
			"recent_operations": recentOperations,
			"operation_stats":   operationStats,
		}
		message = "获取操作统计成功"

	default:
		// 综合统计
		totalEmails, _ := s.emailService.CountUserEmails(userID)
		tags, _ := s.db.Tag.GetAllTags()
		operationStats, _ := s.db.Log.GetOperationStats(userID)

		data = map[string]interface{}{
			"total_emails":    totalEmails,
			"total_tags":      len(tags),
			"operation_stats": operationStats,
			"last_updated":    "2025-08-13T11:10:00Z",
		}
		message = "获取综合统计成功"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// handleGetEmails 获取邮箱列表
func (s *Server) handleGetEmails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取分页参数
	limit := 20
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 获取搜索关键词
	keyword := c.Query("keyword")

	var emails []models.Email
	var total int
	var err error

	if keyword != "" {
		emails, err = s.emailService.SearchEmails(userID, keyword, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取邮箱列表失败",
				Error:   err.Error(),
			})
			return
		}
		// 获取搜索结果总数
		total, err = s.emailService.CountSearchEmails(userID, keyword)
	} else {
		emails, err = s.emailService.GetUserEmails(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "获取邮箱列表失败",
				Error:   err.Error(),
			})
			return
		}
		// 获取用户邮箱总数
		total, err = s.emailService.CountUserEmails(userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取邮箱总数失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回分页格式的数据
	response := map[string]interface{}{
		"list":  emails,
		"total": total,
		"page":  (offset / limit) + 1,
		"size":  limit,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取邮箱列表成功",
		Data:    response,
	})
}

// handleAddEmail 添加邮箱
func (s *Server) handleAddEmail(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req models.AddEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 添加邮箱
	email, err := s.emailService.AddEmail(userID, &req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "添加邮箱失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "添加邮箱成功",
		Data:    email,
	})
}

// handleBatchAddEmails 批量添加邮箱
func (s *Server) handleBatchAddEmails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req models.BatchAddEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 检查数量限制
	if len(req.Emails) > 30 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "批量添加邮箱数量不能超过30个",
			Error:   "too many emails",
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 批量添加邮箱
	successEmails, errors, err := s.emailService.BatchAddEmails(userID, &req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "批量添加邮箱失败",
			Error:   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success_emails": successEmails,
		"errors":         errors,
		"success_count":  len(successEmails),
		"error_count":    len(errors),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "批量添加邮箱完成",
		Data:    response,
	})
}

// handleImportEmails 导入邮箱
func (s *Server) handleImportEmails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "文件上传失败",
			Error:   err.Error(),
		})
		return
	}

	// 检查文件类型
	if file.Header.Get("Content-Type") != "text/plain" && file.Header.Get("Content-Type") != "text/csv" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "不支持的文件类型，请上传txt或csv文件",
			Error:   "unsupported file type",
		})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "打开文件失败",
			Error:   err.Error(),
		})
		return
	}
	defer src.Close()

	// 读取文件内容
	content := make([]byte, file.Size)
	if _, err := src.Read(content); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "读取文件失败",
			Error:   err.Error(),
		})
		return
	}

	// 解析文件内容
	emails, parseErrors := s.parseEmailFile(string(content))
	if len(emails) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "文件中没有有效的邮箱数据",
			Error:   "no valid email data found",
		})
		return
	}

	// 检查数量限制
	if len(emails) > 30 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("导入邮箱数量不能超过30个，当前文件包含%d个邮箱", len(emails)),
			Error:   "too many emails",
		})
		return
	}

	// 批量添加邮箱
	req := &models.BatchAddEmailRequest{Emails: emails}
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	successEmails, errors, err := s.emailService.BatchAddEmails(userID, req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "批量添加邮箱失败",
			Error:   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success_emails": successEmails,
		"errors":         append(errors, parseErrors...),
		"success_count":  len(successEmails),
		"error_count":    len(errors) + len(parseErrors),
		"total_count":    len(emails) + len(parseErrors),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "邮箱导入完成",
		Data:    response,
	})
}

// handleGetLatestMail 获取最新邮件
func (s *Server) handleGetLatestMail(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取邮箱ID
	emailIDStr := c.Param("id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的邮箱ID",
			Error:   "invalid email id",
		})
		return
	}

	// 获取邮箱类型参数
	mailbox := c.DefaultQuery("mailbox", "INBOX")
	if mailbox != "INBOX" && mailbox != "Junk" {
		mailbox = "INBOX"
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 获取最新邮件
	mail, err := s.emailService.GetLatestMail(userID, emailID, mailbox, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "获取最新邮件失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取最新邮件成功",
		Data:    mail,
	})
}

// handleGetAllMails 获取全部邮件
func (s *Server) handleGetAllMails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取邮箱ID
	emailIDStr := c.Param("id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的邮箱ID",
			Error:   "invalid email id",
		})
		return
	}

	// 获取邮箱类型参数
	mailbox := c.DefaultQuery("mailbox", "INBOX")
	if mailbox != "INBOX" && mailbox != "Junk" {
		mailbox = "INBOX"
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 获取全部邮件
	mails, err := s.emailService.GetAllMails(userID, emailID, mailbox, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "获取全部邮件失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取全部邮件成功",
		Data:    mails,
	})
}

// handleClearInbox 清空收件箱
func (s *Server) handleClearInbox(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取邮箱ID
	emailIDStr := c.Param("id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的邮箱ID",
			Error:   "invalid email id",
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 清空收件箱
	if err := s.emailService.ClearInbox(userID, emailID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "清空收件箱失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "清空收件箱成功",
	})
}

// handleTagEmail 标记邮箱
func (s *Server) handleTagEmail(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取邮箱ID
	emailIDStr := c.Param("id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的邮箱ID",
			Error:   "invalid email id",
		})
		return
	}

	var req models.TagEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 检查邮箱是否属于当前用户
	_, err = s.emailService.GetEmailByID(userID, emailID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "邮箱不存在或无权访问",
			Error:   err.Error(),
		})
		return
	}

	// 为邮箱添加标记
	if err := s.db.Tag.AddEmailTag(emailID, req.TagID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "添加标记失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录操作日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogEmail(userID, "email_tagged", emailID,
		fmt.Sprintf("为邮箱添加标记，标记ID: %d", req.TagID),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "添加标记成功",
	})
}

// handleDeleteEmail 删除邮箱
func (s *Server) handleDeleteEmail(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取邮箱ID
	emailIDStr := c.Param("id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的邮箱ID",
			Error:   "invalid email id",
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 删除邮箱
	if err := s.emailService.DeleteEmail(userID, emailID, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "删除邮箱失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "删除邮箱成功",
	})
}

// handleGetTags 获取标记列表
func (s *Server) handleGetTags(c *gin.Context) {
	_, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取所有标记
	tags, err := s.db.Tag.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取标记列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取标记列表成功",
		Data:    tags,
	})
}

// handleCreateTag 创建标记
func (s *Server) handleCreateTag(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 检查标记名称是否已存在
	exists, err := s.db.Tag.TagExists(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "检查标记名称失败",
			Error:   err.Error(),
		})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "标记名称已存在",
			Error:   "tag name already exists",
		})
		return
	}

	// 创建标记
	tag := &models.Tag{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}

	createdTag, err := s.db.Tag.CreateTag(tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "创建标记失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录操作日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogTag(userID, "tag_created", createdTag.ID,
		fmt.Sprintf("创建标记: %s", req.Name),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "创建标记成功",
		Data:    createdTag,
	})
}

// handleUpdateTag 更新标记
func (s *Server) handleUpdateTag(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取标记ID
	tagIDStr := c.Param("id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的标记ID",
			Error:   "invalid tag id",
		})
		return
	}

	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取现有标记
	tag, err := s.db.Tag.GetTagByID(tagID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "标记不存在",
			Error:   err.Error(),
		})
		return
	}

	// 如果要更新名称，检查新名称是否已存在
	if req.Name != "" && req.Name != tag.Name {
		exists, err := s.db.Tag.TagExists(req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "检查标记名称失败",
				Error:   err.Error(),
			})
			return
		}
		if exists {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "标记名称已存在",
				Error:   "tag name already exists",
			})
			return
		}
		tag.Name = req.Name
	}

	// 更新其他字段
	if req.Description != "" {
		tag.Description = req.Description
	}
	if req.Color != "" {
		tag.Color = req.Color
	}

	// 保存更新
	if err := s.db.Tag.UpdateTag(tag); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "更新标记失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录操作日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogTag(userID, "tag_updated", tagID,
		fmt.Sprintf("更新标记: %s", tag.Name),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "更新标记成功",
		Data:    tag,
	})
}

// handleDeleteTag 删除标记
func (s *Server) handleDeleteTag(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取标记ID
	tagIDStr := c.Param("id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "无效的标记ID",
			Error:   "invalid tag id",
		})
		return
	}

	// 获取标记信息（用于日志）
	tag, err := s.db.Tag.GetTagByID(tagID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "标记不存在",
			Error:   err.Error(),
		})
		return
	}

	// 检查是否有邮箱使用此标记
	emailCount, err := s.db.Tag.GetTagEmailCount(tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "检查标记使用情况失败",
			Error:   err.Error(),
		})
		return
	}

	if emailCount > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("标记正在被 %d 个邮箱使用，请先解除关联", emailCount),
			Error:   "tag is in use",
		})
		return
	}

	// 删除标记
	if err := s.db.Tag.DeleteTag(tagID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "删除标记失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录操作日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogTag(userID, "tag_deleted", tagID,
		fmt.Sprintf("删除标记: %s", tag.Name),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "删除标记成功",
	})
}

// handleBatchTagEmails 批量标记邮箱
func (s *Server) handleBatchTagEmails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req models.TagEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证所有邮箱都属于当前用户
	for _, emailID := range req.EmailIDs {
		_, err := s.emailService.GetEmailByID(userID, emailID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: fmt.Sprintf("邮箱ID %d 不存在或无权访问", emailID),
				Error:   err.Error(),
			})
			return
		}
	}

	// 验证标记是否存在
	_, err := s.db.Tag.GetTagByID(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "标记不存在",
			Error:   err.Error(),
		})
		return
	}

	// 批量添加标记
	if err := s.db.Tag.BatchAddEmailTags(req.EmailIDs, req.TagID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "批量标记失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录操作日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogTag(userID, "batch_tag_emails", req.TagID,
		fmt.Sprintf("批量标记邮箱，邮箱数量: %d", len(req.EmailIDs)),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "批量标记成功",
		Data: map[string]interface{}{
			"tagged_count": len(req.EmailIDs),
			"tag_id":       req.TagID,
		},
	})
}

// handleBatchUntagEmails 批量取消标记邮箱
func (s *Server) handleBatchUntagEmails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req models.TagEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证所有邮箱都属于当前用户
	for _, emailID := range req.EmailIDs {
		_, err := s.emailService.GetEmailByID(userID, emailID)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: fmt.Sprintf("邮箱ID %d 不存在或无权访问", emailID),
				Error:   err.Error(),
			})
			return
		}
	}

	// 验证标记是否存在
	_, err := s.db.Tag.GetTagByID(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "标记不存在",
			Error:   err.Error(),
		})
		return
	}

	// 批量移除标记
	if err := s.db.Tag.BatchRemoveEmailTags(req.EmailIDs, req.TagID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "批量取消标记失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录操作日志
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogTag(userID, "batch_untag_emails", req.TagID,
		fmt.Sprintf("批量取消标记邮箱，邮箱数量: %d", len(req.EmailIDs)),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "批量取消标记成功",
		Data: map[string]interface{}{
			"untagged_count": len(req.EmailIDs),
			"tag_id":         req.TagID,
		},
	})
}

// parseEmailFile 解析邮箱文件
func (s *Server) parseEmailFile(content string) ([]models.AddEmailRequest, []string) {
	var emails []models.AddEmailRequest
	var errors []string

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 支持两种格式：
		// 1. 邮箱----密码----客户端ID----RefreshToken
		// 2. CSV格式：邮箱,密码,客户端ID,RefreshToken,备注
		var parts []string
		if strings.Contains(line, "----") {
			parts = strings.Split(line, "----")
		} else if strings.Contains(line, ",") {
			parts = strings.Split(line, ",")
		} else {
			errors = append(errors, fmt.Sprintf("第%d行格式错误: %s", i+1, line))
			continue
		}

		if len(parts) < 4 {
			errors = append(errors, fmt.Sprintf("第%d行数据不完整: %s", i+1, line))
			continue
		}

		email := models.AddEmailRequest{
			EmailAddress: strings.TrimSpace(parts[0]),
			Password:     strings.TrimSpace(parts[1]),
			ClientID:     strings.TrimSpace(parts[2]),
			RefreshToken: strings.TrimSpace(parts[3]),
		}

		// 如果有第5个字段，作为备注
		if len(parts) > 4 {
			email.Remark = strings.TrimSpace(parts[4])
		}

		// 基本验证
		if email.EmailAddress == "" || email.Password == "" || email.ClientID == "" || email.RefreshToken == "" {
			errors = append(errors, fmt.Sprintf("第%d行必填字段为空: %s", i+1, line))
			continue
		}

		emails = append(emails, email)
	}

	return emails, errors
}

// handleBatchDeleteEmails 批量删除邮箱
func (s *Server) handleBatchDeleteEmails(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req struct {
		EmailIDs []int `json:"email_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if len(req.EmailIDs) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "邮箱ID列表不能为空",
			Error:   "email_ids cannot be empty",
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 批量删除邮箱
	if err := s.emailService.BatchDeleteEmails(userID, req.EmailIDs, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "批量删除邮箱失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("成功删除 %d 个邮箱", len(req.EmailIDs)),
	})
}

// handleBatchClearInbox 批量清空收件箱
func (s *Server) handleBatchClearInbox(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	var req struct {
		EmailIDs []int `json:"email_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	if len(req.EmailIDs) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "邮箱ID列表不能为空",
			Error:   "email_ids cannot be empty",
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 批量清空收件箱
	successCount, errors, err := s.emailService.BatchClearInbox(userID, req.EmailIDs, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "批量清空收件箱失败",
			Error:   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success_count": successCount,
		"error_count":   len(errors),
		"errors":        errors,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("批量清空收件箱完成，成功: %d, 失败: %d", successCount, len(errors)),
		Data:    response,
	})
}

// handleGetLogs 获取操作日志（分页）
func (s *Server) handleGetLogs(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "5"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 5
	}

	offset := (page - 1) * pageSize

	// 获取操作日志
	logs, err := s.db.Log.GetRecentLogs(userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取操作日志失败",
			Error:   err.Error(),
		})
		return
	}

	// 获取总数
	total, err := s.db.Log.GetLogCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "获取日志总数失败",
			Error:   err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "获取操作日志成功",
		Data: map[string]interface{}{
			"logs":        logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// handleClearLogs 清空所有操作日志
func (s *Server) handleClearLogs(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "未认证",
			Error:   "user not authenticated",
		})
		return
	}

	// 清空操作日志
	if err := s.db.Log.ClearAllLogs(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "清空操作日志失败",
			Error:   err.Error(),
		})
		return
	}

	// 记录清空日志的操作
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	s.db.Log.LogAuth(userID, "clear_all_logs", "清空所有操作日志", ipAddress, userAgent)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "操作日志已清空",
	})
}

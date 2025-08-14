package auth

import (
	"net/http"
	"strings"

	"outlook-helper/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "缺少认证令牌",
				Error:   "missing authorization header",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "无效的认证格式",
				Error:   "invalid authorization format",
			})
			c.Abort()
			return
		}

		// 提取令牌
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "令牌为空",
				Error:   "empty token",
			})
			c.Abort()
			return
		}

		// 验证令牌
		user, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "无效的认证令牌",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求认证）
func OptionalAuthMiddleware(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.Next()
			return
		}

		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			c.Next()
			return
		}

		// 尝试验证令牌
		user, err := authService.ValidateToken(tokenString)
		if err == nil {
			// 令牌有效，设置用户信息
			c.Set("user", user)
			c.Set("user_id", user.ID)
			c.Set("username", user.Username)
		}

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "需要认证",
				Error:   "authentication required",
			})
			c.Abort()
			return
		}

		userObj, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "用户信息错误",
				Error:   "invalid user object",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员（这里简单地检查用户名是否为admin）
		// 在实际项目中，应该有专门的角色权限系统
		if userObj.Username != "admin" {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "需要管理员权限",
				Error:   "admin privileges required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userObj, ok := user.(*models.User)
	if !ok {
		return nil, false
	}

	return userObj, true
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	if !ok {
		return 0, false
	}

	return id, true
}

// GetCurrentUsername 从上下文中获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	if !ok {
		return "", false
	}

	return name, true
}

// RequireAuth 检查是否已认证的辅助函数
func RequireAuth(c *gin.Context) (*models.User, error) {
	user, exists := GetCurrentUser(c)
	if !exists {
		return nil, gin.Error{
			Err:  http.ErrAbortHandler,
			Type: gin.ErrorTypePublic,
			Meta: "authentication required",
		}
	}
	return user, nil
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 简单的速率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
	// 这里可以实现基于IP的速率限制
	// 为了简化，暂时返回空的中间件
	return func(c *gin.Context) {
		c.Next()
	}
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return ""
	})
}

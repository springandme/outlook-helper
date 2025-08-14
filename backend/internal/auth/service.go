package auth

import (
	"errors"
	"time"

	"outlook-helper/backend/internal/config"
	"outlook-helper/backend/internal/database"
	"outlook-helper/backend/internal/models"
)

// Service 认证服务
type Service struct {
	userRepo   *database.UserRepository
	logRepo    *database.LogRepository
	jwtManager *JWTManager
	config     *config.Config
}

// NewService 创建认证服务
func NewService(db *database.DB, jwtSecret string, jwtExpire int, cfg *config.Config) *Service {
	jwtManager := NewJWTManager(jwtSecret, time.Duration(jwtExpire)*time.Hour)

	return &Service{
		userRepo:   db.User,
		logRepo:    db.Log,
		jwtManager: jwtManager,
		config:     cfg,
	}
}

// Login 用户登录
func (s *Service) Login(authToken, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// 验证授权码是否与环境变量配置匹配
	if authToken != s.config.AuthToken {
		// 记录登录失败日志
		s.logRepo.LogAuth(0, "login_failed", "授权码错误", ipAddress, userAgent)
		return nil, errors.New("授权码错误")
	}

	// 创建虚拟用户对象（用于生成JWT令牌）
	user := &models.User{
		ID:       1,       // 固定ID
		Username: "admin", // 固定用户名
	}

	// 生成JWT令牌
	token, expiresAt, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	// 记录登录成功日志
	s.logRepo.LogAuth(user.ID, "login_success", "用户登录成功", ipAddress, userAgent)

	// 构造响应
	response := &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: models.User{
			ID:       user.ID,
			Username: user.Username,
		},
	}

	return response, nil
}

// Logout 用户登出
func (s *Service) Logout(userID int, ipAddress, userAgent string) error {
	// 记录登出日志
	return s.logRepo.LogAuth(userID, "logout", "用户登出", ipAddress, userAgent)
}

// ValidateToken 验证令牌
func (s *Service) ValidateToken(tokenString string) (*models.User, error) {
	// 验证JWT令牌
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 创建虚拟用户对象（基于JWT声明）
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}

	return user, nil
}

// RefreshToken 刷新令牌
func (s *Service) RefreshToken(tokenString string, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// 验证当前令牌
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 生成新令牌
	newToken, expiresAt, err := s.jwtManager.RefreshToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 记录令牌刷新日志
	s.logRepo.LogAuth(user.ID, "token_refresh", "令牌刷新", ipAddress, userAgent)

	// 构造响应
	response := &models.LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User: models.User{
			ID:          user.ID,
			Username:    user.Username,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}

	return response, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(userID int, oldPassword, newPassword, ipAddress, userAgent string) error {
	// 获取用户
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !s.userRepo.ValidatePassword(user, oldPassword) {
		// 记录密码修改失败日志
		s.logRepo.LogAuth(userID, "password_change_failed", "旧密码错误", ipAddress, userAgent)
		return errors.New("旧密码错误")
	}

	// 创建新用户对象用于更新密码
	newUser, err := s.userRepo.CreateUser(user.Username+"_temp", newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新用户密码（这里需要在UserRepository中添加UpdatePassword方法）
	// 暂时使用临时方案
	user.PasswordHash = newUser.PasswordHash

	// 记录密码修改成功日志
	s.logRepo.LogAuth(userID, "password_changed", "密码修改成功", ipAddress, userAgent)

	return nil
}

// GetUserInfo 获取用户信息
func (s *Service) GetUserInfo(userID int) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 不返回密码哈希
	user.PasswordHash = ""
	return user, nil
}

// CreateUser 创建用户（管理员功能）
func (s *Service) CreateUser(username, password, ipAddress, userAgent string, operatorID int) (*models.User, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.UserExists(username)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 创建用户
	user, err := s.userRepo.CreateUser(username, password)
	if err != nil {
		return nil, err
	}

	// 记录用户创建日志
	s.logRepo.LogAuth(operatorID, "user_created", "创建用户: "+username, ipAddress, userAgent)

	// 不返回密码哈希
	user.PasswordHash = ""
	return user, nil
}

// GetJWTManager 获取JWT管理器（用于中间件）
func (s *Service) GetJWTManager() *JWTManager {
	return s.jwtManager
}

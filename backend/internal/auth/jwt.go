package auth

import (
	"errors"
	"time"

	"outlook-helper/backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken 生成JWT令牌
func (manager *JWTManager) GenerateToken(user *models.User) (string, int64, error) {
	expiresAt := time.Now().Add(manager.tokenDuration)
	
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "outlook-helper",
			Subject:   "user-auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(manager.secretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken 验证JWT令牌
func (manager *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 检查令牌是否过期
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// RefreshToken 刷新JWT令牌
func (manager *JWTManager) RefreshToken(tokenString string) (string, int64, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return "", 0, err
	}

	// 检查令牌是否在刷新窗口内（例如，过期前30分钟内可以刷新）
	refreshWindow := 30 * time.Minute
	if time.Until(claims.ExpiresAt.Time) > refreshWindow {
		return "", 0, errors.New("token not eligible for refresh")
	}

	// 创建新的用户对象用于生成新令牌
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}

	return manager.GenerateToken(user)
}

// ExtractUserID 从令牌中提取用户ID
func (manager *JWTManager) ExtractUserID(tokenString string) (int, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// ExtractUsername 从令牌中提取用户名
func (manager *JWTManager) ExtractUsername(tokenString string) (string, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}

// IsTokenExpired 检查令牌是否过期
func (manager *JWTManager) IsTokenExpired(tokenString string) bool {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return true
	}
	return claims.ExpiresAt.Time.Before(time.Now())
}

// GetTokenRemainingTime 获取令牌剩余有效时间
func (manager *JWTManager) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0, errors.New("token expired")
	}
	
	return remaining, nil
}

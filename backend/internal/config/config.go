package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	AppName                string
	Port                   string
	Env                    string
	DBPath                 string
	JWTSecret              string
	JWTExpire              int
	OutlookAPI             string
	LogLevel               string
	LogFile                string
	CORSOrigins            string
	SkipEmailValidation    bool   // 是否跳过邮箱验证（调试用）
	EmailValidationWorkers int    // 邮箱验证并发数
	AuthToken              string // 授权码（必须配置）
}

// Load 加载配置
func Load() (*Config, error) {
	// 尝试加载 .env 文件
	_ = godotenv.Load()

	// 检查必须的环境变量
	authToken := os.Getenv("AUTH_TOKEN")
	outlookAPI := os.Getenv("OUTLOOK_API_BASE_URL")

	var missingVars []string
	if authToken == "" {
		missingVars = append(missingVars, "AUTH_TOKEN")
	}
	if outlookAPI == "" {
		missingVars = append(missingVars, "OUTLOOK_API_BASE_URL")
	}

	if len(missingVars) > 0 {
		fmt.Println("❌ 以下必须的环境变量未配置：")
		for _, varName := range missingVars {
			fmt.Printf("   - %s\n", varName)
		}
		fmt.Println()
		fmt.Println("请设置以下环境变量后重新启动应用：")
		if authToken == "" {
			fmt.Println("AUTH_TOKEN=your-super-secret-auth-token")
		}
		if outlookAPI == "" {
			fmt.Println("OUTLOOK_API_BASE_URL=https://your-outlook-api-domain.vercel.app")
		}
		fmt.Println()
		fmt.Println("应用将停止运行...")

		// 无限等待，不退出程序
		select {}
	}

	cfg := &Config{
		AppName:                getEnv("APP_NAME", "Outlook取件助手"),
		Port:                   getEnv("APP_PORT", "8080"),
		Env:                    getEnv("APP_ENV", "development"),
		DBPath:                 getEnv("DB_PATH", "./data/outlook_helper.db"),
		JWTSecret:              getEnv("JWT_SECRET", "default-secret-change-this"),
		JWTExpire:              getEnvAsInt("JWT_EXPIRE_HOURS", 6),
		OutlookAPI:             outlookAPI,
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		LogFile:                getEnv("LOG_FILE", "./logs/app.log"),
		CORSOrigins:            getEnv("CORS_ALLOWED_ORIGINS", ""),
		SkipEmailValidation:    getEnvAsBool("SKIP_EMAIL_VALIDATION", false),
		EmailValidationWorkers: getEnvAsInt("EMAIL_VALIDATION_WORKERS", 5),
		AuthToken:              authToken,
	}

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

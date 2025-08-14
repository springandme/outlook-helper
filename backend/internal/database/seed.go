package database

import (
	"database/sql"
	"log"

	"outlook-helper/backend/internal/models"
)

// SeedData 初始化种子数据
func SeedData(db *sql.DB) error {
	dbManager := NewDB(db)
	
	// 检查是否已有用户
	users, err := dbManager.User.GetAllUsers()
	if err != nil {
		return err
	}
	
	// 如果没有用户，创建默认管理员用户
	if len(users) == 0 {
		log.Println("Creating default admin user...")
		
		// 创建默认管理员用户
		adminUser, err := dbManager.User.CreateUser("admin", "admin123")
		if err != nil {
			return err
		}
		
		log.Printf("Default admin user created with ID: %d", adminUser.ID)
		
		// 创建一些默认标记
		if err := createDefaultTags(dbManager); err != nil {
			return err
		}
		
		log.Println("Default tags created successfully")
	}
	
	return nil
}

// createDefaultTags 创建默认标记
func createDefaultTags(db *DB) error {
	defaultTags := []models.Tag{
		{
			Name:        "工作邮箱",
			Description: "用于工作相关的邮箱",
			Color:       "#007bff",
		},
		{
			Name:        "个人邮箱",
			Description: "个人使用的邮箱",
			Color:       "#28a745",
		},
		{
			Name:        "测试邮箱",
			Description: "用于测试的邮箱",
			Color:       "#ffc107",
		},
		{
			Name:        "重要邮箱",
			Description: "重要的邮箱账户",
			Color:       "#dc3545",
		},
		{
			Name:        "临时邮箱",
			Description: "临时使用的邮箱",
			Color:       "#6c757d",
		},
	}
	
	for _, tag := range defaultTags {
		// 检查标记是否已存在
		exists, err := db.Tag.TagExists(tag.Name)
		if err != nil {
			return err
		}
		
		if !exists {
			_, err := db.Tag.CreateTag(&tag)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// CleanupOldData 清理旧数据
func CleanupOldData(db *sql.DB) error {
	dbManager := NewDB(db)
	
	// 清理30天前的操作日志
	err := dbManager.Log.DeleteOldLogs(30)
	if err != nil {
		log.Printf("Failed to cleanup old logs: %v", err)
		return err
	}
	
	log.Println("Old logs cleaned up successfully")
	return nil
}

// GetDatabaseStats 获取数据库统计信息
func GetDatabaseStats(db *sql.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 统计各表的记录数
	tables := []string{"users", "emails", "tags", "email_tags", "operation_logs"}
	
	for _, table := range tables {
		query := "SELECT COUNT(*) FROM " + table
		var count int
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, err
		}
		stats[table] = count
	}
	
	// 获取数据库文件大小
	var pageCount, pageSize int
	err := db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return nil, err
	}
	
	err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return nil, err
	}
	
	stats["database_size_bytes"] = pageCount * pageSize
	stats["database_size_mb"] = float64(pageCount*pageSize) / (1024 * 1024)
	
	return stats, nil
}

// OptimizeDatabase 优化数据库
func OptimizeDatabase(db *sql.DB) error {
	// 执行VACUUM命令清理数据库
	_, err := db.Exec("VACUUM")
	if err != nil {
		return err
	}
	
	// 分析数据库以优化查询计划
	_, err = db.Exec("ANALYZE")
	if err != nil {
		return err
	}
	
	log.Println("Database optimized successfully")
	return nil
}

// BackupDatabase 备份数据库（简单的SQL导出）
func BackupDatabase(db *sql.DB, backupPath string) error {
	// 这里可以实现数据库备份逻辑
	// 由于SQLite的特性，可以直接复制数据库文件
	// 或者导出SQL语句
	
	log.Printf("Database backup functionality not implemented yet. Target path: %s", backupPath)
	return nil
}

// CheckDatabaseIntegrity 检查数据库完整性
func CheckDatabaseIntegrity(db *sql.DB) error {
	var result string
	err := db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return err
	}
	
	if result != "ok" {
		log.Printf("Database integrity check failed: %s", result)
		return err
	}
	
	log.Println("Database integrity check passed")
	return nil
}

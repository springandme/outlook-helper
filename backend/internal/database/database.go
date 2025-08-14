package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB 数据库管理器
type DB struct {
	conn *sql.DB

	// Repository instances
	User  *UserRepository
	Email *EmailRepository
	Tag   *TagRepository
	Log   *LogRepository
}

// NewDB 创建数据库管理器
func NewDB(conn *sql.DB) *DB {
	return &DB{
		conn:  conn,
		User:  NewUserRepository(conn),
		Email: NewEmailRepository(conn),
		Tag:   NewTagRepository(conn),
		Log:   NewLogRepository(conn),
	}
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConnection 获取原始数据库连接
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// Initialize 初始化数据库连接
func Initialize(dbPath string) (*sql.DB, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate 运行数据库迁移
func Migrate(db *sql.DB) error {
	// 创建用户表
	if err := createUsersTable(db); err != nil {
		return err
	}

	// 创建邮箱表
	if err := createEmailsTable(db); err != nil {
		return err
	}

	// 创建标记表
	if err := createTagsTable(db); err != nil {
		return err
	}

	// 创建邮箱标记关联表
	if err := createEmailTagsTable(db); err != nil {
		return err
	}

	// 创建操作日志表
	if err := createOperationLogsTable(db); err != nil {
		return err
	}

	return nil
}

// createUsersTable 创建用户表
func createUsersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		last_login_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	return err
}

// createEmailsTable 创建邮箱表
func createEmailsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS emails (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		email_address VARCHAR(255) NOT NULL,
		password VARCHAR(255) NOT NULL,
		client_id VARCHAR(255) NOT NULL,
		refresh_token TEXT NOT NULL,
		remark TEXT,
		last_operation_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(email_address, user_id)
	)`
	_, err := db.Exec(query)
	return err
}

// createTagsTable 创建标记表
func createTagsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(50) NOT NULL,
		description TEXT,
		color VARCHAR(7) DEFAULT '#007bff',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name)
	)`
	_, err := db.Exec(query)
	return err
}

// createEmailTagsTable 创建邮箱标记关联表
func createEmailTagsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS email_tags (
		email_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (email_id, tag_id),
		FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	)`
	_, err := db.Exec(query)
	return err
}

// createOperationLogsTable 创建操作日志表
func createOperationLogsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		operation_type VARCHAR(50) NOT NULL,
		target_type VARCHAR(50) NOT NULL,
		target_id INTEGER,
		description TEXT,
		ip_address VARCHAR(45),
		user_agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`
	_, err := db.Exec(query)
	return err
}

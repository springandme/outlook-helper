package database

import (
	"database/sql"

	"outlook-helper/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// UserRepository 用户数据库操作
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser 创建用户
func (r *UserRepository) CreateUser(username, password string) (*models.User, error) {
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (username, password_hash, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(query, username, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetUserByID(int(id))
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, last_login_at, created_at, updated_at
		FROM users WHERE id = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, last_login_at, created_at, updated_at
		FROM users WHERE username = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// ValidatePassword 验证密码
func (r *UserRepository) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(userID int) error {
	query := `
		UPDATE users 
		SET last_login_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, userID)
	return err
}

// UserExists 检查用户是否存在
func (r *UserRepository) UserExists(username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = ?`

	var count int
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetAllUsers 获取所有用户（管理功能）
func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, password_hash, last_login_at, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

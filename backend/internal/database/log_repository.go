package database

import (
	"database/sql"

	"outlook-helper/backend/internal/constants"
	"outlook-helper/backend/internal/models"
)

// LogRepository 操作日志数据库操作
type LogRepository struct {
	db *sql.DB
}

// NewLogRepository 创建日志仓库
func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

// CreateLog 创建操作日志
func (r *LogRepository) CreateLog(log *models.OperationLog) error {
	query := `
		INSERT INTO operation_logs (user_id, operation_type, target_type, target_id, description, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := r.db.Exec(query,
		log.UserID,
		log.OperationType,
		log.TargetType,
		log.TargetID,
		log.Description,
		log.IPAddress,
		log.UserAgent,
	)

	return err
}

// GetLogsByUserID 根据用户ID获取操作日志
func (r *LogRepository) GetLogsByUserID(userID int, limit, offset int) ([]models.OperationLog, error) {
	query := `
		SELECT id, user_id, operation_type, target_type, target_id, description, ip_address, user_agent, created_at
		FROM operation_logs 
		WHERE user_id = ? 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.OperationLog
	for rows.Next() {
		var log models.OperationLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.OperationType,
			&log.TargetType,
			&log.TargetID,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetRecentLogs 获取最近的操作日志（分页）
func (r *LogRepository) GetRecentLogs(userID int, limit, offset int) ([]models.OperationLog, error) {
	query := `
		SELECT id, user_id, operation_type, target_type, target_id, description, ip_address, user_agent, created_at
		FROM operation_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.OperationLog
	for rows.Next() {
		var log models.OperationLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.OperationType,
			&log.TargetType,
			&log.TargetID,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		// 添加操作类型中文名称
		log.OperationName = constants.GetOperationTypeName(log.OperationType)
		logs = append(logs, log)
	}

	return logs, nil
}

// GetLogCount 获取用户操作日志总数
func (r *LogRepository) GetLogCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM operation_logs WHERE user_id = ?`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// ClearAllLogs 清空用户的所有操作日志
func (r *LogRepository) ClearAllLogs(userID int) error {
	query := `DELETE FROM operation_logs WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// GetLogsByType 根据操作类型获取日志
func (r *LogRepository) GetLogsByType(userID int, operationType string, limit, offset int) ([]models.OperationLog, error) {
	query := `
		SELECT id, user_id, operation_type, target_type, target_id, description, ip_address, user_agent, created_at
		FROM operation_logs 
		WHERE user_id = ? AND operation_type = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, operationType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.OperationLog
	for rows.Next() {
		var log models.OperationLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.OperationType,
			&log.TargetType,
			&log.TargetID,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetOperationStats 获取操作统计
func (r *LogRepository) GetOperationStats(userID int) (map[string]int, error) {
	query := `
		SELECT operation_type, COUNT(*) as count
		FROM operation_logs
		WHERE user_id = ?
		GROUP BY operation_type
		ORDER BY count DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var operationType string
		var count int
		err := rows.Scan(&operationType, &count)
		if err != nil {
			return nil, err
		}
		// 使用中文名称作为key
		operationName := constants.GetOperationTypeName(operationType)
		stats[operationName] = count
	}

	return stats, nil
}

// DeleteOldLogs 删除旧日志（清理功能）
func (r *LogRepository) DeleteOldLogs(days int) error {
	query := `
		DELETE FROM operation_logs 
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`

	_, err := r.db.Exec(query, days)
	return err
}

// CountLogsByUserID 统计用户日志数量
func (r *LogRepository) CountLogsByUserID(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM operation_logs WHERE user_id = ?`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// LogEmail 记录邮箱相关操作
func (r *LogRepository) LogEmail(userID int, operation string, emailID int, description, ipAddress, userAgent string) error {
	log := &models.OperationLog{
		UserID:        userID,
		OperationType: operation,
		TargetType:    "email",
		TargetID:      &emailID,
		Description:   description,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}
	return r.CreateLog(log)
}

// LogTag 记录标记相关操作
func (r *LogRepository) LogTag(userID int, operation string, tagID int, description, ipAddress, userAgent string) error {
	log := &models.OperationLog{
		UserID:        userID,
		OperationType: operation,
		TargetType:    "tag",
		TargetID:      &tagID,
		Description:   description,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}
	return r.CreateLog(log)
}

// LogAuth 记录认证相关操作
func (r *LogRepository) LogAuth(userID int, operation, description, ipAddress, userAgent string) error {
	log := &models.OperationLog{
		UserID:        userID,
		OperationType: operation,
		TargetType:    "auth",
		TargetID:      nil,
		Description:   description,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}
	return r.CreateLog(log)
}

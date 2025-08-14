package database

import (
	"database/sql"

	"outlook-helper/backend/internal/models"
)

// EmailRepository 邮箱数据库操作
type EmailRepository struct {
	db *sql.DB
}

// NewEmailRepository 创建邮箱仓库
func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// CreateEmail 创建邮箱
func (r *EmailRepository) CreateEmail(email *models.Email) (*models.Email, error) {
	query := `
		INSERT INTO emails (user_id, email_address, password, client_id, refresh_token, remark, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(query,
		email.UserID,
		email.EmailAddress,
		email.Password,
		email.ClientID,
		email.RefreshToken,
		email.Remark,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetEmailByID(int(id))
}

// GetEmailByID 根据ID获取邮箱
func (r *EmailRepository) GetEmailByID(id int) (*models.Email, error) {
	query := `
		SELECT id, user_id, email_address, password, client_id, refresh_token, remark, 
		       last_operation_at, created_at, updated_at
		FROM emails WHERE id = ?
	`

	email := &models.Email{}
	err := r.db.QueryRow(query, id).Scan(
		&email.ID,
		&email.UserID,
		&email.EmailAddress,
		&email.Password,
		&email.ClientID,
		&email.RefreshToken,
		&email.Remark,
		&email.LastOperationAt,
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// 加载标记
	tags, err := r.GetEmailTags(email.ID)
	if err == nil {
		email.Tags = tags
	}

	return email, nil
}

// GetEmailsByUserID 根据用户ID获取邮箱列表
func (r *EmailRepository) GetEmailsByUserID(userID int, limit, offset int) ([]models.Email, error) {
	query := `
		SELECT id, user_id, email_address, password, client_id, refresh_token, remark, 
		       last_operation_at, created_at, updated_at
		FROM emails 
		WHERE user_id = ? 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []models.Email
	for rows.Next() {
		var email models.Email
		err := rows.Scan(
			&email.ID,
			&email.UserID,
			&email.EmailAddress,
			&email.Password,
			&email.ClientID,
			&email.RefreshToken,
			&email.Remark,
			&email.LastOperationAt,
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 加载标记
		tags, err := r.GetEmailTags(email.ID)
		if err == nil {
			email.Tags = tags
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// SearchEmails 搜索邮箱
func (r *EmailRepository) SearchEmails(userID int, keyword string, limit, offset int) ([]models.Email, error) {
	query := `
		SELECT id, user_id, email_address, password, client_id, refresh_token, remark, 
		       last_operation_at, created_at, updated_at
		FROM emails 
		WHERE user_id = ? AND (email_address LIKE ? OR remark LIKE ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	searchPattern := "%" + keyword + "%"
	rows, err := r.db.Query(query, userID, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []models.Email
	for rows.Next() {
		var email models.Email
		err := rows.Scan(
			&email.ID,
			&email.UserID,
			&email.EmailAddress,
			&email.Password,
			&email.ClientID,
			&email.RefreshToken,
			&email.Remark,
			&email.LastOperationAt,
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 加载标记
		tags, err := r.GetEmailTags(email.ID)
		if err == nil {
			email.Tags = tags
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// UpdateEmail 更新邮箱
func (r *EmailRepository) UpdateEmail(email *models.Email) error {
	query := `
		UPDATE emails 
		SET email_address = ?, password = ?, client_id = ?, refresh_token = ?, remark = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		email.EmailAddress,
		email.Password,
		email.ClientID,
		email.RefreshToken,
		email.Remark,
		email.ID,
	)

	return err
}

// UpdateLastOperation 更新最后操作时间
func (r *EmailRepository) UpdateLastOperation(emailID int) error {
	query := `
		UPDATE emails 
		SET last_operation_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, emailID)
	return err
}

// DeleteEmail 删除邮箱
func (r *EmailRepository) DeleteEmail(id int) error {
	query := `DELETE FROM emails WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// BatchDeleteEmails 批量删除邮箱
func (r *EmailRepository) BatchDeleteEmails(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 先删除相关的标签关联
	tagQuery := `DELETE FROM email_tags WHERE email_id = ?`
	tagStmt, err := tx.Prepare(tagQuery)
	if err != nil {
		return err
	}
	defer tagStmt.Close()

	for _, id := range ids {
		_, err := tagStmt.Exec(id)
		if err != nil {
			return err
		}
	}

	// 再删除邮箱
	emailQuery := `DELETE FROM emails WHERE id = ?`
	emailStmt, err := tx.Prepare(emailQuery)
	if err != nil {
		return err
	}
	defer emailStmt.Close()

	for _, id := range ids {
		_, err := emailStmt.Exec(id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchCreateEmails 批量创建邮箱
func (r *EmailRepository) BatchCreateEmails(emails []*models.Email) ([]models.Email, error) {
	if len(emails) == 0 {
		return []models.Email{}, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO emails (user_id, email_address, password, client_id, refresh_token, remark, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var createdEmails []models.Email
	for _, email := range emails {
		result, err := stmt.Exec(
			email.UserID,
			email.EmailAddress,
			email.Password,
			email.ClientID,
			email.RefreshToken,
			email.Remark,
		)
		if err != nil {
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}

		// 创建返回的邮箱对象
		createdEmail := models.Email{
			ID:           int(id),
			UserID:       email.UserID,
			EmailAddress: email.EmailAddress,
			Remark:       email.Remark,
			CreatedAt:    email.CreatedAt,
			UpdatedAt:    email.UpdatedAt,
		}
		createdEmails = append(createdEmails, createdEmail)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return createdEmails, nil
}

// GetEmailTags 获取邮箱的标记
func (r *EmailRepository) GetEmailTags(emailID int) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.description, t.color, t.created_at, t.updated_at
		FROM tags t
		INNER JOIN email_tags et ON t.id = et.tag_id
		WHERE et.email_id = ?
		ORDER BY t.name
	`

	rows, err := r.db.Query(query, emailID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.Description,
			&tag.Color,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// CountEmailsByUserID 统计用户邮箱数量
func (r *EmailRepository) CountEmailsByUserID(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM emails WHERE user_id = ?`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// CountSearchEmails 统计搜索结果数量
func (r *EmailRepository) CountSearchEmails(userID int, keyword string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM emails
		WHERE user_id = ? AND (email_address LIKE ? OR remark LIKE ?)
	`

	searchPattern := "%" + keyword + "%"
	var count int
	err := r.db.QueryRow(query, userID, searchPattern, searchPattern).Scan(&count)
	return count, err
}

// EmailExists 检查邮箱是否已存在
func (r *EmailRepository) EmailExists(userID int, emailAddress string) (bool, error) {
	query := `SELECT COUNT(*) FROM emails WHERE user_id = ? AND email_address = ?`

	var count int
	err := r.db.QueryRow(query, userID, emailAddress).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

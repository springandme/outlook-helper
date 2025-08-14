package database

import (
	"database/sql"

	"outlook-helper/backend/internal/models"
)

// TagRepository 标记数据库操作
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository 创建标记仓库
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

// CreateTag 创建标记
func (r *TagRepository) CreateTag(tag *models.Tag) (*models.Tag, error) {
	query := `
		INSERT INTO tags (name, description, color, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(query, tag.Name, tag.Description, tag.Color)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetTagByID(int(id))
}

// GetTagByID 根据ID获取标记
func (r *TagRepository) GetTagByID(id int) (*models.Tag, error) {
	query := `
		SELECT id, name, description, color, created_at, updated_at
		FROM tags WHERE id = ?
	`

	tag := &models.Tag{}
	err := r.db.QueryRow(query, id).Scan(
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

	// 获取关联的邮箱数量
	count, err := r.GetTagEmailCount(tag.ID)
	if err == nil {
		tag.EmailCount = count
	}

	return tag, nil
}

// GetTagByName 根据名称获取标记
func (r *TagRepository) GetTagByName(name string) (*models.Tag, error) {
	query := `
		SELECT id, name, description, color, created_at, updated_at
		FROM tags WHERE name = ?
	`

	tag := &models.Tag{}
	err := r.db.QueryRow(query, name).Scan(
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

	// 获取关联的邮箱数量
	count, err := r.GetTagEmailCount(tag.ID)
	if err == nil {
		tag.EmailCount = count
	}

	return tag, nil
}

// GetAllTags 获取所有标记
func (r *TagRepository) GetAllTags() ([]models.Tag, error) {
	query := `
		SELECT id, name, description, color, created_at, updated_at
		FROM tags ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
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

		// 获取关联的邮箱数量
		count, err := r.GetTagEmailCount(tag.ID)
		if err == nil {
			tag.EmailCount = count
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// UpdateTag 更新标记
func (r *TagRepository) UpdateTag(tag *models.Tag) error {
	query := `
		UPDATE tags 
		SET name = ?, description = ?, color = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query, tag.Name, tag.Description, tag.Color, tag.ID)
	return err
}

// DeleteTag 删除标记
func (r *TagRepository) DeleteTag(id int) error {
	// 先删除关联关系
	if err := r.RemoveAllEmailTags(id); err != nil {
		return err
	}

	// 再删除标记
	query := `DELETE FROM tags WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// TagExists 检查标记是否存在
func (r *TagRepository) TagExists(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM tags WHERE name = ?`

	var count int
	err := r.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AddEmailTag 为邮箱添加标记
func (r *TagRepository) AddEmailTag(emailID, tagID int) error {
	query := `
		INSERT OR IGNORE INTO email_tags (email_id, tag_id, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err := r.db.Exec(query, emailID, tagID)
	return err
}

// RemoveEmailTag 移除邮箱标记
func (r *TagRepository) RemoveEmailTag(emailID, tagID int) error {
	query := `DELETE FROM email_tags WHERE email_id = ? AND tag_id = ?`
	_, err := r.db.Exec(query, emailID, tagID)
	return err
}

// RemoveAllEmailTags 移除标记的所有邮箱关联
func (r *TagRepository) RemoveAllEmailTags(tagID int) error {
	query := `DELETE FROM email_tags WHERE tag_id = ?`
	_, err := r.db.Exec(query, tagID)
	return err
}

// GetTagEmailCount 获取标记关联的邮箱数量
func (r *TagRepository) GetTagEmailCount(tagID int) (int, error) {
	query := `SELECT COUNT(*) FROM email_tags WHERE tag_id = ?`

	var count int
	err := r.db.QueryRow(query, tagID).Scan(&count)
	return count, err
}

// GetEmailsByTag 根据标记获取邮箱列表
func (r *TagRepository) GetEmailsByTag(tagID int, limit, offset int) ([]models.Email, error) {
	query := `
		SELECT e.id, e.user_id, e.email_address, e.password, e.client_id, e.refresh_token, e.remark, 
		       e.last_operation_at, e.created_at, e.updated_at
		FROM emails e
		INNER JOIN email_tags et ON e.id = et.email_id
		WHERE et.tag_id = ?
		ORDER BY e.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, tagID, limit, offset)
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
		emails = append(emails, email)
	}

	return emails, nil
}

// BatchAddEmailTags 批量为邮箱添加标记
func (r *TagRepository) BatchAddEmailTags(emailIDs []int, tagID int) error {
	if len(emailIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT OR IGNORE INTO email_tags (email_id, tag_id, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, emailID := range emailIDs {
		_, err := stmt.Exec(emailID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchRemoveEmailTags 批量移除邮箱标记
func (r *TagRepository) BatchRemoveEmailTags(emailIDs []int, tagID int) error {
	if len(emailIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM email_tags WHERE email_id = ? AND tag_id = ?`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, emailID := range emailIDs {
		_, err := stmt.Exec(emailID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetTagsWithEmailCount 获取用户的标签列表（包含邮箱数量）
func (r *TagRepository) GetTagsWithEmailCount(userID int) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.description, t.color, t.created_at, t.updated_at,
		       COALESCE(COUNT(et.email_id), 0) as email_count
		FROM tags t
		LEFT JOIN email_tags et ON t.id = et.tag_id
		LEFT JOIN emails e ON et.email_id = e.id AND e.user_id = ?
		GROUP BY t.id, t.name, t.description, t.color, t.created_at, t.updated_at
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
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
			&tag.EmailCount,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

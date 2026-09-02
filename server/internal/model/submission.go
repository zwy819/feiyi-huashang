package model

import (
	"database/sql"
	"time"
)

// Submission 用户图片投稿记录
type Submission struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"user_id"`
	Username          string    `json:"username,omitempty"`
	ImageFilename     string    `json:"image_filename"`
	ImageOriginalName string    `json:"image_original_name"`
	ImageFileSize     int64     `json:"image_file_size"`
	ImageMimeType     string    `json:"image_mime_type"`
	ImageFilePath     string    `json:"image_file_path"`
	Name              string    `json:"name"`
	Dynasty           string    `json:"dynasty"`
	Ethnic            string    `json:"ethnic"`
	Craft             string    `json:"craft"`
	Material          string    `json:"material"`
	Pattern           string    `json:"pattern"`
	CulturalMeaning   string    `json:"cultural_meaning"`
	Province          string    `json:"province"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
	RejectReason      string    `json:"reject_reason"`
	AdminNote         string    `json:"admin_note"`
	CreatedGarmentID  *int64    `json:"created_garment_id"`
	CreatedModelID    *int64    `json:"created_model_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SubmissionQueryParams 投稿列表查询参数
type SubmissionQueryParams struct {
	Page   int
	Limit  int
	Status string
	UserID int64
}

// SubmissionStore 投稿数据访问层
type SubmissionStore struct {
	db *sql.DB
}

// NewSubmissionStore 创建投稿 Store
func NewSubmissionStore(db *sql.DB) *SubmissionStore {
	return &SubmissionStore{db: db}
}

// Create 创建投稿记录
func (s *SubmissionStore) Create(sub *Submission) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO submissions (user_id, image_filename, image_original_name, image_file_size, image_mime_type, image_file_path, name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.UserID, sub.ImageFilename, sub.ImageOriginalName, sub.ImageFileSize,
		sub.ImageMimeType, sub.ImageFilePath, sub.Name, sub.Dynasty, sub.Ethnic,
		sub.Craft, sub.Material, sub.Pattern, sub.CulturalMeaning, sub.Province,
		sub.Description, sub.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetByID 根据 ID 获取投稿（含用户名）
func (s *SubmissionStore) GetByID(id int64) (*Submission, error) {
	sub := &Submission{}
	var rejectReason, adminNote sql.NullString
	var createdGarmentID, createdModelID sql.NullInt64
	var username sql.NullString

	err := s.db.QueryRow(
		`SELECT s.id, s.user_id, u.username, s.image_filename, s.image_original_name, s.image_file_size,
		        s.image_mime_type, s.image_file_path, s.name, s.dynasty, s.ethnic, s.craft, s.material,
		        s.pattern, s.cultural_meaning, s.province, s.description, s.status,
		        s.reject_reason, s.admin_note, s.created_garment_id, s.created_model_id,
		        s.created_at, s.updated_at
		 FROM submissions s
		 LEFT JOIN users u ON s.user_id = u.id
		 WHERE s.id = ?`, id,
	).Scan(&sub.ID, &sub.UserID, &username, &sub.ImageFilename, &sub.ImageOriginalName,
		&sub.ImageFileSize, &sub.ImageMimeType, &sub.ImageFilePath, &sub.Name, &sub.Dynasty,
		&sub.Ethnic, &sub.Craft, &sub.Material, &sub.Pattern, &sub.CulturalMeaning,
		&sub.Province, &sub.Description, &sub.Status, &rejectReason, &adminNote,
		&createdGarmentID, &createdModelID, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if username.Valid {
		sub.Username = username.String
	}
	if rejectReason.Valid {
		sub.RejectReason = rejectReason.String
	}
	if adminNote.Valid {
		sub.AdminNote = adminNote.String
	}
	if createdGarmentID.Valid {
		sub.CreatedGarmentID = &createdGarmentID.Int64
	}
	if createdModelID.Valid {
		sub.CreatedModelID = &createdModelID.Int64
	}
	return sub, nil
}

// List 管理员查询投稿列表（可按 status 筛选）
func (s *SubmissionStore) List(params SubmissionQueryParams) ([]Submission, int, error) {
	where := "1=1"
	args := []interface{}{}

	if params.Status != "" {
		where += " AND s.status = ?"
		args = append(args, params.Status)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM submissions s WHERE " + where
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query := `SELECT s.id, s.user_id, u.username, s.image_filename, s.image_original_name, s.image_file_size,
	                  s.image_mime_type, s.image_file_path, s.name, s.dynasty, s.ethnic, s.craft, s.material,
	                  s.pattern, s.cultural_meaning, s.province, s.description, s.status,
	                  s.reject_reason, s.admin_note, s.created_garment_id, s.created_model_id,
	                  s.created_at, s.updated_at
	           FROM submissions s
	           LEFT JOIN users u ON s.user_id = u.id
	           WHERE ` + where + ` ORDER BY s.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, params.Limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subs := make([]Submission, 0)
	for rows.Next() {
		var sub Submission
		var rejectReason, adminNote sql.NullString
		var createdGarmentID, createdModelID sql.NullInt64
		var username sql.NullString
		if err := rows.Scan(&sub.ID, &sub.UserID, &username, &sub.ImageFilename, &sub.ImageOriginalName,
			&sub.ImageFileSize, &sub.ImageMimeType, &sub.ImageFilePath, &sub.Name, &sub.Dynasty,
			&sub.Ethnic, &sub.Craft, &sub.Material, &sub.Pattern, &sub.CulturalMeaning,
			&sub.Province, &sub.Description, &sub.Status, &rejectReason, &adminNote,
			&createdGarmentID, &createdModelID, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if username.Valid {
			sub.Username = username.String
		}
		if rejectReason.Valid {
			sub.RejectReason = rejectReason.String
		}
		if adminNote.Valid {
			sub.AdminNote = adminNote.String
		}
		if createdGarmentID.Valid {
			sub.CreatedGarmentID = &createdGarmentID.Int64
		}
		if createdModelID.Valid {
			sub.CreatedModelID = &createdModelID.Int64
		}
		subs = append(subs, sub)
	}

	return subs, total, nil
}

// ListByUser 查询指定用户的投稿列表
func (s *SubmissionStore) ListByUser(userID int64, params SubmissionQueryParams) ([]Submission, int, error) {
	where := "s.user_id = ?"
	args := []interface{}{userID}

	if params.Status != "" {
		where += " AND s.status = ?"
		args = append(args, params.Status)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM submissions s WHERE " + where
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query := `SELECT s.id, s.user_id, u.username, s.image_filename, s.image_original_name, s.image_file_size,
	                  s.image_mime_type, s.image_file_path, s.name, s.dynasty, s.ethnic, s.craft, s.material,
	                  s.pattern, s.cultural_meaning, s.province, s.description, s.status,
	                  s.reject_reason, s.admin_note, s.created_garment_id, s.created_model_id,
	                  s.created_at, s.updated_at
	           FROM submissions s
	           LEFT JOIN users u ON s.user_id = u.id
	           WHERE ` + where + ` ORDER BY s.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, params.Limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subs := make([]Submission, 0)
	for rows.Next() {
		var sub Submission
		var rejectReason, adminNote sql.NullString
		var createdGarmentID, createdModelID sql.NullInt64
		var username sql.NullString
		if err := rows.Scan(&sub.ID, &sub.UserID, &username, &sub.ImageFilename, &sub.ImageOriginalName,
			&sub.ImageFileSize, &sub.ImageMimeType, &sub.ImageFilePath, &sub.Name, &sub.Dynasty,
			&sub.Ethnic, &sub.Craft, &sub.Material, &sub.Pattern, &sub.CulturalMeaning,
			&sub.Province, &sub.Description, &sub.Status, &rejectReason, &adminNote,
			&createdGarmentID, &createdModelID, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if username.Valid {
			sub.Username = username.String
		}
		if rejectReason.Valid {
			sub.RejectReason = rejectReason.String
		}
		if adminNote.Valid {
			sub.AdminNote = adminNote.String
		}
		if createdGarmentID.Valid {
			sub.CreatedGarmentID = &createdGarmentID.Int64
		}
		if createdModelID.Valid {
			sub.CreatedModelID = &createdModelID.Int64
		}
		subs = append(subs, sub)
	}

	return subs, total, nil
}

// ListPublicApproved 公开接口 — 获取已审核通过的投稿（approved 或 completed）
func (s *SubmissionStore) ListPublicApproved(page, limit int) ([]Submission, int, error) {
	where := "s.status IN ('approved', 'completed')"
	args := []interface{}{}

	var total int
	countQuery := "SELECT COUNT(*) FROM submissions s WHERE " + where
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `SELECT s.id, s.user_id, u.username, s.image_filename, s.image_original_name, s.image_file_size,
	                  s.image_mime_type, s.image_file_path, s.name, s.dynasty, s.ethnic, s.craft, s.material,
	                  s.pattern, s.cultural_meaning, s.province, s.description, s.status,
	                  s.reject_reason, s.admin_note, s.created_garment_id, s.created_model_id,
	                  s.created_at, s.updated_at
	           FROM submissions s
	           LEFT JOIN users u ON s.user_id = u.id
	           WHERE ` + where + ` ORDER BY s.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subs := make([]Submission, 0)
	for rows.Next() {
		var sub Submission
		var rejectReason, adminNote sql.NullString
		var createdGarmentID, createdModelID sql.NullInt64
		var username sql.NullString
		if err := rows.Scan(&sub.ID, &sub.UserID, &username, &sub.ImageFilename, &sub.ImageOriginalName,
			&sub.ImageFileSize, &sub.ImageMimeType, &sub.ImageFilePath, &sub.Name, &sub.Dynasty,
			&sub.Ethnic, &sub.Craft, &sub.Material, &sub.Pattern, &sub.CulturalMeaning,
			&sub.Province, &sub.Description, &sub.Status, &rejectReason, &adminNote,
			&createdGarmentID, &createdModelID, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if username.Valid {
			sub.Username = username.String
		}
		if rejectReason.Valid {
			sub.RejectReason = rejectReason.String
		}
		if adminNote.Valid {
			sub.AdminNote = adminNote.String
		}
		if createdGarmentID.Valid {
			sub.CreatedGarmentID = &createdGarmentID.Int64
		}
		if createdModelID.Valid {
			sub.CreatedModelID = &createdModelID.Int64
		}
		subs = append(subs, sub)
	}

	return subs, total, nil
}

// UpdateStatus 更新投稿状态（审核通过/驳回）
func (s *SubmissionStore) UpdateStatus(id int64, status, rejectReason, adminNote string) error {
	_, err := s.db.Exec(
		`UPDATE submissions SET status=?, reject_reason=?, admin_note=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		status, rejectReason, adminNote, id,
	)
	return err
}

// Complete 标记为已完成 — 关联 garment 和 model
func (s *SubmissionStore) Complete(id int64, garmentID, modelID *int64) error {
	_, err := s.db.Exec(
		`UPDATE submissions SET status='completed', created_garment_id=?, created_model_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		garmentID, modelID, id,
	)
	return err
}

// Delete 删除投稿记录
func (s *SubmissionStore) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM submissions WHERE id=?", id)
	return err
}

// Retract 用户撤回投稿 — 设为 retracted 状态，同时删除自动创建的服饰记录
func (s *SubmissionStore) Retract(id int64) error {
	_, err := s.db.Exec(
		`UPDATE submissions SET status='retracted', updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		id,
	)
	return err
}

// DeleteGarment 删除服饰记录
func (s *SubmissionStore) DeleteGarment(garmentID int64) error {
	_, err := s.db.Exec("DELETE FROM garments WHERE id=?", garmentID)
	return err
}

// UpdateFilePath 更新投稿图片文件路径信息
func (s *SubmissionStore) UpdateFilePath(id int64, filename, filePath string) error {
	_, err := s.db.Exec(
		"UPDATE submissions SET image_filename=?, image_file_path=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		filename, filePath, id,
	)
	return err
}

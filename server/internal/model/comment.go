package model

import (
	"database/sql"
	"time"
)

// Comment 服饰评论记录（需管理员审核后公开）
type Comment struct {
	ID           int64     `json:"id"`
	GarmentID    int64     `json:"garment_id"`
	GarmentName  string    `json:"garment_name,omitempty"`
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	RejectReason string    `json:"reject_reason"`
	LikeCount    int       `json:"like_count"`
	IsLiked      bool      `json:"is_liked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CommentQueryParams 评论列表查询参数
type CommentQueryParams struct {
	Page   int
	Limit  int
	Status string
	UserID int64
}

// CommentStore 评论数据访问层
type CommentStore struct {
	db *sql.DB
}

// NewCommentStore 创建评论 Store
func NewCommentStore(db *sql.DB) *CommentStore {
	return &CommentStore{db: db}
}

// commentSelectColumns 评论查询公共字段（like_count / is_liked 由子查询计算）
const commentSelectColumns = `c.id, c.garment_id, c.user_id, u.username, c.content, c.status, c.reject_reason,
	(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id) AS like_count,
	EXISTS(SELECT 1 FROM comment_likes cl2 WHERE cl2.comment_id = c.id AND cl2.user_id = ?) AS is_liked,
	c.created_at, c.updated_at`

func scanComment(scanner interface{ Scan(...interface{}) error }) (*Comment, error) {
	com := &Comment{}
	err := scanner.Scan(&com.ID, &com.GarmentID, &com.UserID, &com.Username,
		&com.Content, &com.Status, &com.RejectReason, &com.LikeCount, &com.IsLiked,
		&com.CreatedAt, &com.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return com, nil
}

// Create 创建评论（初始状态 pending，审核通过后公开）
func (s *CommentStore) Create(com *Comment) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO comments (garment_id, user_id, content, status) VALUES (?, ?, ?, 'pending')`,
		com.GarmentID, com.UserID, com.Content,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetByID 根据 ID 获取评论（含用户名）
// viewerID 用于计算 is_liked，传 0 表示未登录视角（is_liked 恒为 false）
func (s *CommentStore) GetByID(id, viewerID int64) (*Comment, error) {
	row := s.db.QueryRow(
		`SELECT `+commentSelectColumns+`
		 FROM comments c
		 LEFT JOIN users u ON c.user_id = u.id
		 WHERE c.id = ?`, viewerID, id,
	)
	return scanComment(row)
}

// CountByGarment 统计某服饰已通过审核的评论数
func (s *CommentStore) CountByGarment(garmentID int64) (int, error) {
	var total int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM comments WHERE garment_id = ? AND status = 'approved'`,
		garmentID,
	).Scan(&total)
	return total, err
}

// ListByGarment 公开接口 — 某服饰的评论列表
// 可见性：所有人可见 approved；登录用户额外可见自己的 pending / rejected
// viewerID 为 0 表示未登录
func (s *CommentStore) ListByGarment(garmentID, viewerID int64, page, limit int) ([]Comment, int, error) {
	where := `c.garment_id = ? AND (c.status = 'approved' OR (c.user_id = ? AND c.status IN ('pending', 'rejected')))`
	countArgs := []interface{}{garmentID, viewerID}

	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM comments c WHERE `+where, countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `SELECT ` + commentSelectColumns + `
	          FROM comments c
	          LEFT JOIN users u ON c.user_id = u.id
	          WHERE ` + where + `
	          ORDER BY c.created_at DESC, c.id DESC LIMIT ? OFFSET ?`
	// 参数顺序：is_liked 子查询的 viewerID 在 SELECT 中最先出现
	args := append([]interface{}{viewerID}, countArgs...)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	coms := make([]Comment, 0)
	for rows.Next() {
		com, err := scanComment(rows)
		if err != nil {
			return nil, 0, err
		}
		coms = append(coms, *com)
	}
	return coms, total, nil
}

// List 管理员查询评论列表（可按 status 筛选，含服饰名称）
func (s *CommentStore) List(params CommentQueryParams) ([]Comment, int, error) {
	where := "1=1"
	args := []interface{}{}

	if params.Status != "" {
		where += " AND c.status = ?"
		args = append(args, params.Status)
	}

	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM comments c WHERE `+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query := `SELECT c.id, c.garment_id, g.name, c.user_id, u.username, c.content, c.status, c.reject_reason,
	                  (SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id) AS like_count,
	                  0 AS is_liked,
	                  c.created_at, c.updated_at
	           FROM comments c
	           LEFT JOIN users u ON c.user_id = u.id
	           LEFT JOIN garments g ON c.garment_id = g.id
	           WHERE ` + where + ` ORDER BY c.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, params.Limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	coms := make([]Comment, 0)
	for rows.Next() {
		var com Comment
		if err := rows.Scan(&com.ID, &com.GarmentID, &com.GarmentName, &com.UserID, &com.Username,
			&com.Content, &com.Status, &com.RejectReason, &com.LikeCount, &com.IsLiked,
			&com.CreatedAt, &com.UpdatedAt); err != nil {
			return nil, 0, err
		}
		coms = append(coms, com)
	}
	return coms, total, nil
}

// UpdateContent 用户编辑评论内容 — 重置为 pending 重新审核
func (s *CommentStore) UpdateContent(id int64, content string) error {
	_, err := s.db.Exec(
		`UPDATE comments SET content=?, status='pending', reject_reason='', updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		content, id,
	)
	return err
}

// UpdateStatus 管理员审核评论（通过/驳回）
func (s *CommentStore) UpdateStatus(id int64, status, rejectReason string) error {
	_, err := s.db.Exec(
		`UPDATE comments SET status=?, reject_reason=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		status, rejectReason, id,
	)
	return err
}

// Delete 删除评论（点赞记录由外键级联删除）
func (s *CommentStore) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM comments WHERE id=?", id)
	return err
}

// HasLiked 检查用户是否已点赞某评论
func (s *CommentStore) HasLiked(commentID, userID int64) (bool, error) {
	var exists int
	err := s.db.QueryRow(
		`SELECT 1 FROM comment_likes WHERE comment_id=? AND user_id=?`,
		commentID, userID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddLike 点赞评论
func (s *CommentStore) AddLike(commentID, userID int64) error {
	_, err := s.db.Exec(
		`INSERT INTO comment_likes (comment_id, user_id) VALUES (?, ?)`,
		commentID, userID,
	)
	return err
}

// RemoveLike 取消点赞，返回是否实际删除了记录
func (s *CommentStore) RemoveLike(commentID, userID int64) (bool, error) {
	result, err := s.db.Exec(
		`DELETE FROM comment_likes WHERE comment_id=? AND user_id=?`,
		commentID, userID,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

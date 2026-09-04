package model

import (
	"database/sql"
	"time"
)

// QuizRecord 问答小游戏答题记录
type QuizRecord struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Username        string    `json:"username,omitempty"`
	Score           int       `json:"score"`
	Total           int       `json:"total"`
	CorrectCount    int       `json:"correct_count"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
}

// QuizSource 出题素材（服饰关键字段快照）
type QuizSource struct {
	ID              int64
	Name            string
	Dynasty         string
	Ethnic          string
	Province        string
	Craft           string
	Pattern         string
	CulturalMeaning string
	Description     string
}

// QuizStore 问答数据访问层
type QuizStore struct {
	db *sql.DB
}

// NewQuizStore 创建问答 Store
func NewQuizStore(db *sql.DB) *QuizStore {
	return &QuizStore{db: db}
}

// ListSource 读取全部服饰的出题字段
func (s *QuizStore) ListSource() ([]QuizSource, error) {
	rows, err := s.db.Query(
		`SELECT id, name, dynasty, ethnic, province, craft, pattern, cultural_meaning, description
		 FROM garments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]QuizSource, 0)
	for rows.Next() {
		var q QuizSource
		if err := rows.Scan(&q.ID, &q.Name, &q.Dynasty, &q.Ethnic, &q.Province,
			&q.Craft, &q.Pattern, &q.CulturalMeaning, &q.Description); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, rows.Err()
}

// FieldValue 查询服饰某字段的实际值（判分用）
// field 仅接受白名单列名，其余返回空串
func (s *QuizStore) FieldValue(garmentID int64, field string) (string, error) {
	allowed := map[string]bool{
		"name": true, "dynasty": true, "ethnic": true,
		"province": true, "craft": true, "pattern": true,
	}
	if !allowed[field] {
		return "", nil
	}
	var value string
	err := s.db.QueryRow(`SELECT `+field+` FROM garments WHERE id = ?`, garmentID).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// InsertRecord 记录一次答题成绩
func (s *QuizStore) InsertRecord(userID int64, score, total, correctCount, durationSeconds int) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO quiz_records (user_id, score, total, correct_count, duration_seconds)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, score, total, correctCount, durationSeconds,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetBestScore 用户历史最高分
func (s *QuizStore) GetBestScore(userID int64) (int, error) {
	var best int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(score), 0) FROM quiz_records WHERE user_id = ?`, userID,
	).Scan(&best)
	return best, err
}

// ListLeaderboard 排行榜：每用户最高分 Top N
func (s *QuizStore) ListLeaderboard(limit int) ([]QuizRecord, error) {
	rows, err := s.db.Query(
		`SELECT r.user_id, u.username, MAX(r.score) AS score,
		        MAX(r.total) AS total, MAX(r.correct_count) AS correct_count,
		        MIN(r.duration_seconds) AS duration_seconds, MAX(r.created_at) AS created_at
		 FROM quiz_records r
		 LEFT JOIN users u ON r.user_id = u.id
		 GROUP BY r.user_id
		 ORDER BY score DESC, duration_seconds ASC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]QuizRecord, 0)
	for rows.Next() {
		var r QuizRecord
		// MAX() 聚合会丢失 DATETIME 声明类型，驱动不再自动转 time.Time，需手动解析
		var createdAt interface{}
		if err := rows.Scan(&r.UserID, &r.Username, &r.Score, &r.Total,
			&r.CorrectCount, &r.DurationSeconds, &createdAt); err != nil {
			return nil, err
		}
		r.CreatedAt = parseSQLiteTime(createdAt)
		list = append(list, r)
	}
	return list, rows.Err()
}

// parseSQLiteTime 兼容驱动转好的 time.Time 与原始 TEXT 两种形态
func parseSQLiteTime(value interface{}) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case []byte:
		if t, err := time.Parse("2006-01-02 15:04:05", string(v)); err == nil {
			return t
		}
	case string:
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ListByUser 我的最近战绩
func (s *QuizStore) ListByUser(userID int64, limit int) ([]QuizRecord, error) {
	rows, err := s.db.Query(
		`SELECT user_id, score, total, correct_count, duration_seconds, created_at
		 FROM quiz_records WHERE user_id = ?
		 ORDER BY created_at DESC, id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]QuizRecord, 0)
	for rows.Next() {
		var r QuizRecord
		if err := rows.Scan(&r.UserID, &r.Score, &r.Total, &r.CorrectCount,
			&r.DurationSeconds, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

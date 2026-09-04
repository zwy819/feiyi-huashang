package model

import (
	"database/sql"
	"time"
)

// Annotation 服饰 3D 模型热点标注（纹样/工艺讲解点）
// 坐标为前端展示场景的世界坐标：模型经包围盒平移居中后不旋转，坐标稳定可复现
type Annotation struct {
	ID           int64     `json:"id"`
	GarmentID    int64     `json:"garment_id"`
	GarmentName  string    `json:"garment_name,omitempty"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	PositionX    float64   `json:"position_x"`
	PositionY    float64   `json:"position_y"`
	PositionZ    float64   `json:"position_z"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AnnotationStore 热点标注数据访问层
type AnnotationStore struct {
	db *sql.DB
}

// NewAnnotationStore 创建热点标注 Store
func NewAnnotationStore(db *sql.DB) *AnnotationStore {
	return &AnnotationStore{db: db}
}

func scanAnnotation(scanner interface{ Scan(...interface{}) error }) (*Annotation, error) {
	a := &Annotation{}
	err := scanner.Scan(&a.ID, &a.GarmentID, &a.Title, &a.Content,
		&a.PositionX, &a.PositionY, &a.PositionZ, &a.SortOrder,
		&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ListByGarment 某服饰的标注列表（按 sort_order 升序，公开接口）
func (s *AnnotationStore) ListByGarment(garmentID int64) ([]Annotation, error) {
	rows, err := s.db.Query(
		`SELECT id, garment_id, title, content, position_x, position_y, position_z, sort_order, created_at, updated_at
		 FROM annotations WHERE garment_id = ? ORDER BY sort_order ASC, id ASC`, garmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Annotation, 0)
	for rows.Next() {
		a, err := scanAnnotation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *a)
	}
	return list, rows.Err()
}

// List 管理员查询标注列表（可按 garment_id 过滤，含服饰名称）
func (s *AnnotationStore) List(garmentID int64) ([]Annotation, error) {
	where := "1=1"
	args := []interface{}{}
	if garmentID > 0 {
		where = "a.garment_id = ?"
		args = append(args, garmentID)
	}

	rows, err := s.db.Query(
		`SELECT a.id, a.garment_id, g.name, a.title, a.content,
		        a.position_x, a.position_y, a.position_z, a.sort_order, a.created_at, a.updated_at
		 FROM annotations a
		 LEFT JOIN garments g ON a.garment_id = g.id
		 WHERE `+where+` ORDER BY a.garment_id ASC, a.sort_order ASC, a.id ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Annotation, 0)
	for rows.Next() {
		a := &Annotation{}
		if err := rows.Scan(&a.ID, &a.GarmentID, &a.GarmentName, &a.Title, &a.Content,
			&a.PositionX, &a.PositionY, &a.PositionZ, &a.SortOrder,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, *a)
	}
	return list, rows.Err()
}

// GetByID 根据 ID 获取标注
func (s *AnnotationStore) GetByID(id int64) (*Annotation, error) {
	row := s.db.QueryRow(
		`SELECT id, garment_id, title, content, position_x, position_y, position_z, sort_order, created_at, updated_at
		 FROM annotations WHERE id = ?`, id)
	return scanAnnotation(row)
}

// Create 创建标注
func (s *AnnotationStore) Create(a *Annotation) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO annotations (garment_id, title, content, position_x, position_y, position_z, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.GarmentID, a.Title, a.Content, a.PositionX, a.PositionY, a.PositionZ, a.SortOrder,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update 更新标注
func (s *AnnotationStore) Update(a *Annotation) error {
	_, err := s.db.Exec(
		`UPDATE annotations SET title=?, content=?, position_x=?, position_y=?, position_z=?, sort_order=?,
		        updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		a.Title, a.Content, a.PositionX, a.PositionY, a.PositionZ, a.SortOrder, a.ID,
	)
	return err
}

// Delete 删除标注
func (s *AnnotationStore) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM annotations WHERE id=?", id)
	return err
}

// CountAll 统计标注总数（seed 判断用）
func (s *AnnotationStore) CountAll() (int, error) {
	var total int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM annotations`).Scan(&total)
	return total, err
}

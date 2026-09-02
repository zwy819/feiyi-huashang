package model

import (
	"database/sql"
	"time"
)

type Model3D struct {
	ID           int64     `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	FilePath     string    `json:"file_path"`
	GarmentID    *int64    `json:"garment_id"`
	GarmentName  string    `json:"garment_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ModelQueryParams struct {
	Page      int
	Limit     int
	GarmentID *int64
	Search    string
}

type ModelStore struct {
	db *sql.DB
}

func NewModelStore(db *sql.DB) *ModelStore {
	return &ModelStore{db: db}
}

func (s *ModelStore) Create(m *Model3D) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO models (filename, original_name, file_size, mime_type, file_path, garment_id) VALUES (?, ?, ?, ?, ?, ?)`,
		m.Filename, m.OriginalName, m.FileSize, m.MimeType, m.FilePath, m.GarmentID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *ModelStore) GetByID(id int64) (*Model3D, error) {
	m := &Model3D{}
	var garmentName sql.NullString
	err := s.db.QueryRow(
		`SELECT m.id, m.filename, m.original_name, m.file_size, m.mime_type, m.file_path, m.garment_id, g.name, m.created_at, m.updated_at
		 FROM models m
		 LEFT JOIN garments g ON m.garment_id = g.id
		 WHERE m.id = ?`, id,
	).Scan(&m.ID, &m.Filename, &m.OriginalName, &m.FileSize, &m.MimeType, &m.FilePath, &m.GarmentID, &garmentName, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if garmentName.Valid {
		m.GarmentName = garmentName.String
	}
	return m, nil
}

func (s *ModelStore) List(params ModelQueryParams) ([]Model3D, int, error) {
	where := "1=1"
	args := []interface{}{}

	if params.GarmentID != nil {
		where += " AND m.garment_id = ?"
		args = append(args, *params.GarmentID)
	}
	if params.Search != "" {
		where += " AND (m.original_name LIKE ? OR m.filename LIKE ?)"
		s := "%" + params.Search + "%"
		args = append(args, s, s)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM models m WHERE " + where
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query := `SELECT m.id, m.filename, m.original_name, m.file_size, m.mime_type, m.file_path, m.garment_id, g.name, m.created_at, m.updated_at
		FROM models m
		LEFT JOIN garments g ON m.garment_id = g.id
		WHERE ` + where + ` ORDER BY m.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, params.Limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var models []Model3D
	for rows.Next() {
		var m Model3D
		var garmentName sql.NullString
		if err := rows.Scan(&m.ID, &m.Filename, &m.OriginalName, &m.FileSize, &m.MimeType, &m.FilePath, &m.GarmentID, &garmentName, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if garmentName.Valid {
			m.GarmentName = garmentName.String
		}
		models = append(models, m)
	}
	if models == nil {
		models = []Model3D{}
	}
	return models, total, nil
}

func (s *ModelStore) Update(id int64, m *Model3D) error {
	_, err := s.db.Exec(
		"UPDATE models SET original_name=?, garment_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		m.OriginalName, m.GarmentID, id,
	)
	return err
}

func (s *ModelStore) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM models WHERE id=?", id)
	return err
}

func (s *ModelStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM models").Scan(&count)
	return count, err
}

package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Garment struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Dynasty         string    `json:"dynasty"`
	Ethnic          string    `json:"ethnic"`
	Craft           string    `json:"craft"`
	Material        string    `json:"material"`
	Pattern         string    `json:"pattern"`
	CulturalMeaning string    `json:"cultural_meaning"`
	Province        string    `json:"province"`
	Description     string    `json:"description"`
	ImageURL        string    `json:"image_url"`
	ModelURL        string    `json:"model_url"`
	ModelID         *int64    `json:"model_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GarmentQueryParams struct {
	Page    int
	Limit   int
	Search  string
	Dynasty string
	Ethnic  string
	Craft   string
}

type GarmentStore struct {
	db *sql.DB
}

func NewGarmentStore(db *sql.DB) *GarmentStore {
	return &GarmentStore{db: db}
}

func (s *GarmentStore) List(params GarmentQueryParams) ([]Garment, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if params.Search != "" {
		where = append(where, "(name LIKE ? OR description LIKE ? OR craft LIKE ? OR ethnic LIKE ?)")
		s := "%" + params.Search + "%"
		args = append(args, s, s, s, s)
	}
	if params.Dynasty != "" {
		where = append(where, "dynasty = ?")
		args = append(args, params.Dynasty)
	}
	if params.Ethnic != "" {
		where = append(where, "ethnic = ?")
		args = append(args, params.Ethnic)
	}
	if params.Craft != "" {
		where = append(where, "craft = ?")
		args = append(args, params.Craft)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM garments WHERE %s", strings.Join(where, " AND "))
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Limit
	query := fmt.Sprintf(
		"SELECT id, name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, image_url, model_url, model_id, created_at, updated_at FROM garments WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?",
		strings.Join(where, " AND "),
	)
	queryArgs := append(args, params.Limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var garments []Garment
	for rows.Next() {
		var g Garment
		if err := rows.Scan(&g.ID, &g.Name, &g.Dynasty, &g.Ethnic, &g.Craft, &g.Material, &g.Pattern, &g.CulturalMeaning, &g.Province, &g.Description, &g.ImageURL, &g.ModelURL, &g.ModelID, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		garments = append(garments, g)
	}

	if garments == nil {
		garments = []Garment{}
	}

	return garments, total, nil
}

func (s *GarmentStore) GetByID(id int64) (*Garment, error) {
	g := &Garment{}
	err := s.db.QueryRow(
		`SELECT id, name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, image_url, model_url, model_id, created_at, updated_at FROM garments WHERE id = ?`,
		id,
	).Scan(&g.ID, &g.Name, &g.Dynasty, &g.Ethnic, &g.Craft, &g.Material, &g.Pattern, &g.CulturalMeaning, &g.Province, &g.Description, &g.ImageURL, &g.ModelURL, &g.ModelID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GarmentStore) Create(g *Garment) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO garments (name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, image_url, model_url, model_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.Name, g.Dynasty, g.Ethnic, g.Craft, g.Material, g.Pattern, g.CulturalMeaning, g.Province, g.Description, g.ImageURL, g.ModelURL, g.ModelID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *GarmentStore) Update(id int64, g *Garment) error {
	_, err := s.db.Exec(
		`UPDATE garments SET name=?, dynasty=?, ethnic=?, craft=?, material=?, pattern=?, cultural_meaning=?, province=?, description=?, image_url=?, model_url=?, model_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		g.Name, g.Dynasty, g.Ethnic, g.Craft, g.Material, g.Pattern, g.CulturalMeaning, g.Province, g.Description, g.ImageURL, g.ModelURL, g.ModelID, id,
	)
	return err
}

func (s *GarmentStore) UpdateModelID(id int64, modelID *int64) error {
	_, err := s.db.Exec("UPDATE garments SET model_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", modelID, id)
	return err
}

func (s *GarmentStore) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM garments WHERE id=?", id)
	return err
}

func (s *GarmentStore) GetByEthnicAndCraft(ethnic, craft string, excludeID int64) ([]Garment, error) {
	rows, err := s.db.Query(
		`SELECT id, name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, image_url, model_url, model_id, created_at, updated_at FROM garments WHERE id != ? AND (ethnic = ? OR craft = ?) ORDER BY CASE WHEN ethnic = ? AND craft = ? THEN 2 WHEN ethnic = ? OR craft = ? THEN 1 ELSE 0 END DESC, id DESC LIMIT 4`,
		excludeID, ethnic, craft, ethnic, craft, ethnic, craft,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var garments []Garment
	for rows.Next() {
		var g Garment
		if err := rows.Scan(&g.ID, &g.Name, &g.Dynasty, &g.Ethnic, &g.Craft, &g.Material, &g.Pattern, &g.CulturalMeaning, &g.Province, &g.Description, &g.ImageURL, &g.ModelURL, &g.ModelID, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		garments = append(garments, g)
	}
	if garments == nil {
		garments = []Garment{}
	}
	return garments, nil
}

func (s *GarmentStore) GetRandom() (*Garment, error) {
	g := &Garment{}
	err := s.db.QueryRow(
		`SELECT id, name, dynasty, ethnic, craft, material, pattern, cultural_meaning, province, description, image_url, model_url, model_id, created_at, updated_at FROM garments ORDER BY RANDOM() LIMIT 1`,
	).Scan(&g.ID, &g.Name, &g.Dynasty, &g.Ethnic, &g.Craft, &g.Material, &g.Pattern, &g.CulturalMeaning, &g.Province, &g.Description, &g.ImageURL, &g.ModelURL, &g.ModelID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GarmentStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM garments").Scan(&count)
	return count, err
}

func (s *GarmentStore) CountDistinctEthnics() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(DISTINCT ethnic) FROM garments WHERE ethnic != ''").Scan(&count)
	return count, err
}

func (s *GarmentStore) CountDistinctCrafts() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(DISTINCT craft) FROM garments WHERE craft != ''").Scan(&count)
	return count, err
}

func (s *GarmentStore) CountDistinctProvinces() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(DISTINCT province) FROM garments WHERE province != ''").Scan(&count)
	return count, err
}

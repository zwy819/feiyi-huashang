package model

import (
	"database/sql"
	"time"
)

type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	GarmentID int64     `json:"garment_id"`
	CreatedAt time.Time `json:"created_at"`
	// Joined fields
	GarmentName     string `json:"garment_name,omitempty"`
	GarmentDynasty  string `json:"garment_dynasty,omitempty"`
	GarmentEthnic   string `json:"garment_ethnic,omitempty"`
	GarmentCraft    string `json:"garment_craft,omitempty"`
	GarmentProvince string `json:"garment_province,omitempty"`
	GarmentImageURL string `json:"garment_image_url,omitempty"`
}

type FavoriteStore struct {
	db *sql.DB
}

func NewFavoriteStore(db *sql.DB) *FavoriteStore {
	return &FavoriteStore{db: db}
}

func (s *FavoriteStore) Add(userID, garmentID int64) error {
	_, err := s.db.Exec(
		"INSERT INTO favorites (user_id, garment_id) VALUES (?, ?)",
		userID, garmentID,
	)
	return err
}

func (s *FavoriteStore) Remove(userID, garmentID int64) error {
	_, err := s.db.Exec(
		"DELETE FROM favorites WHERE user_id = ? AND garment_id = ?",
		userID, garmentID,
	)
	return err
}

func (s *FavoriteStore) ListByUser(userID int64) ([]Favorite, error) {
	rows, err := s.db.Query(
		`SELECT f.id, f.user_id, f.garment_id, f.created_at,
		        g.name, g.dynasty, g.ethnic, g.craft, g.province, g.image_url
		 FROM favorites f
		 JOIN garments g ON f.garment_id = g.id
		 WHERE f.user_id = ?
		 ORDER BY f.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favs []Favorite
	for rows.Next() {
		var f Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.GarmentID, &f.CreatedAt,
			&f.GarmentName, &f.GarmentDynasty, &f.GarmentEthnic, &f.GarmentCraft, &f.GarmentProvince, &f.GarmentImageURL); err != nil {
			return nil, err
		}
		favs = append(favs, f)
	}
	if favs == nil {
		favs = []Favorite{}
	}
	return favs, nil
}

func (s *FavoriteStore) IsFavorited(userID, garmentID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM favorites WHERE user_id = ? AND garment_id = ?",
		userID, garmentID,
	).Scan(&count)
	return count > 0, err
}

func (s *FavoriteStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM favorites").Scan(&count)
	return count, err
}

func (s *FavoriteStore) CountByUser(userID int64) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM favorites WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

package model

import (
	"database/sql"
	"time"
)

type DailyPick struct {
	ID        int64     `json:"id"`
	GarmentID int64     `json:"garment_id"`
	PickDate  string    `json:"pick_date"`
	CreatedAt time.Time `json:"created_at"`
}

type DailyPickStore struct {
	db *sql.DB
}

func NewDailyPickStore(db *sql.DB) *DailyPickStore {
	return &DailyPickStore{db: db}
}

func (s *DailyPickStore) GetByDate(date string) (*DailyPick, error) {
	dp := &DailyPick{}
	err := s.db.QueryRow(
		"SELECT id, garment_id, pick_date, created_at FROM daily_picks WHERE pick_date = ?",
		date,
	).Scan(&dp.ID, &dp.GarmentID, &dp.PickDate, &dp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dp, nil
}

func (s *DailyPickStore) Create(garmentID int64, date string) error {
	_, err := s.db.Exec(
		"INSERT INTO daily_picks (garment_id, pick_date) VALUES (?, ?)",
		garmentID, date,
	)
	return err
}

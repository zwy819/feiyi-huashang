package service

import (
	"nonlegacy-garment-3d/server/internal/model"
	"time"
)

type DailyService struct {
	dailyPickStore *model.DailyPickStore
	garmentStore   *model.GarmentStore
}

func NewDailyService(dailyPickStore *model.DailyPickStore, garmentStore *model.GarmentStore) *DailyService {
	return &DailyService{dailyPickStore: dailyPickStore, garmentStore: garmentStore}
}

func (s *DailyService) GetDaily() (*model.Garment, error) {
	today := time.Now().Format("2006-01-02")

	pick, err := s.dailyPickStore.GetByDate(today)
	if err == nil {
		return s.garmentStore.GetByID(pick.GarmentID)
	}

	// 随机选一件
	garment, err := s.garmentStore.GetRandom()
	if err != nil {
		return nil, err
	}

	// 记录今日推荐
	_ = s.dailyPickStore.Create(garment.ID, today)

	return garment, nil
}

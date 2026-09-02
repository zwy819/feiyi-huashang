package service

import (
	"nonlegacy-garment-3d/server/internal/model"
)

type GarmentService struct {
	garmentStore *model.GarmentStore
}

func NewGarmentService(garmentStore *model.GarmentStore) *GarmentService {
	return &GarmentService{garmentStore: garmentStore}
}

type RecommendItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Similarity int    `json:"similarity"`
}

func (s *GarmentService) GetRecommendChain(id int64) ([]RecommendItem, error) {
	garment, err := s.garmentStore.GetByID(id)
	if err != nil {
		return nil, err
	}

	similar, err := s.garmentStore.GetByEthnicAndCraft(garment.Ethnic, garment.Craft, id)
	if err != nil {
		return nil, err
	}

	var items []RecommendItem
	for _, g := range similar {
		similarity := 0
		if g.Ethnic == garment.Ethnic {
			similarity++
		}
		if g.Craft == garment.Craft {
			similarity++
		}
		items = append(items, RecommendItem{
			ID:         g.ID,
			Name:       g.Name,
			Similarity: similarity,
		})
	}

	if items == nil {
		items = []RecommendItem{}
	}

	return items, nil
}

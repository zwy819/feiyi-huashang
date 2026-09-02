package service

import (
	"database/sql"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
)

type StatisticsService struct {
	db            *sql.DB
	garmentStore  *model.GarmentStore
	favoriteStore *model.FavoriteStore
}

func NewStatisticsService(db *sql.DB, garmentStore *model.GarmentStore, favoriteStore *model.FavoriteStore) *StatisticsService {
	return &StatisticsService{db: db, garmentStore: garmentStore, favoriteStore: favoriteStore}
}

type GeoItem struct {
	Province string `json:"province"`
	Count    int    `json:"count"`
}

type CraftFlowItem struct {
	Craft string `json:"craft"`
	Count int    `json:"count"`
}

type OverviewStats struct {
	TotalGarments  int `json:"totalGarments"`
	TotalEthnics   int `json:"totalEthnics"`
	TotalCrafts    int `json:"totalCrafts"`
	TotalProvinces int `json:"totalProvinces"`
}

// SankeyNode 桑基图节点
type SankeyNode struct {
	Name string `json:"name"`
}

// SankeyLink 桑基图连接
type SankeyLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

// SankeyData 桑基图完整数据
type SankeyData struct {
	Nodes []SankeyNode `json:"nodes"`
	Links []SankeyLink `json:"links"`
}

// RawGarmentRow 原始服饰记录（用于构建桑基图）
type RawGarmentRow struct {
	Dynasty string
	Ethnic  string
	Craft   string
}

func (s *StatisticsService) GetGeoDistribution() ([]GeoItem, error) {
	rows, err := s.db.Query(
		"SELECT province, COUNT(*) as count FROM garments WHERE province != '' GROUP BY province ORDER BY count DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GeoItem
	for rows.Next() {
		var item GeoItem
		if err := rows.Scan(&item.Province, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []GeoItem{}
	}
	return items, nil
}

func (s *StatisticsService) GetCraftFlow() ([]CraftFlowItem, error) {
	rows, err := s.db.Query(
		"SELECT craft, COUNT(*) as count FROM garments WHERE craft != '' GROUP BY craft ORDER BY count DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CraftFlowItem
	for rows.Next() {
		var item CraftFlowItem
		if err := rows.Scan(&item.Craft, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []CraftFlowItem{}
	}
	return items, nil
}

func (s *StatisticsService) GetOverview() (*OverviewStats, error) {
	stats := &OverviewStats{}

	var err error
	stats.TotalGarments, err = s.garmentStore.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalEthnics, err = s.garmentStore.CountDistinctEthnics()
	if err != nil {
		return nil, err
	}
	stats.TotalCrafts, err = s.garmentStore.CountDistinctCrafts()
	if err != nil {
		return nil, err
	}
	stats.TotalProvinces, err = s.garmentStore.CountDistinctProvinces()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetSankeyData 构建 朝代→民族→工艺 三层桑基图数据
func (s *StatisticsService) GetSankeyData() (*SankeyData, error) {
	rows, err := s.db.Query(
		"SELECT dynasty, ethnic, craft FROM garments WHERE dynasty != '' AND ethnic != '' AND craft != ''",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RawGarmentRow
	for rows.Next() {
		var r RawGarmentRow
		if err := rows.Scan(&r.Dynasty, &r.Ethnic, &r.Craft); err != nil {
			return nil, err
		}
		// 拆分复合工艺（如 "蜡染+刺绣" → ["蜡染", "刺绣"]）
		crafts := splitCrafts(r.Craft)
		for _, craft := range crafts {
			records = append(records, RawGarmentRow{
				Dynasty: r.Dynasty,
				Ethnic:  r.Ethnic,
				Craft:   craft,
			})
		}
	}

	// 统计朝代→民族
	type linkKey struct {
		Source string
		Target string
	}
	dynastyToEthnic := make(map[linkKey]int)
	for _, r := range records {
		dynastyToEthnic[linkKey{r.Dynasty, r.Ethnic}]++
	}

	// 统计民族→工艺
	ethnicToCraft := make(map[linkKey]int)
	for _, r := range records {
		ethnicToCraft[linkKey{r.Ethnic, r.Craft}]++
	}

	// 收集所有不重复节点
	nodeSet := make(map[string]struct{})
	for _, r := range records {
		nodeSet[r.Dynasty] = struct{}{}
		nodeSet[r.Ethnic] = struct{}{}
		nodeSet[r.Craft] = struct{}{}
	}

	var nodes []SankeyNode
	for name := range nodeSet {
		nodes = append(nodes, SankeyNode{Name: name})
	}

	// 构建 links
	var links []SankeyLink
	for k, v := range dynastyToEthnic {
		links = append(links, SankeyLink{Source: k.Source, Target: k.Target, Value: v})
	}
	for k, v := range ethnicToCraft {
		links = append(links, SankeyLink{Source: k.Source, Target: k.Target, Value: v})
	}

	if nodes == nil {
		nodes = []SankeyNode{}
	}
	if links == nil {
		links = []SankeyLink{}
	}

	return &SankeyData{Nodes: nodes, Links: links}, nil
}

// splitCrafts 拆分复合工艺字符串
func splitCrafts(craft string) []string {
	parts := strings.Split(craft, "+")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{craft}
	}
	return result
}

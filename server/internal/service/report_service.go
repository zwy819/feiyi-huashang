package service

import (
	"bytes"
	"fmt"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/jung-kurt/gofpdf"
)

type ReportService struct {
	garmentStore *model.GarmentStore
}

func NewReportService(garmentStore *model.GarmentStore) *ReportService {
	return &ReportService{garmentStore: garmentStore}
}

func (s *ReportService) GeneratePDF(ids []int64) ([]byte, error) {
	var garments []*model.Garment
	for _, id := range ids {
		g, err := s.garmentStore.GetByID(id)
		if err != nil {
			continue
		}
		garments = append(garments, g)
	}

	if len(garments) == 0 {
		return nil, fmt.Errorf("未找到指定的服饰")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	// 封面
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 24)
	pdf.CellFormat(0, 15, "非遗华裳", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 14)
	pdf.CellFormat(0, 10, "中国非物质文化遗产服饰数字博物馆", "", 1, "C", false, 0, "")
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(0, 8, fmt.Sprintf("共 %d 件服饰", len(garments)), "", 1, "C", false, 0, "")

	for _, g := range garments {
		pdf.AddPage()

		// 服饰名称
		pdf.SetFont("Helvetica", "B", 18)
		pdf.CellFormat(0, 12, g.Name, "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 11)
		pdf.Ln(4)

		info := [][2]string{
			{"朝代", g.Dynasty + "代"},
			{"民族", g.Ethnic},
			{"工艺", g.Craft},
			{"材质", g.Material},
			{"纹样", g.Pattern},
			{"省份", g.Province},
		}

		for _, item := range info {
			if item[1] == "" || item[1] == "代" {
				continue
			}
			pdf.SetFont("Helvetica", "B", 11)
			pdf.CellFormat(30, 8, item[0]+":", "", 0, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 11)
			pdf.CellFormat(0, 8, item[1], "", 1, "L", false, 0, "")
		}

		pdf.Ln(4)

		// 文化寓意
		if g.CulturalMeaning != "" {
			pdf.SetFont("Helvetica", "B", 11)
			pdf.CellFormat(0, 8, "文化寓意:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 10)
			pdf.MultiCell(0, 6, g.CulturalMeaning, "", "L", false)
			pdf.Ln(2)
		}

		// 描述
		if g.Description != "" {
			pdf.SetFont("Helvetica", "B", 11)
			pdf.CellFormat(0, 8, "描述:", "", 1, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 10)
			pdf.MultiCell(0, 6, g.Description, "", "L", false)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// parseIDs parses comma-separated ID string
func ParseIDs(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var ids []int64
	for _, p := range parts {
		var id int64
		if _, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &id); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

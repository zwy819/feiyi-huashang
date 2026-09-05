package service

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/jung-kurt/gofpdf"
)

//go:embed fonts/simhei.ttf
var simheiTTF []byte

// reportFontFamily 支持中文的嵌入字体；gofpdf 内置 Helvetica 仅覆盖拉丁字符，
// 中文会渲染为乱码，因此报告统一使用嵌入的黑体（B 样式复用同字节模拟粗体）。
const reportFontFamily = "simhei"

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
	pdf.AddUTF8FontFromBytes(reportFontFamily, "", simheiTTF)
	pdf.AddUTF8FontFromBytes(reportFontFamily, "B", simheiTTF)

	// 封面
	pdf.AddPage()
	pdf.SetFont(reportFontFamily, "B", 24)
	pdf.CellFormat(0, 15, "非遗华裳", "", 1, "C", false, 0, "")
	pdf.SetFont(reportFontFamily, "", 14)
	pdf.CellFormat(0, 10, "中国非物质文化遗产服饰数字博物馆", "", 1, "C", false, 0, "")
	pdf.Ln(10)
	pdf.SetFont(reportFontFamily, "", 12)
	pdf.CellFormat(0, 8, fmt.Sprintf("共 %d 件服饰", len(garments)), "", 1, "C", false, 0, "")

	for _, g := range garments {
		pdf.AddPage()

		// 服饰名称
		pdf.SetFont(reportFontFamily, "B", 18)
		pdf.CellFormat(0, 12, g.Name, "", 1, "L", false, 0, "")

		pdf.SetFont(reportFontFamily, "", 11)
		pdf.Ln(4)

		// "清"→"清代"，已是"现代/古代"这类以"代"结尾的则原样保留
		dynasty := g.Dynasty
		if dynasty != "" && !strings.HasSuffix(dynasty, "代") {
			dynasty += "代"
		}

		info := [][2]string{
			{"朝代", dynasty},
			{"民族", g.Ethnic},
			{"工艺", g.Craft},
			{"材质", g.Material},
			{"纹样", g.Pattern},
			{"省份", g.Province},
		}

		for _, item := range info {
			if item[1] == "" {
				continue
			}
			pdf.SetFont(reportFontFamily, "B", 11)
			pdf.CellFormat(30, 8, item[0]+":", "", 0, "L", false, 0, "")
			pdf.SetFont(reportFontFamily, "", 11)
			pdf.CellFormat(0, 8, item[1], "", 1, "L", false, 0, "")
		}

		pdf.Ln(4)

		// 文化寓意
		if g.CulturalMeaning != "" {
			pdf.SetFont(reportFontFamily, "B", 11)
			pdf.CellFormat(0, 8, "文化寓意:", "", 1, "L", false, 0, "")
			pdf.SetFont(reportFontFamily, "", 10)
			pdf.MultiCell(0, 6, g.CulturalMeaning, "", "L", false)
			pdf.Ln(2)
		}

		// 描述
		if g.Description != "" {
			pdf.SetFont(reportFontFamily, "B", 11)
			pdf.CellFormat(0, 8, "描述:", "", 1, "L", false, 0, "")
			pdf.SetFont(reportFontFamily, "", 10)
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

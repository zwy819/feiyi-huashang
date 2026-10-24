package service

import (
	"bytes"
	"encoding/base64"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/jung-kurt/gofpdf"
)

//go:embed fonts/simhei.ttf
var simheiTTF []byte

//go:embed fonts/simkai.ttf
var simkaiTTF []byte

// 嵌入字体族：楷体做标题（书卷气），黑体做正文（可读性）。
// gofpdf 内置 Helvetica 仅覆盖拉丁字符，中文会渲染为乱码。
const (
	fontKai = "kai"
	fontHei = "hei"
)

// 莫高窟配色，与 client/tailwind.config.js 的 mogao 色板保持一致
type pdfColor struct{ R, G, B int }

var (
	colorCinnabar     = pdfColor{196, 58, 49}   // 朱砂红 mogao-cinnabar
	colorCinnabarDark = pdfColor{131, 35, 30}   // 朱砂-700
	colorGold         = pdfColor{184, 134, 45}  // 描金 mogao-gold
	colorOchre        = pdfColor{200, 155, 94}  // 土黄 mogao-ochre
	colorOchreDeep    = pdfColor{143, 102, 57}  // 土黄-700
	colorLapis        = pdfColor{46, 74, 122}   // 青金石蓝 mogao-lapis
	colorPaper        = pdfColor{252, 250, 247} // 铅白纸色 leadwhite-100
	colorCream        = pdfColor{245, 232, 216} // 信息条底色 ochre-100
	colorInk          = pdfColor{61, 48, 42}    // 墨色正文
	colorMuted        = pdfColor{138, 126, 110} // 弱化文字
)

// ImagePayload 报告内嵌图片（客户端离屏渲染的服饰预览图）
type ImagePayload struct {
	data []byte
	tp   string // gofpdf 图片类型："JPG" / "PNG"
}

type ReportService struct {
	garmentStore *model.GarmentStore
}

func NewReportService(garmentStore *model.GarmentStore) *ReportService {
	return &ReportService{garmentStore: garmentStore}
}

// ParseImageDataURL 解析客户端提交的 dataURL（仅接受 JPEG/PNG），非法返回 nil
func ParseImageDataURL(dataURL string) *ImagePayload {
	var tp string
	switch {
	case strings.HasPrefix(dataURL, "data:image/jpeg;base64,"):
		tp, dataURL = "JPG", strings.TrimPrefix(dataURL, "data:image/jpeg;base64,")
	case strings.HasPrefix(dataURL, "data:image/png;base64,"):
		tp, dataURL = "PNG", strings.TrimPrefix(dataURL, "data:image/png;base64,")
	default:
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil || len(data) == 0 {
		return nil
	}
	return &ImagePayload{data: data, tp: tp}
}

// GeneratePDF 生成收藏报告。images 为服饰 ID → 预览图（可缺失，缺失时该页显示占位面板）。
func (s *ReportService) GeneratePDF(ids []int64, images map[int64]*ImagePayload) ([]byte, error) {
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
	pdf.SetMargins(18, 20, 18)
	pdf.SetAutoPageBreak(true, 24)
	pdf.AliasNbPages("{nb}")
	pdf.AddUTF8FontFromBytes(fontKai, "", simkaiTTF)
	pdf.AddUTF8FontFromBytes(fontKai, "B", simkaiTTF)
	pdf.AddUTF8FontFromBytes(fontHei, "", simheiTTF)
	pdf.AddUTF8FontFromBytes(fontHei, "B", simheiTTF)

	writeCover(pdf, len(garments))

	for _, g := range garments {
		writeGarmentPage(pdf, g, images[g.ID])
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func setFill(pdf *gofpdf.Fpdf, c pdfColor)  { pdf.SetFillColor(c.R, c.G, c.B) }
func setText(pdf *gofpdf.Fpdf, c pdfColor)  { pdf.SetTextColor(c.R, c.G, c.B) }
func setDraw(pdf *gofpdf.Fpdf, c pdfColor)  { pdf.SetDrawColor(c.R, c.G, c.B) }

func writeCover(pdf *gofpdf.Fpdf, count int) {
	pdf.AddPage()

	// 封面全部为固定位置的装饰内容，关闭自动分页，
	// 避免接近页底时单元格越过分页线（297-24=273mm）触发隐式换页
	pdf.SetAutoPageBreak(false, 0)

	// 纸色底
	setFill(pdf, colorPaper)
	pdf.Rect(0, 0, 210, 297, "F")

	// 双线装饰框：外粗金线 + 内细土黄线
	setDraw(pdf, colorGold)
	pdf.SetLineWidth(0.8)
	pdf.Rect(10, 10, 190, 277, "")
	setDraw(pdf, colorOchre)
	pdf.SetLineWidth(0.2)
	pdf.Rect(13, 13, 184, 271, "")

	// 顶部题签：朱砂小方块 + 馆名
	pdf.SetXY(0, 34)
	setText(pdf, colorCinnabar)
	pdf.SetFont(fontHei, "", 10)
	pdf.CellFormat(0, 6, "中 国 非 物 质 文 化 遗 产 服 饰 数 字 博 物 馆", "", 0, "C", false, 0, "")

	// 金色分隔饰线（中央菱形）
	drawDiamondDivider(pdf, 62)

	// 主标题（楷体，疏散排布）
	pdf.SetY(88)
	setText(pdf, colorCinnabar)
	pdf.SetFont(fontKai, "", 46)
	pdf.CellFormat(0, 22, "非 遗 华 裳", "", 1, "C", false, 0, "")

	// 报告名
	pdf.SetY(122)
	setText(pdf, colorInk)
	pdf.SetFont(fontKai, "", 26)
	pdf.CellFormat(0, 14, "收 藏 报 告", "", 1, "C", false, 0, "")

	drawDiamondDivider(pdf, 146)

	// 元信息
	pdf.SetY(158)
	setText(pdf, colorMuted)
	pdf.SetFont(fontHei, "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("共收录  %d  件非遗服饰", count), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 7, time.Now().Format("2006 年 01 月 02 日"), "", 1, "C", false, 0, "")

	// 底部落款
	pdf.SetY(262)
	setText(pdf, colorOchreDeep)
	pdf.SetFont(fontKai, "", 11)
	pdf.CellFormat(0, 7, "数字传承非物质文化遗产", "", 1, "C", false, 0, "")
	setText(pdf, colorMuted)
	pdf.SetFont(fontHei, "", 8.5)
	pdf.CellFormat(0, 6, "本报告由 非遗华裳 3D 数字博物馆 自动生成", "", 1, "C", false, 0, "")
}

// drawDiamondDivider 居中金色菱形分隔饰线
func drawDiamondDivider(pdf *gofpdf.Fpdf, y float64) {
	cx, r := 105.0, 2.2
	setFill(pdf, colorGold)
	pdf.Polygon(
		[]gofpdf.PointType{{X: cx, Y: y - r}, {X: cx + r, Y: y}, {X: cx, Y: y + r}, {X: cx - r, Y: y}},
		"F",
	)
	setDraw(pdf, colorOchre)
	pdf.SetLineWidth(0.25)
	pdf.Line(cx-34, y, cx-r-3, y)
	pdf.Line(cx+r+3, y, cx+34, y)
}

func writeGarmentPage(pdf *gofpdf.Fpdf, g *model.Garment, img *ImagePayload) {
	pdf.AddPage()

	// 色带、图片面板、信息条、页脚都是固定位置内容，关闭自动分页防止隐式换页；
	// 仅在最后的"文化寓意/描述"正文段落前重新开启，让超长文本自然流向下一页
	pdf.SetAutoPageBreak(false, 0)

	// 纸色底
	setFill(pdf, colorPaper)
	pdf.Rect(0, 0, 210, 297, "F")

	// 顶部朱砂标题色带 + 金色压边线
	setFill(pdf, colorCinnabar)
	pdf.Rect(0, 0, 210, 26, "F")
	setDraw(pdf, colorGold)
	pdf.SetLineWidth(0.6)
	pdf.Line(0, 26, 210, 26)

	// 朝代处理："清"→"清代"；"现代/古代"这类已含"代"的原样保留
	dynasty := g.Dynasty
	if dynasty != "" && !strings.HasSuffix(dynasty, "代") {
		dynasty += "代"
	}

	// 色带内：名称（白字楷体）+ 右侧朝代·民族标签
	setText(pdf, pdfColor{255, 255, 255})
	pdf.SetFont(fontKai, "", 19)
	pdf.SetXY(18, 6)
	pdf.CellFormat(130, 10, g.Name, "", 0, "L", false, 0, "")
	pdf.SetFont(fontHei, "", 10)
	pdf.SetXY(140, 6)
	pdf.CellFormat(52, 10, dynasty+" · "+g.Ethnic, "", 0, "R", false, 0, "")

	// 页脚（AliasNbPages 会在输出时把 {nb} 替换为总页数）
	setDraw(pdf, colorOchre)
	pdf.SetLineWidth(0.2)
	pdf.Line(18, 284, 192, 284)
	setText(pdf, colorMuted)
	pdf.SetFont(fontHei, "", 8)
	pdf.SetXY(18, 286)
	pdf.CellFormat(90, 5, "非遗华裳 · 收藏报告", "", 0, "L", false, 0, "")
	pdf.SetXY(102, 286)
	pdf.CellFormat(90, 5, fmt.Sprintf("第 %d 页 / 共 {nb} 页", pdf.PageNo()), "", 0, "R", false, 0, "")

	// 服饰预览图面板（120mm 见方，居中）
	const (
		panelX = 45.0
		panelY = 34.0
		panelW = 120.0
	)
	if img != nil {
		imgID := fmt.Sprintf("garment-%d", g.ID)
		pdf.RegisterImageReader(imgID, img.tp, bytes.NewReader(img.data))
		pdf.Image(imgID, panelX, panelY, panelW, panelW, false, img.tp, 0, "")
	} else {
		// 占位面板
		setFill(pdf, colorCream)
		pdf.Rect(panelX, panelY, panelW, panelW, "F")
		setText(pdf, colorMuted)
		pdf.SetFont(fontHei, "", 10)
		pdf.SetXY(panelX, panelY+panelW/2-3)
		pdf.CellFormat(panelW, 6, "暂无预览图", "", 0, "C", false, 0, "")
	}
	setDraw(pdf, colorOchre)
	pdf.SetLineWidth(0.3)
	pdf.Rect(panelX, panelY, panelW, panelW, "")

	// 图注
	setText(pdf, colorOchreDeep)
	pdf.SetFont(fontKai, "", 9.5)
	pdf.SetXY(0, panelY+panelW+3)
	pdf.CellFormat(0, 6, "· 服 饰 三 维 预 览 ·", "", 1, "C", false, 0, "")

	// 基本信息区：两列三行信息条
	info := [][2]string{
		{"朝代", dynasty},
		{"民族", g.Ethnic},
		{"工艺", g.Craft},
		{"材质", g.Material},
		{"纹样", g.Pattern},
		{"省份", g.Province},
	}
	const (
		colW   = 86.0
		colGap = 2.0
		rowH   = 11.0
	)
	baseY := panelY + panelW + 13.0
	for i, item := range info {
		if item[1] == "" {
			continue
		}
		col, row := i%2, i/2
		x := 18.0 + float64(col)*(colW+colGap)
		y := baseY + float64(row)*(rowH+1.6)
		setFill(pdf, colorCream)
		pdf.Rect(x, y, colW, rowH, "F")
		// 左侧朱砂细条
		setFill(pdf, colorCinnabar)
		pdf.Rect(x, y, 0.9, rowH, "F")
		setText(pdf, colorCinnabarDark)
		pdf.SetFont(fontHei, "", 9)
		pdf.SetXY(x+3, y)
		pdf.CellFormat(14, rowH, item[0], "", 0, "L", false, 0, "")
		setText(pdf, colorInk)
		pdf.SetFont(fontHei, "", 10.5)
		pdf.SetXY(x+18, y)
		pdf.CellFormat(colW-20, rowH, item[1], "", 0, "L", false, 0, "")
	}

	// 文化寓意 / 描述段落（重新开启自动分页，超长文本自然续页）
	pdf.SetAutoPageBreak(true, 24)
	y := baseY + 3*(rowH+1.6) + 6
	y = writeSection(pdf, "文化寓意", g.CulturalMeaning, y)
	writeSection(pdf, "描述", g.Description, y)
}

// writeSection 段落标题（朱砂方点 + 青金石标题 + 金色下划线）与正文，返回下一段落起始 y
func writeSection(pdf *gofpdf.Fpdf, title, body string, y float64) float64 {
	if body == "" {
		return y
	}
	// 标题
	setFill(pdf, colorCinnabar)
	pdf.Rect(18, y+1.1, 2.4, 2.4, "F")
	setText(pdf, colorLapis)
	pdf.SetFont(fontHei, "", 12.5)
	pdf.SetXY(23, y)
	pdf.CellFormat(60, 5.5, title, "", 0, "L", false, 0, "")
	setDraw(pdf, colorGold)
	pdf.SetLineWidth(0.4)
	pdf.Line(23, y+6.6, 53, y+6.6)

	// 正文
	setText(pdf, colorInk)
	pdf.SetFont(fontHei, "", 10.5)
	pdf.SetXY(18, y+8.5)
	pdf.MultiCell(0, 6.2, body, "", "L", false)

	_, afterY := pdf.GetXY()
	return afterY + 3.5
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

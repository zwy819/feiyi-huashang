package handler

import (
	"fmt"
	"net/http"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// reportRequest POST /api/report/generate 的请求体。
// images 为服饰 ID → 客户端离屏渲染的预览图 dataURL（JPEG/PNG），可缺省。
type reportRequest struct {
	IDs    []int64           `json:"ids"`
	Images map[string]string `json:"images"`
}

func (h *ReportHandler) Generate(c *gin.Context) {
	var (
		ids    []int64
		images = map[int64]*service.ImagePayload(nil)
	)

	var req reportRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		// POST：JSON 携带 ids 与预览图
		ids = req.IDs
		if len(req.Images) > 0 {
			images = make(map[int64]*service.ImagePayload, len(req.Images))
			for gidStr, dataURL := range req.Images {
				var gid int64
				if _, err := fmt.Sscanf(gidStr, "%d", &gid); err != nil || gid <= 0 {
					continue
				}
				if payload := service.ParseImageDataURL(dataURL); payload != nil {
					images[gid] = payload
				}
			}
		}
	} else {
		// GET 兼容：?ids=1,2,3（无预览图，报告使用占位面板）
		ids = service.ParseIDs(c.Query("ids"))
	}

	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "请提供有效的服饰 ID 列表"))
		return
	}

	pdfBytes, err := h.reportService.GeneratePDF(ids, images)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, fmt.Sprintf("生成报告失败: %v", err)))
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"nonlegacy-report-%d.pdf\"", len(ids)))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

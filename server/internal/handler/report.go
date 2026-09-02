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

func (h *ReportHandler) Generate(c *gin.Context) {
	idsStr := c.Query("ids")
	ids := service.ParseIDs(idsStr)
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "请提供有效的服饰 ID 列表"))
		return
	}

	pdfBytes, err := h.reportService.GeneratePDF(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, fmt.Sprintf("生成报告失败: %v", err)))
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"nonlegacy-report-%d.pdf\"", len(ids)))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

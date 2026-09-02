package handler

import (
	"net/http"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	statsService *service.StatisticsService
}

func NewStatisticsHandler(statsService *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{statsService: statsService}
}

func (h *StatisticsHandler) GetGeoDistribution(c *gin.Context) {
	data, err := h.statsService.GetGeoDistribution()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取地理分布数据失败"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(data))
}

func (h *StatisticsHandler) GetCraftFlow(c *gin.Context) {
	data, err := h.statsService.GetCraftFlow()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取工艺流向数据失败"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(data))
}

func (h *StatisticsHandler) GetOverview(c *gin.Context) {
	data, err := h.statsService.GetOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取概览统计失败"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(data))
}

func (h *StatisticsHandler) GetSankey(c *gin.Context) {
	data, err := h.statsService.GetSankeyData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取桑基图数据失败"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(data))
}

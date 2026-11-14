package handler

import (
	"net/http"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/tour"

	"github.com/gin-gonic/gin"
)

type TourHandler struct {
	service *tour.Service
}

func NewTourHandler(service *tour.Service) *TourHandler {
	return &TourHandler{service: service}
}

type tourPlanRequest struct {
	Query string `json:"query"`
}

func (h *TourHandler) Plan(c *gin.Context) {
	var req tourPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求格式不正确"))
		return
	}
	if len([]rune(req.Query)) > 200 {
		req.Query = string([]rune(req.Query)[:200])
	}
	plan, err := h.service.Plan(c.Request.Context(), strings.TrimSpace(req.Query))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.ErrorResponse(50203, "路线生成失败，请稍后重试"))
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Data: plan})
}

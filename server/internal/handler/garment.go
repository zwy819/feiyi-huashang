package handler

import (
	"net/http"
	"strconv"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

type GarmentHandler struct {
	garmentStore    *model.GarmentStore
	garmentService  *service.GarmentService
	dailyService    *service.DailyService
}

func NewGarmentHandler(garmentStore *model.GarmentStore, garmentService *service.GarmentService, dailyService *service.DailyService) *GarmentHandler {
	return &GarmentHandler{
		garmentStore:   garmentStore,
		garmentService: garmentService,
		dailyService:   dailyService,
	}
}

func (h *GarmentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	params := model.GarmentQueryParams{
		Page:    page,
		Limit:   limit,
		Search:  c.Query("search"),
		Dynasty: c.Query("dynasty"),
		Ethnic:  c.Query("ethnic"),
		Craft:   c.Query("craft"),
	}

	garments, total, err := h.garmentStore.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取服饰列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: garments,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

func (h *GarmentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	garment, err := h.garmentStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "服饰不存在"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(garment))
}

func (h *GarmentHandler) GetRecommendChain(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	items, err := h.garmentService.GetRecommendChain(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "服饰不存在"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(map[string]interface{}{
		"items": items,
	}))
}

func (h *GarmentHandler) GetDaily(c *gin.Context) {
	garment, err := h.dailyService.GetDaily()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取每日推荐失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(garment))
}

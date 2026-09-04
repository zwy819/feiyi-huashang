package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

// AnnotationHandler 服饰热点标注 handler
type AnnotationHandler struct {
	annotationStore *model.AnnotationStore
	garmentStore    *model.GarmentStore
}

// NewAnnotationHandler 创建热点标注 handler
func NewAnnotationHandler(annotationStore *model.AnnotationStore, garmentStore *model.GarmentStore) *AnnotationHandler {
	return &AnnotationHandler{
		annotationStore: annotationStore,
		garmentStore:    garmentStore,
	}
}

// -------- 公开端 --------

// ListByGarment 某服饰的热点标注列表 GET /api/garments/:id/annotations
func (h *AnnotationHandler) ListByGarment(c *gin.Context) {
	garmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || garmentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	list, err := h.annotationStore.ListByGarment(garmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取热点标注失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(list))
}

// -------- 管理端 --------

// AdminList 管理员标注列表 GET /api/admin/annotations?garment_id=1
func (h *AnnotationHandler) AdminList(c *gin.Context) {
	garmentID := int64(0)
	if raw := c.Query("garment_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数 garment_id 不合法"))
			return
		}
		garmentID = id
	}

	list, err := h.annotationStore.List(garmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取热点标注失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(list))
}

type annotationBody struct {
	GarmentID int64   `json:"garment_id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	PositionZ float64 `json:"position_z"`
	SortOrder int     `json:"sort_order"`
}

// validateAnnotationBody 校验标注请求体
func validateAnnotationBody(body *annotationBody) (string, bool) {
	body.Title = strings.TrimSpace(body.Title)
	body.Content = strings.TrimSpace(body.Content)

	if body.Title == "" {
		return "标注标题不能为空", false
	}
	if len([]rune(body.Title)) > 100 {
		return "标注标题不能超过 100 字", false
	}
	if len([]rune(body.Content)) > 1000 {
		return "标注内容不能超过 1000 字", false
	}
	if math.IsNaN(body.PositionX) || math.IsInf(body.PositionX, 0) ||
		math.IsNaN(body.PositionY) || math.IsInf(body.PositionY, 0) ||
		math.IsNaN(body.PositionZ) || math.IsInf(body.PositionZ, 0) {
		return "热点坐标不合法", false
	}
	return "", true
}

// AdminCreate 创建标注 POST /api/admin/annotations
func (h *AnnotationHandler) AdminCreate(c *gin.Context) {
	var body annotationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	if msg, ok := validateAnnotationBody(&body); !ok {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, msg))
		return
	}
	if body.GarmentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数 garment_id 不合法"))
		return
	}
	if _, err := h.garmentStore.GetByID(body.GarmentID); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "服饰不存在"))
		return
	}

	a := &model.Annotation{
		GarmentID: body.GarmentID,
		Title:     body.Title,
		Content:   body.Content,
		PositionX: body.PositionX,
		PositionY: body.PositionY,
		PositionZ: body.PositionZ,
		SortOrder: body.SortOrder,
	}
	id, err := h.annotationStore.Create(a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "创建热点标注失败"))
		return
	}

	created, _ := h.annotationStore.GetByID(id)
	c.JSON(http.StatusCreated, model.SuccessMessage("热点标注已创建", created))
}

// AdminUpdate 更新标注 PUT /api/admin/annotations/:id
func (h *AnnotationHandler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	existing, err := h.annotationStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "热点标注不存在"))
		return
	}

	var body annotationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}
	if msg, ok := validateAnnotationBody(&body); !ok {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, msg))
		return
	}

	existing.Title = body.Title
	existing.Content = body.Content
	existing.PositionX = body.PositionX
	existing.PositionY = body.PositionY
	existing.PositionZ = body.PositionZ
	existing.SortOrder = body.SortOrder

	if err := h.annotationStore.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "更新热点标注失败"))
		return
	}

	updated, _ := h.annotationStore.GetByID(id)
	c.JSON(http.StatusOK, model.SuccessMessage("热点标注已更新", updated))
}

// AdminDelete 删除标注 DELETE /api/admin/annotations/:id
func (h *AnnotationHandler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.annotationStore.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "热点标注不存在"))
		return
	}

	if err := h.annotationStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "删除热点标注失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("热点标注已删除", nil))
}

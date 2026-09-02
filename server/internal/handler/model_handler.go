package handler

import (
	"net/http"
	"os"
	"strconv"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelHandler struct {
	modelStore   *model.ModelStore
	modelService *service.ModelService
	garmentStore *model.GarmentStore
}

func NewModelHandler(modelStore *model.ModelStore, modelService *service.ModelService, garmentStore *model.GarmentStore) *ModelHandler {
	return &ModelHandler{modelStore: modelStore, modelService: modelService, garmentStore: garmentStore}
}

// ListPublic 公开模型列表（不含 file_path）
func (h *ModelHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var garmentID *int64
	if gidStr := c.Query("garment_id"); gidStr != "" {
		if gid, err := strconv.ParseInt(gidStr, 10, 64); err == nil {
			garmentID = &gid
		}
	}

	params := model.ModelQueryParams{
		Page:      page,
		Limit:     limit,
		GarmentID: garmentID,
		Search:    c.Query("search"),
	}

	models, total, err := h.modelStore.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取模型列表失败"))
		return
	}

	// 公开接口不暴露 file_path
	type publicModel struct {
		ID           int64  `json:"id"`
		Filename     string `json:"filename"`
		OriginalName string `json:"original_name"`
		FileSize     int64  `json:"file_size"`
		MimeType     string `json:"mime_type"`
		GarmentID    *int64 `json:"garment_id"`
		GarmentName  string `json:"garment_name"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}

	items := make([]publicModel, 0, len(models))
	for _, m := range models {
		items = append(items, publicModel{
			ID:           m.ID,
			Filename:     m.Filename,
			OriginalName: m.OriginalName,
			FileSize:     m.FileSize,
			MimeType:     m.MimeType,
			GarmentID:    m.GarmentID,
			GarmentName:  m.GarmentName,
			CreatedAt:    m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

// ServeFile 提供模型文件下载
func (h *ModelHandler) ServeFile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	m, err := h.modelStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "模型文件不存在"))
		return
	}

	if _, err := os.Stat(m.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "模型文件不存在"))
		return
	}

	c.Header("Content-Type", m.MimeType)
	c.Header("Content-Disposition", "inline; filename=\""+m.OriginalName+"\"")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(m.FilePath)
}

// --- Admin Model Handlers ---

func (h *ModelHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var garmentID *int64
	if gidStr := c.Query("garment_id"); gidStr != "" {
		if gid, err := strconv.ParseInt(gidStr, 10, 64); err == nil {
			garmentID = &gid
		}
	}

	params := model.ModelQueryParams{
		Page:      page,
		Limit:     limit,
		GarmentID: garmentID,
		Search:    c.Query("search"),
	}

	models, total, err := h.modelStore.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取模型列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: models,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

func (h *ModelHandler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	m, err := h.modelStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "模型不存在"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(m))
}

func (h *ModelHandler) AdminUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "未提供文件"))
		return
	}

	var garmentID *int64
	if gidStr := c.PostForm("garment_id"); gidStr != "" {
		if gid, err := strconv.ParseInt(gidStr, 10, 64); err == nil && gid > 0 {
			garmentID = &gid
		}
	}

	m, err := h.modelService.Upload(file, garmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, model.SuccessMessage("模型上传成功", m))
}

func (h *ModelHandler) AdminUpdate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	existing, err := h.modelStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "模型不存在"))
		return
	}

	var body struct {
		OriginalName *string `json:"original_name"`
		GarmentID    *int64  `json:"garment_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求体为空"))
		return
	}

	if body.OriginalName != nil && *body.OriginalName != "" {
		existing.OriginalName = *body.OriginalName
	}
	if body.GarmentID != nil {
		existing.GarmentID = body.GarmentID
	}

	if err := h.modelStore.Update(id, existing); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "更新模型信息失败"))
		return
	}

	// 同步更新 garment 的 model_id
	if body.GarmentID != nil && *body.GarmentID > 0 {
		h.garmentStore.UpdateModelID(*body.GarmentID, &id)
	}

	updated, _ := h.modelStore.GetByID(id)
	c.JSON(http.StatusOK, model.SuccessMessage("模型信息已更新", updated))
}

func (h *ModelHandler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if err := h.modelService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("模型已删除", nil))
}

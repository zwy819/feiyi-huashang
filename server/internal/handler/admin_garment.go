package handler

import (
	"net/http"
	"strconv"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

type AdminGarmentHandler struct {
	garmentStore *model.GarmentStore
}

func NewAdminGarmentHandler(garmentStore *model.GarmentStore) *AdminGarmentHandler {
	return &AdminGarmentHandler{garmentStore: garmentStore}
}

func (h *AdminGarmentHandler) Create(c *gin.Context) {
	var g model.Garment
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	if g.Name == "" || g.Dynasty == "" || g.Ethnic == "" || g.Province == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "名称、朝代、民族、省份为必填项"))
		return
	}

	id, err := h.garmentStore.Create(&g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "创建服饰失败"))
		return
	}

	created, _ := h.garmentStore.GetByID(id)
	c.JSON(http.StatusCreated, model.SuccessMessage("服饰创建成功", created))
}

func (h *AdminGarmentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	existing, err := h.garmentStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "服饰不存在"))
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求体为空"))
		return
	}

	// 只更新提供的字段
	if v, ok := body["name"]; ok {
		existing.Name = v.(string)
	}
	if v, ok := body["dynasty"]; ok {
		existing.Dynasty = v.(string)
	}
	if v, ok := body["ethnic"]; ok {
		existing.Ethnic = v.(string)
	}
	if v, ok := body["craft"]; ok {
		existing.Craft = v.(string)
	}
	if v, ok := body["material"]; ok {
		existing.Material = v.(string)
	}
	if v, ok := body["pattern"]; ok {
		existing.Pattern = v.(string)
	}
	if v, ok := body["cultural_meaning"]; ok {
		existing.CulturalMeaning = v.(string)
	}
	if v, ok := body["province"]; ok {
		existing.Province = v.(string)
	}
	if v, ok := body["description"]; ok {
		existing.Description = v.(string)
	}
	if v, ok := body["image_url"]; ok {
		existing.ImageURL = v.(string)
	}
	if v, ok := body["model_url"]; ok {
		existing.ModelURL = v.(string)
	}
	if v, ok := body["model_id"]; ok {
		if v == nil {
			existing.ModelID = nil
		} else {
			mid := int64(v.(float64))
			existing.ModelID = &mid
		}
	}

	if err := h.garmentStore.Update(id, existing); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "更新服饰信息失败"))
		return
	}

	updated, _ := h.garmentStore.GetByID(id)
	c.JSON(http.StatusOK, model.SuccessMessage("服饰信息已更新", updated))
}

func (h *AdminGarmentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.garmentStore.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "服饰不存在"))
		return
	}

	if err := h.garmentStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "删除服饰失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("服饰已删除", nil))
}

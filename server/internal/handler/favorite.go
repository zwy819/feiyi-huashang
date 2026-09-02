package handler

import (
	"net/http"
	"strconv"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	favoriteStore *model.FavoriteStore
}

func NewFavoriteHandler(favoriteStore *model.FavoriteStore) *FavoriteHandler {
	return &FavoriteHandler{favoriteStore: favoriteStore}
}

func (h *FavoriteHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	favs, err := h.favoriteStore.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取收藏列表失败"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(map[string]interface{}{"favorites": favs}))
}

func (h *FavoriteHandler) Add(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var body struct {
		GarmentID int64 `json:"garment_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	if body.GarmentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	err := h.favoriteStore.Add(userID, body.GarmentID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, model.ErrorResponse(40901, "已收藏该服饰"))
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "添加收藏失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("收藏成功", nil))
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID := c.GetInt64("user_id")
	garmentID, err := strconv.ParseInt(c.Param("garment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if err := h.favoriteStore.Remove(userID, garmentID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "取消收藏失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("已取消收藏", nil))
}

func (h *FavoriteHandler) Check(c *gin.Context) {
	userID := c.GetInt64("user_id")
	garmentID, err := strconv.ParseInt(c.Param("garment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	isFav, err := h.favoriteStore.IsFavorited(userID, garmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "检查收藏状态失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(map[string]bool{"isFavorited": isFav}))
}

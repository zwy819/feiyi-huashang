package handler

import (
	"net/http"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

type AdminDashboardHandler struct {
	garmentStore  *model.GarmentStore
	modelStore    *model.ModelStore
	userStore     *model.UserStore
	favoriteStore *model.FavoriteStore
}

func NewAdminDashboardHandler(garmentStore *model.GarmentStore, modelStore *model.ModelStore, userStore *model.UserStore, favoriteStore *model.FavoriteStore) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		garmentStore:  garmentStore,
		modelStore:    modelStore,
		userStore:     userStore,
		favoriteStore: favoriteStore,
	}
}

func (h *AdminDashboardHandler) GetDashboard(c *gin.Context) {
	totalGarments, _ := h.garmentStore.Count()
	totalModels, _ := h.modelStore.Count()
	totalUsers, _ := h.userStore.Count()
	totalFavorites, _ := h.favoriteStore.Count()

	// Recent uploads
	recentModels, _, _ := h.modelStore.List(model.ModelQueryParams{Page: 1, Limit: 5})

	c.JSON(http.StatusOK, model.SuccessResponse(map[string]interface{}{
		"total_garments":  totalGarments,
		"total_models":    totalModels,
		"total_users":     totalUsers,
		"total_favorites": totalFavorites,
		"recent_uploads":  recentModels,
	}))
}

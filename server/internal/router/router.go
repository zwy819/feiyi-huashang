package router

import (
	"database/sql"

	"nonlegacy-garment-3d/server/internal/config"
	"nonlegacy-garment-3d/server/internal/handler"
	"nonlegacy-garment-3d/server/internal/middleware"
	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	db *sql.DB,
	garmentStore *model.GarmentStore,
	modelStore *model.ModelStore,
	userStore *model.UserStore,
	favoriteStore *model.FavoriteStore,
	dailyPickStore *model.DailyPickStore,
	submissionStore *model.SubmissionStore,
	cfg *config.Config,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg.CORSOrigin))

	// --- Services ---
	authService := service.NewAuthService(userStore, cfg.JWTSecret, cfg.JWTExpiresIn)
	garmentService := service.NewGarmentService(garmentStore)
	dailyService := service.NewDailyService(dailyPickStore, garmentStore)
	statsService := service.NewStatisticsService(db, garmentStore, favoriteStore)
	reportService := service.NewReportService(garmentStore)
	modelService := service.NewModelService(modelStore, garmentStore, cfg.UploadDir, cfg.MaxUploadSize)
	submissionService := service.NewSubmissionService(submissionStore, cfg.UploadDir, cfg.MaxImageUploadSize)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService)
	garmentHandler := handler.NewGarmentHandler(garmentStore, garmentService, dailyService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteStore)
	statsHandler := handler.NewStatisticsHandler(statsService)
	reportHandler := handler.NewReportHandler(reportService)
	modelHandler := handler.NewModelHandler(modelStore, modelService, garmentStore)
	adminGarmentHandler := handler.NewAdminGarmentHandler(garmentStore)
	dashboardHandler := handler.NewAdminDashboardHandler(garmentStore, modelStore, userStore, favoriteStore)
	submissionHandler := handler.NewSubmissionHandler(submissionStore, submissionService, garmentStore, modelStore)

	// --- Public Routes ---
	api := r.Group("/api")
	{
		// Health
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, model.Response{Code: 0, Message: nil, Data: "ok"})
		})

		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile",
				middleware.JWT(cfg.JWTSecret, true),
				authHandler.Profile,
			)
		}

		// Garments
		garments := api.Group("/garments")
		{
			garments.GET("", garmentHandler.List)
			garments.GET("/:id", garmentHandler.GetByID)
			garments.GET("/:id/recommend", garmentHandler.GetRecommendChain)
		}

		// Daily
		api.GET("/daily", garmentHandler.GetDaily)

		// Favorites (JWT required)
		favorites := api.Group("/favorites")
		favorites.Use(middleware.JWT(cfg.JWTSecret, true))
		{
			favorites.GET("", favoriteHandler.List)
			favorites.POST("", favoriteHandler.Add)
			favorites.DELETE("/:garment_id", favoriteHandler.Remove)
			favorites.GET("/check/:garment_id", favoriteHandler.Check)
		}

		// Statistics
		stats := api.Group("/statistics")
		{
			stats.GET("/geo", statsHandler.GetGeoDistribution)
			stats.GET("/craft-flow", statsHandler.GetCraftFlow)
			stats.GET("/overview", statsHandler.GetOverview)
			stats.GET("/sankey", statsHandler.GetSankey)
		}

		// Report
		api.GET("/report/generate", reportHandler.Generate)

		// Models (public)
		models := api.Group("/models")
		{
			models.GET("", modelHandler.ListPublic)
			models.GET("/:id/file", modelHandler.ServeFile)
		}

		// Submissions — 公开的"用户贡献"列表
		api.GET("/public-submissions", submissionHandler.ListPublic)

		// Submission image — 公开图片访问
		api.GET("/submissions/:id/image", submissionHandler.ServeImage)
	}

	// --- User Routes (JWT required) ---
	userRoutes := api.Group("")
	userRoutes.Use(middleware.JWT(cfg.JWTSecret, true))
	{
		submissions := userRoutes.Group("/submissions")
		{
			submissions.POST("", submissionHandler.Submit)
			submissions.GET("", submissionHandler.ListMy)
			submissions.GET("/:id", submissionHandler.GetMy)
				submissions.PUT("/:id/retract", submissionHandler.Retract)
		}
	}

	// --- Admin Routes (JWT + admin role required) ---
	admin := r.Group("/api/admin")
	admin.Use(middleware.JWT(cfg.JWTSecret, true))
	admin.Use(middleware.AdminOnly())
	{
		admin.GET("/dashboard", dashboardHandler.GetDashboard)

		adminModels := admin.Group("/models")
		{
			adminModels.GET("", modelHandler.AdminList)
			adminModels.GET("/:id", modelHandler.AdminGet)
			adminModels.POST("", modelHandler.AdminUpload)
			adminModels.PUT("/:id", modelHandler.AdminUpdate)
			adminModels.DELETE("/:id", modelHandler.AdminDelete)
		}

		adminGarments := admin.Group("/garments")
		{
			adminGarments.POST("", adminGarmentHandler.Create)
			adminGarments.PUT("/:id", adminGarmentHandler.Update)
			adminGarments.DELETE("/:id", adminGarmentHandler.Delete)
		}

		adminSubmissions := admin.Group("/submissions")
		{
			adminSubmissions.GET("", submissionHandler.AdminList)
			adminSubmissions.GET("/:id", submissionHandler.AdminGet)
			adminSubmissions.PUT("/:id/approve", submissionHandler.AdminApprove)
			adminSubmissions.PUT("/:id/reject", submissionHandler.AdminReject)
			adminSubmissions.PUT("/:id/complete", submissionHandler.AdminComplete)
			adminSubmissions.DELETE("/:id", submissionHandler.AdminDelete)
		}
	}

	return r
}

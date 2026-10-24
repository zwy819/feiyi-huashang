package router

import (
	"database/sql"

	"nonlegacy-garment-3d/server/internal/ai"
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
	commentStore *model.CommentStore,
	annotationStore *model.AnnotationStore,
	quizStore *model.QuizStore,
	aiService *ai.Service,
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
	quizService := service.NewQuizService(quizStore)

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
	commentHandler := handler.NewCommentHandler(commentStore, garmentStore)
	annotationHandler := handler.NewAnnotationHandler(annotationStore, garmentStore)
	quizHandler := handler.NewQuizHandler(quizStore, quizService)
	aiHandler := handler.NewAIHandler(aiService)

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
			garments.GET("/:id/annotations", annotationHandler.ListByGarment)
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

		// Report（POST 携带客户端渲染的预览图；GET 兼容旧调用，无图占位）
		api.GET("/report/generate", reportHandler.Generate)
		api.POST("/report/generate", reportHandler.Generate)

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

		// Comments — 评论列表（可选登录：登录后额外可见自己的待审核/已驳回评论）
		api.GET("/comments", middleware.JWT(cfg.JWTSecret, false), commentHandler.List)

		// Quiz — 华裳问答（出题与排行榜公开；提交判分可选登录：未登录判分但不记录成绩）
		quiz := api.Group("/quiz")
		{
			quiz.GET("/questions", quizHandler.GetQuestions)
			quiz.GET("/leaderboard", quizHandler.Leaderboard)
			quiz.POST("/submit", middleware.JWT(cfg.JWTSecret, false), quizHandler.Submit)
			quiz.POST("/check", quizHandler.Check)
		}

		// AI docent — anonymous, rate-limited in the AI service.
		aiGroup := api.Group("/ai")
		{
			aiGroup.GET("/status", aiHandler.Status)
			aiGroup.POST("/chat", aiHandler.Chat)
			aiGroup.POST("/feedback", aiHandler.Feedback)
		}
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

		// Comments — 评论写操作（发表 / 编辑 / 删除 / 点赞）
		comments := userRoutes.Group("/comments")
		{
			comments.POST("", commentHandler.Create)
			comments.PUT("/:id", commentHandler.Update)
			comments.DELETE("/:id", commentHandler.Delete)
			comments.POST("/:id/like", commentHandler.Like)
			comments.DELETE("/:id/like", commentHandler.Unlike)
		}

		// Quiz — 个人战绩（需登录）
		userRoutes.GET("/quiz/records", quizHandler.MyRecords)
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

		// 评论审核
		adminComments := admin.Group("/comments")
		{
			adminComments.GET("", commentHandler.AdminList)
			adminComments.PUT("/:id/approve", commentHandler.AdminApprove)
			adminComments.PUT("/:id/reject", commentHandler.AdminReject)
			adminComments.DELETE("/:id", commentHandler.AdminDelete)
		}

		// 热点标注管理
		adminAnnotations := admin.Group("/annotations")
		{
			adminAnnotations.GET("", annotationHandler.AdminList)
			adminAnnotations.POST("", annotationHandler.AdminCreate)
			adminAnnotations.PUT("/:id", annotationHandler.AdminUpdate)
			adminAnnotations.DELETE("/:id", annotationHandler.AdminDelete)
		}

		// AI 知识库重建（管理员）
		admin.POST("/ai/reindex", aiHandler.Reindex)
	}

	return r
}

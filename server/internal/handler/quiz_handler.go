package handler

import (
	"net/http"
	"strconv"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

// QuizHandler 华裳问答 handler
type QuizHandler struct {
	quizStore   *model.QuizStore
	quizService *service.QuizService
}

// NewQuizHandler 创建问答 handler
func NewQuizHandler(quizStore *model.QuizStore, quizService *service.QuizService) *QuizHandler {
	return &QuizHandler{
		quizStore:   quizStore,
		quizService: quizService,
	}
}

// GetQuestions 生成题目 GET /api/quiz/questions?count=10（公开）
func (h *QuizHandler) GetQuestions(c *gin.Context) {
	count, err := strconv.Atoi(c.DefaultQuery("count", "10"))
	if err != nil || count < 1 {
		count = 10
	}
	if count > 20 {
		count = 20
	}

	questions, err := h.quizService.GenerateQuestions(count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "题目生成失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{
		"questions": questions,
	}))
}

// quizAnswerItem 提交的单题答案（choice 为所选选项文本）
type quizAnswerItem struct {
	GarmentID int64  `json:"garment_id"`
	Field     string `json:"field"`
	Choice    string `json:"choice"`
}

// Submit 提交答案 POST /api/quiz/submit（可选登录：未登录判分但不记录成绩）
// 服务端按 garment_id + field 重查真实值判分，防止前端篡改
func (h *QuizHandler) Submit(c *gin.Context) {
	userID := int64(0)
	if raw, exists := c.Get("user_id"); exists {
		if id, ok := raw.(int64); ok {
			userID = id
		}
	}

	var body struct {
		DurationSeconds int              `json:"duration_seconds"`
		Answers         []quizAnswerItem `json:"answers"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}
	if len(body.Answers) == 0 || len(body.Answers) > 50 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "答案数量不合法"))
		return
	}

	type resultItem struct {
		GarmentID int64  `json:"garment_id"`
		Field     string `json:"field"`
		Choice    string `json:"choice"`
		Answer    string `json:"answer"`
		Correct   bool   `json:"correct"`
	}
	results := make([]resultItem, 0, len(body.Answers))
	correctCount := 0

	for _, item := range body.Answers {
		field := strings.TrimSpace(item.Field)
		actual, err := h.quizStore.FieldValue(item.GarmentID, field)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "判分失败"))
			return
		}
		choice := strings.TrimSpace(item.Choice)
		correct := actual != "" && choice == actual
		if correct {
			correctCount++
		}
		results = append(results, resultItem{
			GarmentID: item.GarmentID,
			Field:     field,
			Choice:    choice,
			Answer:    actual,
			Correct:   correct,
		})
	}

	total := len(body.Answers)
	score := correctCount * 10
	duration := body.DurationSeconds
	if duration < 0 || duration > 3600 {
		duration = 0
	}

	// 登录用户才记录成绩
	if userID > 0 {
		if _, err := h.quizStore.InsertRecord(userID, score, total, correctCount, duration); err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "成绩记录失败"))
			return
		}
	}

	best, _ := h.quizStore.GetBestScore(userID)

	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{
		"score":          score,
		"total":          total,
		"correct_count":  correctCount,
		"best_score":     best,
		"duration":       duration,
		"results":        results,
	}))
}

// Check 单题即时判定 POST /api/quiz/check（公开）
// 选择后才知道正误与答案；最终成绩仍以 submit 的服务端判分为准
func (h *QuizHandler) Check(c *gin.Context) {
	var body quizAnswerItem
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	field := strings.TrimSpace(body.Field)
	actual, err := h.quizStore.FieldValue(body.GarmentID, field)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "判分失败"))
		return
	}
	correct := actual != "" && strings.TrimSpace(body.Choice) == actual

	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{
		"correct": correct,
		"answer":  actual,
	}))
}

// Leaderboard 排行榜 GET /api/quiz/leaderboard（公开）
func (h *QuizHandler) Leaderboard(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	records, err := h.quizStore.ListLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取排行榜失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{"items": records}))
}

// MyRecords 我的战绩 GET /api/quiz/records（JWT）
func (h *QuizHandler) MyRecords(c *gin.Context) {
	userID := c.GetInt64("user_id")

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	records, err := h.quizStore.ListByUser(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取战绩失败"))
		return
	}

	best, _ := h.quizStore.GetBestScore(userID)

	c.JSON(http.StatusOK, model.SuccessResponse(gin.H{
		"best":    best,
		"records": records,
	}))
}

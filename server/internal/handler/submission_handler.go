package handler

import (
	"net/http"
	"os"
	"strconv"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

// SubmissionHandler 投稿相关 handler
type SubmissionHandler struct {
	submissionStore   *model.SubmissionStore
	submissionService *service.SubmissionService
	garmentStore      *model.GarmentStore
	modelStore        *model.ModelStore
}

// NewSubmissionHandler 创建投稿 handler
func NewSubmissionHandler(
	submissionStore *model.SubmissionStore,
	submissionService *service.SubmissionService,
	garmentStore *model.GarmentStore,
	modelStore *model.ModelStore,
) *SubmissionHandler {
	return &SubmissionHandler{
		submissionStore:   submissionStore,
		submissionService: submissionService,
		garmentStore:      garmentStore,
		modelStore:        modelStore,
	}
}

// -------- 用户端 --------

// Submit 用户提交投稿 POST /api/submissions
func (h *SubmissionHandler) Submit(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(int64)

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "未提供图片文件"))
		return
	}

	metadata := service.SubmissionMetadata{
		Name:            c.PostForm("name"),
		Dynasty:         c.PostForm("dynasty"),
		Ethnic:          c.PostForm("ethnic"),
		Craft:           c.PostForm("craft"),
		Material:        c.PostForm("material"),
		Pattern:         c.PostForm("pattern"),
		CulturalMeaning: c.PostForm("cultural_meaning"),
		Province:        c.PostForm("province"),
		Description:     c.PostForm("description"),
	}

	if metadata.Name == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "服饰名称为必填项"))
		return
	}

	sub, err := h.submissionService.Upload(file, userID, metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, model.SuccessMessage("投稿已提交，审核通过后将公开展示", sub))
}

// ListMy 用户查看自己的投稿 GET /api/submissions
func (h *SubmissionHandler) ListMy(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	params := model.SubmissionQueryParams{
		Page:   page,
		Limit:  limit,
		Status: c.Query("status"),
		UserID: userID,
	}

	subs, total, err := h.submissionStore.ListByUser(userID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取投稿列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: subs,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

// GetMy 用户查看自己的单条投稿 GET /api/submissions/:id
func (h *SubmissionHandler) GetMy(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	sub, err := h.submissionStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	if sub.UserID != userID {
		c.JSON(http.StatusForbidden, model.ErrorResponse(40301, "无权查看该投稿"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(sub))
}

// ServeImage 提供投稿图片文件 GET /api/submissions/:id/image
func (h *SubmissionHandler) ServeImage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	sub, err := h.submissionStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	// 仅已通过/已完成的投稿图片可公开访问；
	// 其他状态仅投稿者本人和管理员可查看
	isPublished := sub.Status == "approved" || sub.Status == "completed"
	if !isPublished {
		userIDRaw, exists := c.Get("user_id")
		isAdmin, _ := c.Get("is_admin")
		if !exists || (userIDRaw.(int64) != sub.UserID && !isAdmin.(bool)) {
			c.JSON(http.StatusForbidden, model.ErrorResponse(40301, "无权查看该投稿图片"))
			return
		}
	}

	if _, err := os.Stat(sub.ImageFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "图片文件不存在"))
		return
	}

	c.Header("Content-Type", sub.ImageMimeType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.File(sub.ImageFilePath)
}

// -------- 公开端 --------

// ListPublic 公开"用户贡献"区域 GET /api/public-submissions
func (h *SubmissionHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	subs, total, err := h.submissionStore.ListPublicApproved(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取用户贡献列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: subs,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

// Retract 用户撤回自己的投稿 PUT /api/submissions/:id/retract
func (h *SubmissionHandler) Retract(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	sub, err := h.submissionStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	if sub.UserID != userID {
		c.JSON(http.StatusForbidden, model.ErrorResponse(40301, "无权操作该投稿"))
		return
	}

	if sub.Status == "retracted" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "该投稿已被撤回"))
		return
	}

	// 如果已自动创建了服饰，一并删除
	if sub.CreatedGarmentID != nil {
		_ = h.submissionStore.DeleteGarment(*sub.CreatedGarmentID)
	}

	if err := h.submissionStore.Retract(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "撤回操作失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("投稿已撤回", nil))
}

// -------- 管理端 --------

// AdminList 管理员投稿列表 GET /api/admin/submissions
func (h *SubmissionHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	params := model.SubmissionQueryParams{
		Page:   page,
		Limit:  limit,
		Status: c.Query("status"),
	}

	subs, total, err := h.submissionStore.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取投稿列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: subs,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

// AdminGet 管理员查看单条投稿详情 GET /api/admin/submissions/:id
func (h *SubmissionHandler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	sub, err := h.submissionStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(sub))
}

// AdminApprove 管理员审核通过 PUT /api/admin/submissions/:id/approve
func (h *SubmissionHandler) AdminApprove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.submissionStore.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	if err := h.submissionStore.UpdateStatus(id, "approved", "", ""); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "审核操作失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("审核已通过", nil))
}

// AdminDelete 管理员删除投稿 DELETE /api/admin/submissions/:id
func (h *SubmissionHandler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.submissionStore.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	if err := h.submissionService.DeleteSubmission(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "删除投稿失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("投稿已删除", nil))
}

// AdminReject 管理员驳回投稿 PUT /api/admin/submissions/:id/reject
func (h *SubmissionHandler) AdminReject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.submissionStore.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	var body struct {
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求体为空"))
		return
	}

	if err := h.submissionStore.UpdateStatus(id, "rejected", body.RejectReason, ""); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "驳回操作失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("投稿已驳回", nil))
}

// AdminComplete 标记投稿完成，关联 garment 和 model
// PUT /api/admin/submissions/:id/complete
// 如果未提供 garment_id，则自动从投稿元数据创建服饰记录
func (h *SubmissionHandler) AdminComplete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	sub, err := h.submissionStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "投稿不存在"))
		return
	}

	if sub.Status != "approved" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "只有已审核通过的投稿才能标记完成"))
		return
	}

	var body struct {
		GarmentID *int64 `json:"garment_id"`
		ModelID   *int64 `json:"model_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求体为空"))
		return
	}

	if body.GarmentID == nil && body.ModelID == nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "至少需要提供 garment_id 或 model_id"))
		return
	}

	// 验证 model_id
	if body.ModelID != nil {
		if _, err := h.modelStore.GetByID(*body.ModelID); err != nil {
			c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "关联的模型不存在"))
			return
		}
	}

	garmentID := body.GarmentID

	// 如果未提供 garment_id，自动从投稿元数据创建服饰记录
	if garmentID == nil {
		// 用投稿图片的访问路径作为 image_url
		imageURL := ""
		if sub.ImageFilePath != "" {
			imageURL = "/api/submissions/" + strconv.FormatInt(id, 10) + "/image"
		}

		newGarment := &model.Garment{
			Name:            sub.Name,
			Dynasty:         sub.Dynasty,
			Ethnic:          sub.Ethnic,
			Craft:           sub.Craft,
			Material:        sub.Material,
			Pattern:         sub.Pattern,
			CulturalMeaning: sub.CulturalMeaning,
			Province:        sub.Province,
			Description:     sub.Description,
			ImageURL:        imageURL,
			ModelID:         body.ModelID,
		}

		createdID, err := h.garmentStore.Create(newGarment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "自动创建服饰记录失败"))
			return
		}
		garmentID = &createdID
	} else {
		// 验证手动提供的 garment_id
		if _, err := h.garmentStore.GetByID(*garmentID); err != nil {
			c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "关联的服饰不存在"))
			return
		}
	}

	if err := h.submissionStore.Complete(id, garmentID, body.ModelID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "标记完成失败"))
		return
	}

	updated, _ := h.submissionStore.GetByID(id)
	c.JSON(http.StatusOK, model.SuccessMessage("投稿已标记为完成，服饰已自动创建", updated))
}

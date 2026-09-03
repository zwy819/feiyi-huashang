package handler

import (
	"net/http"
	"strconv"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

// CommentHandler 评论相关 handler
type CommentHandler struct {
	commentStore  *model.CommentStore
	garmentStore  *model.GarmentStore
}

// NewCommentHandler 创建评论 handler
func NewCommentHandler(commentStore *model.CommentStore, garmentStore *model.GarmentStore) *CommentHandler {
	return &CommentHandler{
		commentStore: commentStore,
		garmentStore: garmentStore,
	}
}

// -------- 用户端 / 公开端 --------

// List 服饰评论列表 GET /api/comments?garment_id=1
// 公开接口：所有人可见 approved；携带有效 Token 时额外可见自己的 pending / rejected
func (h *CommentHandler) List(c *gin.Context) {
	garmentID, err := strconv.ParseInt(c.Query("garment_id"), 10, 64)
	if err != nil || garmentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数 garment_id 不合法"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// 可选登录：JWT 中间件（required=false）已尝试解析 Token
	viewerID := int64(0)
	if userIDRaw, exists := c.Get("user_id"); exists {
		viewerID = userIDRaw.(int64)
	}

	coms, total, err := h.commentStore.ListByGarment(garmentID, viewerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取评论列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: coms,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

// Create 发表评论 POST /api/comments（JWT）
func (h *CommentHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var body struct {
		GarmentID int64  `json:"garment_id"`
		Content   string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	body.Content = strings.TrimSpace(body.Content)
	if body.GarmentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数 garment_id 不合法"))
		return
	}
	if body.Content == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "评论内容不能为空"))
		return
	}
	if len([]rune(body.Content)) > 500 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "评论内容不能超过 500 字"))
		return
	}

	if _, err := h.garmentStore.GetByID(body.GarmentID); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "服饰不存在"))
		return
	}

	com := &model.Comment{
		GarmentID: body.GarmentID,
		UserID:    userID,
		Content:   body.Content,
	}
	id, err := h.commentStore.Create(com)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "发表评论失败"))
		return
	}

	created, _ := h.commentStore.GetByID(id, userID)
	c.JSON(http.StatusCreated, model.SuccessMessage("评论已提交，审核通过后将公开展示", created))
}

// Update 编辑自己的评论 PUT /api/comments/:id（JWT）
// 编辑后评论重置为 pending 重新审核
func (h *CommentHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	body.Content = strings.TrimSpace(body.Content)
	if body.Content == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "评论内容不能为空"))
		return
	}
	if len([]rune(body.Content)) > 500 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "评论内容不能超过 500 字"))
		return
	}

	com, err := h.commentStore.GetByID(id, 0)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	if com.UserID != userID {
		c.JSON(http.StatusForbidden, model.ErrorResponse(40301, "无权操作该评论"))
		return
	}

	if err := h.commentStore.UpdateContent(id, body.Content); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "编辑评论失败"))
		return
	}

	updated, _ := h.commentStore.GetByID(id, userID)
	c.JSON(http.StatusOK, model.SuccessMessage("评论已更新，将重新进入审核", updated))
}

// Delete 删除自己的评论 DELETE /api/comments/:id（JWT）
func (h *CommentHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	com, err := h.commentStore.GetByID(id, 0)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	if com.UserID != userID {
		c.JSON(http.StatusForbidden, model.ErrorResponse(40301, "无权操作该评论"))
		return
	}

	if err := h.commentStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "删除评论失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("评论已删除", nil))
}

// Like 点赞评论 POST /api/comments/:id/like（JWT）
func (h *CommentHandler) Like(c *gin.Context) {
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	com, err := h.commentStore.GetByID(id, 0)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	// 仅审核通过的评论可点赞
	if com.Status != "approved" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "该评论尚未通过审核，无法点赞"))
		return
	}

	liked, err := h.commentStore.HasLiked(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "点赞操作失败"))
		return
	}
	if liked {
		c.JSON(http.StatusConflict, model.ErrorResponse(40901, "已点赞过该评论"))
		return
	}

	if err := h.commentStore.AddLike(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "点赞操作失败"))
		return
	}

	updated, _ := h.commentStore.GetByID(id, userID)
	c.JSON(http.StatusOK, model.SuccessResponse(updated))
}

// Unlike 取消点赞 DELETE /api/comments/:id/like（JWT）
func (h *CommentHandler) Unlike(c *gin.Context) {
	userID := c.GetInt64("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.commentStore.GetByID(id, 0); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	removed, err := h.commentStore.RemoveLike(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "取消点赞失败"))
		return
	}
	if !removed {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "尚未点赞该评论"))
		return
	}

	updated, _ := h.commentStore.GetByID(id, userID)
	c.JSON(http.StatusOK, model.SuccessResponse(updated))
}

// -------- 管理端（审核） --------

// AdminList 管理员评论列表 GET /api/admin/comments?status=pending
func (h *CommentHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	params := model.CommentQueryParams{
		Page:   page,
		Limit:  limit,
		Status: c.Query("status"),
	}

	coms, total, err := h.commentStore.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "获取评论列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(model.PaginatedData{
		Items: coms,
		Total: total,
		Page:  page,
		Limit: limit,
	}))
}

// AdminApprove 审核通过评论 PUT /api/admin/comments/:id/approve
func (h *CommentHandler) AdminApprove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	com, err := h.commentStore.GetByID(id, 0)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	if com.Status == "approved" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "该评论已通过审核"))
		return
	}

	if err := h.commentStore.UpdateStatus(id, "approved", ""); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "审核操作失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("评论审核已通过", nil))
}

// AdminReject 驳回评论 PUT /api/admin/comments/:id/reject
func (h *CommentHandler) AdminReject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.commentStore.GetByID(id, 0); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	var body struct {
		RejectReason string `json:"reject_reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求体为空"))
		return
	}

	if err := h.commentStore.UpdateStatus(id, "rejected", body.RejectReason); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "驳回操作失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("评论已驳回", nil))
}

// AdminDelete 管理员删除评论 DELETE /api/admin/comments/:id
func (h *CommentHandler) AdminDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "参数值不合法"))
		return
	}

	if _, err := h.commentStore.GetByID(id, 0); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "评论不存在"))
		return
	}

	if err := h.commentStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "删除评论失败"))
		return
	}

	c.JSON(http.StatusOK, model.SuccessMessage("评论已删除", nil))
}

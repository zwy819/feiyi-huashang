package handler

import (
	"net/http"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	result, err := h.authService.Register(input)
	if err != nil {
		msg := err.Error()
		code := 40002
		if strings.Contains(msg, "已存在") || strings.Contains(msg, "已被注册") {
			code = 40901
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse(code, msg))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "参数缺失或格式错误"))
		return
	}

	result, err := h.authService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(40103, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse(result))
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse(40401, "用户不存在"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(user))
}

package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"nonlegacy-garment-3d/server/internal/ai"
	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	service *ai.Service
}

func NewAIHandler(service *ai.Service) *AIHandler {
	return &AIHandler{service: service}
}

func (h *AIHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, model.Response{Code: 0, Data: h.service.Status(c.Request.Context())})
}

type aiFeedbackRequest struct {
	MessageID string `json:"message_id"`
	Rating    int    `json:"rating"`
	Reason    string `json:"reason"`
}

func (h *AIHandler) Feedback(c *gin.Context) {
	var req aiFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求格式不正确"))
		return
	}
	if err := h.service.SaveFeedback(c.Request.Context(), req.MessageID, req.Rating, req.Reason); err != nil {
		if chatErr, ok := err.(*ai.ChatError); ok {
			c.JSON(http.StatusBadRequest, model.ErrorResponse(chatErr.Code, chatErr.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "反馈保存失败"))
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Data: gin.H{"saved": true}})
}

func (h *AIHandler) Reindex(c *gin.Context) {
	stats, err := h.service.Reindex(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "知识库重建失败"))
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Data: stats})
}

// Chat streams server-sent events. Header is written only after the request has
// passed the availability check; all later failures are delivered as an SSE
// `error` event so the browser can keep the partial answer and offer a retry.
func (h *AIHandler) Chat(c *gin.Context) {
	if !h.service.Available() {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse(50301, "AI 导览暂未配置，请先设置 AI_API_KEY"))
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024)
	var req ai.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40001, "请求格式不正确"))
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	clientKey := c.ClientIP() + "|" + c.Request.UserAgent()
	emit := func(event string, data any) error {
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, raw); err != nil {
			return err
		}
		c.Writer.Flush()
		return nil
	}

	if err := h.service.Chat(c.Request.Context(), req, clientKey, emit); err != nil {
		log.Printf("[AI][HTTP] chat ended: %v", err)
	}
}

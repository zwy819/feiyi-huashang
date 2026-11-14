package handler

import (
	"log"
	"net/http"
	"strconv"

	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/tts"

	"github.com/gin-gonic/gin"
)

type TTSHandler struct {
	service *tts.Service
}

func NewTTSHandler(service *tts.Service) *TTSHandler {
	return &TTSHandler{service: service}
}

// Garment 返回一件服饰简介的自然语音 WAV。
func (h *TTSHandler) Garment(c *gin.Context) {
	if h.service == nil || !h.service.Available() {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse(50302, "语音讲解暂未配置"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "服饰 ID 不正确"))
		return
	}
	audio, err := h.service.GarmentAudio(c.Request.Context(), id)
	if err != nil {
		log.Printf("[TTS] 服饰 %d 合成失败: %v", id, err)
		c.JSON(http.StatusBadGateway, model.ErrorResponse(50201, "语音生成失败"))
		return
	}
	writeWAV(c, audio)
}

// Annotation 返回一个 3D 热点的讲解语音；?step=2 会播报“第 2 站”。
func (h *TTSHandler) Annotation(c *gin.Context) {
	if h.service == nil || !h.service.Available() {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse(50302, "语音讲解暂未配置"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse(40002, "热点 ID 不正确"))
		return
	}
	step, _ := strconv.Atoi(c.DefaultQuery("step", "0"))
	if step < 0 || step > 99 {
		step = 0
	}
	audio, err := h.service.AnnotationAudio(c.Request.Context(), id, step)
	if err != nil {
		log.Printf("[TTS] 热点 %d 合成失败: %v", id, err)
		c.JSON(http.StatusBadGateway, model.ErrorResponse(50201, "语音生成失败"))
		return
	}
	writeWAV(c, audio)
}

func writeWAV(c *gin.Context, audio []byte) {
	c.Header("Content-Type", "audio/wav")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Content-Length", strconv.Itoa(len(audio)))
	c.Data(http.StatusOK, "audio/wav", audio)
}

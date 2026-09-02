package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"nonlegacy-garment-3d/server/internal/model"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Recovery] panic recovered: %v\n%s", r, debug.Stack())
				c.JSON(http.StatusInternalServerError, model.ErrorResponse(50001, "服务器内部错误"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

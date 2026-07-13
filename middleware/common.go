package middleware

import (
	"flower-lottery-backend/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http_request", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path), zap.Int("status", c.Writer.Status()), zap.Duration("duration", time.Since(start)))
	}
}
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic", zap.Any("error", recovered))
		response.Error(c, 500, 10000, "系统繁忙，请稍后重试")
	})
}

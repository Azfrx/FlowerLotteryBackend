package controller

import (
	"flower-lottery-backend/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthController struct{ DB *gorm.DB }

func (h HealthController) Check(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		response.Error(c, 503, 10001, "database unavailable")
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

type ModuleController struct{}

func (ModuleController) NotImplemented(c *gin.Context) {
	response.Error(c, 501, 10002, "module endpoint is scaffolded and pending implementation")
}

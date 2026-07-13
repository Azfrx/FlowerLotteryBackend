package controller

import (
	"flower-lottery-backend/common"
	"flower-lottery-backend/response"
	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err error) {
	if app, ok := common.AsApp(err); ok {
		response.Error(c, app.HTTPStatus, app.Code, app.Message)
		return
	}
	response.Error(c, 500, 10000, "系统繁忙，请稍后重试")
}

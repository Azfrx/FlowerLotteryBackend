package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Success(c *gin.Context, data any) { c.JSON(200, Body{Code: 0, Msg: "", Data: data}) }
func Error(c *gin.Context, httpStatus, code int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Body{Code: code, Msg: msg, Data: gin.H{}})
}

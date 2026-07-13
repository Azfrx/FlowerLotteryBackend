package middleware

import (
	"flower-lottery-backend/common"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/response"
	"github.com/gin-gonic/gin"
	"strings"
)

const UserIDKey = "user_id"

func UserAuth(tokens *tokenjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeAuthError(c, common.ErrUnauthorized)
			return
		}
		claims, err := tokens.Parse(parts[1], "access")
		if err != nil || claims.SubjectType != "user" {
			writeAuthError(c, common.ErrUnauthorized)
			return
		}
		c.Set(UserIDKey, claims.SubjectID)
		c.Next()
	}
}
func AdminAuth(tokens *tokenjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 {
			writeAuthError(c, common.ErrUnauthorized)
			return
		}
		claims, err := tokens.Parse(parts[1], "access")
		if err != nil || claims.SubjectType != "admin" {
			writeAuthError(c, common.ErrUnauthorized)
			return
		}
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint64 {
	value, ok := c.Get(UserIDKey)
	if !ok {
		return 0
	}
	id, _ := value.(uint64)
	return id
}
func writeAuthError(c *gin.Context, err error) {
	if app, ok := common.AsApp(err); ok {
		response.Error(c, app.HTTPStatus, app.Code, app.Message)
		return
	}
	response.Error(c, 401, 11003, "登录状态已失效")
}

package controller

import (
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/request"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"github.com/gin-gonic/gin"
)

type AuthController struct{ service *service.AuthService }

func NewAuthController(s *service.AuthService) *AuthController { return &AuthController{service: s} }
func (cn *AuthController) Register(c *gin.Context) {
	var req request.Register
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	user, pair, err := cn.service.Register(req.UserID, req.Nickname, req.AvatarURL, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"user": user, "tokens": pair})
}
func (cn *AuthController) Login(c *gin.Context) {
	var req request.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	user, pair, err := cn.service.Login(req.UserID, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"user": user, "tokens": pair})
}
func (cn *AuthController) Refresh(c *gin.Context) {
	var req request.Refresh
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	pair, err := cn.service.Refresh(req.RefreshToken)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, pair)
}
func (cn *AuthController) Logout(c *gin.Context) {
	var req request.Refresh
	_ = c.ShouldBindJSON(&req)
	if err := cn.service.Logout(req.RefreshToken); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{})
}
func (cn *AuthController) Me(c *gin.Context) {
	user, err := cn.service.Me(middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, user)
}
func (cn *AuthController) UpdateProfile(c *gin.Context) {
	var req request.UpdateProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	user, err := cn.service.UpdateProfile(middleware.CurrentUserID(c), req.Nickname, req.AvatarURL)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, user)
}
func (cn *AuthController) ChangePassword(c *gin.Context) {
	var req request.ChangePassword
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	if err := cn.service.ChangePassword(middleware.CurrentUserID(c), req.CurrentPassword, req.NewPassword); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{})
}

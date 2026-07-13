package controller

import (
	"errors"
	"flower-lottery-backend/common"
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/request"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AuthController struct {
	service *service.AuthService
	avatars *service.AvatarService
}

func NewAuthController(s *service.AuthService, avatars *service.AvatarService) *AuthController {
	return &AuthController{service: s, avatars: avatars}
}

func (cn *AuthController) Register(c *gin.Context) {
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		cn.registerMultipart(c)
		return
	}
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

func (cn *AuthController) registerMultipart(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cn.avatars.MaxUploadBytes()+1<<20)
	if err := c.Request.ParseMultipartForm(cn.avatars.MaxUploadBytes()); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(c, common.ErrAvatarTooLarge)
			return
		}
		response.Error(c, http.StatusBadRequest, 10001, "表单数据无效")
		return
	}
	var avatarURL string
	file, err := c.FormFile("avatar")
	if err == nil {
		avatarURL, err = cn.avatars.Save(file)
		if err != nil {
			writeError(c, err)
			return
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		writeError(c, common.ErrAvatarDecode)
		return
	}
	user, pair, err := cn.service.Register(c.PostForm("user_id"), c.PostForm("nickname"), avatarURL, c.PostForm("password"))
	if err != nil {
		cn.avatars.DeleteLocal(avatarURL)
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"user": user, "tokens": pair})
}

func (cn *AuthController) UploadAvatar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cn.avatars.MaxUploadBytes()+1<<20)
	file, err := c.FormFile("avatar")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(c, common.ErrAvatarTooLarge)
			return
		}
		writeError(c, common.ErrAvatarRequired)
		return
	}
	avatarURL, err := cn.avatars.Save(file)
	if err != nil {
		writeError(c, err)
		return
	}
	current, err := cn.service.Me(middleware.CurrentUserID(c))
	if err != nil {
		cn.avatars.DeleteLocal(avatarURL)
		writeError(c, err)
		return
	}
	user, err := cn.service.UpdateAvatar(current.ID, avatarURL)
	if err != nil {
		cn.avatars.DeleteLocal(avatarURL)
		writeError(c, err)
		return
	}
	cn.avatars.DeleteLocal(current.AvatarURL)
	response.Success(c, user)
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
	if c.ShouldBindJSON(&req) != nil {
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
	if c.ShouldBindJSON(&req) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	user, err := cn.service.UpdateProfile(middleware.CurrentUserID(c), req.Nickname)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, user)
}
func (cn *AuthController) ChangePassword(c *gin.Context) {
	var req request.ChangePassword
	if c.ShouldBindJSON(&req) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	if err := cn.service.ChangePassword(middleware.CurrentUserID(c), req.CurrentPassword, req.NewPassword); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{})
}

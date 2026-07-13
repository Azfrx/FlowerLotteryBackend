package controller

import (
	"flower-lottery-backend/config"
	"flower-lottery-backend/model"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/response"
	"flower-lottery-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type AdminController struct {
	db     *gorm.DB
	tokens *tokenjwt.Manager
}

func NewAdminController(db *gorm.DB, c config.JWT) *AdminController {
	return &AdminController{db: db, tokens: tokenjwt.New(c)}
}
func (a *AdminController) Login(c *gin.Context) {
	var q struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if c.ShouldBindJSON(&q) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	var u model.AdminUser
	if a.db.Where("username=? AND deleted_at IS NULL", q.Username).First(&u).Error != nil || !utils.CheckPassword(u.PasswordHash, q.Password) || u.Status != 1 {
		response.Error(c, 401, 11001, "账号或密码错误")
		return
	}
	pair, _, e := a.tokens.Issue(u.ID, "admin")
	if e != nil {
		writeError(c, e)
		return
	}
	now := time.Now()
	a.db.Model(&u).Update("last_login_at", now)
	response.Success(c, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken, "admin": u})
}
func (a *AdminController) Dashboard(c *gin.Context) {
	var users, orders, rewards int64
	var petals int64
	a.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&users)
	a.db.Model(&model.LotteryOrder{}).Count(&orders)
	a.db.Model(&model.UserReward{}).Count(&rewards)
	a.db.Model(&model.LotteryOrder{}).Select("COALESCE(SUM(petal_cost),0)").Scan(&petals)
	response.Success(c, gin.H{"users": users, "orders": orders, "rewards": rewards, "petals": petals})
}
func (a *AdminController) Users(c *gin.Context) {
	var list []struct {
		ID           uint64
		UserNo       string
		Nickname     string
		Status       uint8
		CoinBalance  int64
		PetalBalance int64
	}
	a.db.Table("users u").Select("u.id,u.user_no,u.nickname,u.status,w.coin_balance,w.petal_balance").Joins("LEFT JOIN user_wallets w ON w.user_id=u.id").Where("u.deleted_at IS NULL").Order("u.id DESC").Limit(200).Scan(&list)
	response.Success(c, list)
}
func (a *AdminController) Assets(c *gin.Context) {
	var v []model.AssetTransaction
	a.db.Order("id DESC").Limit(500).Find(&v)
	response.Success(c, v)
}
func (a *AdminController) Orders(c *gin.Context) {
	var v []model.LotteryOrder
	a.db.Order("id DESC").Limit(500).Find(&v)
	response.Success(c, v)
}
func (a *AdminController) Rewards(c *gin.Context) {
	var v []model.UserReward
	a.db.Preload("RewardItem").Order("id DESC").Limit(500).Find(&v)
	response.Success(c, v)
}
func (a *AdminController) Rounds(c *gin.Context) {
	var v []model.UserActivityRound
	a.db.Order("id DESC").Limit(500).Find(&v)
	response.Success(c, v)
}
func (a *AdminController) Leaderboard(c *gin.Context) {
	var v []model.LeaderboardEntry
	a.db.Preload("User").Order("score DESC,reached_at ASC").Limit(200).Find(&v)
	response.Success(c, v)
}

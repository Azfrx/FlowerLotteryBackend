package controller

import (
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/request"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"fmt"
	"github.com/gin-gonic/gin"
)

type GameController struct {
	s             *service.GameService
	announcements *service.RewardAnnouncementHub
}

func NewGameController(s *service.GameService, announcements *service.RewardAnnouncementHub) *GameController {
	return &GameController{s: s, announcements: announcements}
}
func (x *GameController) Home(c *gin.Context) {
	v, e := x.s.Home(middleware.CurrentUserID(c))
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) ActivityContent(c *gin.Context) {
	v, e := x.s.ActivityContent()
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) Catalog(c *gin.Context) {
	v, e := x.s.Catalog()
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) Draw(c *gin.Context) {
	var q request.Lottery
	if e := c.ShouldBindJSON(&q); e != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	v, e := x.s.Draw(middleware.CurrentUserID(c), q.PoolCode, q.DrawCount, q.RequestID)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) Orders(c *gin.Context) {
	var p request.Page
	_ = c.ShouldBindQuery(&p)
	p.Normalize()
	v, n, e := x.s.Orders(middleware.CurrentUserID(c), p.Page, p.PageSize)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, gin.H{"list": v, "page": p.Page, "page_size": p.PageSize, "total": n})
}
func (x *GameController) LotteryHistory(c *gin.Context) {
	v, e := x.s.LotteryHistory(middleware.CurrentUserID(c))
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) LotteryRewards(c *gin.Context) {
	v, e := x.s.LotteryRewardInventory(middleware.CurrentUserID(c))
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) ChestHistory(c *gin.Context) {
	v, e := x.s.ChestHistory(middleware.CurrentUserID(c))
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) Rewards(c *gin.Context) {
	var p request.Page
	_ = c.ShouldBindQuery(&p)
	p.Normalize()
	v, n, e := x.s.Rewards(middleware.CurrentUserID(c), p.Page, p.PageSize)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, gin.H{"list": v, "page": p.Page, "page_size": p.PageSize, "total": n})
}
func (x *GameController) Leaderboard(c *gin.Context) {
	top, self, rank, e := x.s.Leaderboard(middleware.CurrentUserID(c))
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, gin.H{"top": top, "self": self, "self_rank": rank})
}

func (x *GameController) OpenChest(c *gin.Context) {
	var q request.Action
	if c.ShouldBindJSON(&q) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	var id uint64
	fmt.Sscan(c.Param("id"), &id)
	v, e := x.s.OpenChest(middleware.CurrentUserID(c), id, q.RequestID)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) SelectChest(c *gin.Context) {
	var q request.SelectChest
	if c.ShouldBindJSON(&q) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	var id uint64
	fmt.Sscan(c.Param("id"), &id)
	v, e := x.s.SelectChest(middleware.CurrentUserID(c), id, q.ItemCode, q.RequestID)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) ClaimStage(c *gin.Context) {
	var q request.Action
	if c.ShouldBindJSON(&q) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	var id uint64
	fmt.Sscan(c.Param("id"), &id)
	v, e := x.s.ClaimStage(middleware.CurrentUserID(c), id, q.RequestID)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) NextRound(c *gin.Context) {
	v, e := x.s.NextRound(middleware.CurrentUserID(c))
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}

func (x *GameController) Preview(c *gin.Context) {
	var q request.Action
	if c.ShouldBindJSON(&q) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	v, e := x.s.Preview180(middleware.CurrentUserID(c), q.RequestID)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) ConfirmPreview(c *gin.Context) {
	var id uint64
	fmt.Sscan(c.Param("id"), &id)
	v, e := x.s.ConfirmPreview(middleware.CurrentUserID(c), id)
	if e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, v)
}
func (x *GameController) CancelPreview(c *gin.Context) {
	var id uint64
	fmt.Sscan(c.Param("id"), &id)
	if e := x.s.CancelPreview(middleware.CurrentUserID(c), id); e != nil {
		writeError(c, e)
		return
	}
	response.Success(c, gin.H{})
}

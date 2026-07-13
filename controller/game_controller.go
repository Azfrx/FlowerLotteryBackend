package controller

import (
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/request"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"fmt"
	"github.com/gin-gonic/gin"
)

type GameController struct{ s *service.GameService }

func NewGameController(s *service.GameService) *GameController { return &GameController{s: s} }
func (x *GameController) Home(c *gin.Context) {
	v, e := x.s.Home(middleware.CurrentUserID(c))
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
	v, e := x.s.OpenChest(middleware.CurrentUserID(c), id)
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
	v, e := x.s.SelectChest(middleware.CurrentUserID(c), id, q.CandidateID)
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

package controller

import (
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/request"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"github.com/gin-gonic/gin"
)

type WalletController struct{ service *service.WalletService }

func NewWalletController(s *service.WalletService) *WalletController {
	return &WalletController{service: s}
}
func (cn *WalletController) Get(c *gin.Context) {
	wallet, err := cn.service.Get(middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, wallet)
}
func (cn *WalletController) Options(c *gin.Context) {
	list, err := cn.service.Options()
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (cn *WalletController) Exchange(c *gin.Context) {
	var req request.Exchange
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	order, wallet, err := cn.service.Exchange(
		middleware.CurrentUserID(c), req.OptionID,
		req.ExpectedPetalAmount, req.ExpectedCoinCost, req.RequestID,
	)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"order": order, "wallet": wallet})
}
func (cn *WalletController) PetalGiftPackStatus(c *gin.Context) {
	purchased, err := cn.service.PetalGiftPackStatus(middleware.CurrentUserID(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"purchased": purchased})
}
func (cn *WalletController) PurchasePetalGiftPack(c *gin.Context) {
	var req request.Action
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	result, err := cn.service.PurchasePetalGiftPack(middleware.CurrentUserID(c), req.RequestID)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, result)
}
func (cn *WalletController) Transactions(c *gin.Context) {
	var page request.Page
	if err := c.ShouldBindQuery(&page); err != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	page.Normalize()
	list, total, err := cn.service.Transactions(middleware.CurrentUserID(c), page.Page, page.PageSize)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"list": list, "page": page.Page, "page_size": page.PageSize, "total": total})
}

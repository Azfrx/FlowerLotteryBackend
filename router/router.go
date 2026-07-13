package router

import (
	"flower-lottery-backend/config"
	"flower-lottery-backend/controller"
	"flower-lottery-backend/middleware"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/repository"
	"flower-lottery-backend/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func New(db *gorm.DB, log *zap.Logger, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestLogger(log), middleware.Recovery(log))
	tokens := tokenjwt.New(cfg.JWT)
	auth := controller.NewAuthController(service.NewAuthService(repository.NewUserRepository(db), tokens))
	wallet := controller.NewWalletController(service.NewWalletService(repository.NewWalletRepository(db)))
	game := controller.NewGameController(service.NewGameService(repository.NewGameRepository(db)))
	adminController := controller.NewAdminController(db, cfg.JWT)
	health := controller.HealthController{DB: db}
	v1 := r.Group("/api/v1")
	v1.GET("/health", health.Check)
	v1.POST("/auth/login", auth.Login)
	v1.POST("/auth/refresh", auth.Refresh)
	protected := v1.Group("")
	protected.Use(middleware.UserAuth(tokens))
	protected.POST("/auth/logout", auth.Logout)
	protected.GET("/me", auth.Me)
	protected.GET("/wallet", wallet.Get)
	protected.GET("/wallet/exchange-options", wallet.Options)
	protected.POST("/wallet/exchanges", wallet.Exchange)
	protected.GET("/wallet/transactions", wallet.Transactions)
	protected.GET("/activities/current/home", game.Home)
	protected.GET("/activities/current/rewards/catalog", game.Catalog)
	protected.POST("/lottery/orders", game.Draw)
	protected.GET("/lottery/orders", game.Orders)
	protected.GET("/flower/round", game.Home)
	protected.POST("/flower/chests/:id/open", game.OpenChest)
	protected.POST("/flower/chests/:id/select", game.SelectChest)
	protected.POST("/flower/stage-rewards/:id/claim", game.ClaimStage)
	protected.POST("/flower/round/next", game.NextRound)
	protected.GET("/rewards", game.Rewards)
	protected.GET("/leaderboard", game.Leaderboard)
	admin := v1.Group("/admin")
	admin.POST("/auth/login", adminController.Login)
	admin.Use(middleware.AdminAuth(tokens))
	admin.GET("/dashboard", adminController.Dashboard)
	admin.GET("/users", adminController.Users)
	admin.GET("/asset-transactions", adminController.Assets)
	admin.GET("/lottery-orders", adminController.Orders)
	admin.GET("/rewards", adminController.Rewards)
	admin.GET("/rounds", adminController.Rounds)
	admin.GET("/leaderboard", adminController.Leaderboard)
	return r
}

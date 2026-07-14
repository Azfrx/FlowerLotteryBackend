package controller

import (
	"flower-lottery-backend/model"
	"flower-lottery-backend/response"
	"github.com/gin-gonic/gin"
	"time"
)

type adminDashboardActivity struct {
	ID       uint64    `json:"id"`
	Name     string    `json:"name"`
	Status   uint8     `json:"status"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

type adminDailyLotteryPoint struct {
	Day     string `json:"day"`
	Orders  int64  `json:"orders"`
	Draws   int64  `json:"draws"`
	Petals  int64  `json:"petals"`
	Flowers int64  `json:"flowers"`
}

type adminDailyAssetPoint struct {
	Day      string `json:"day"`
	CoinIn   int64  `json:"coin_in"`
	CoinOut  int64  `json:"coin_out"`
	CoinNet  int64  `json:"coin_net"`
	PetalIn  int64  `json:"petal_in"`
	PetalOut int64  `json:"petal_out"`
	PetalNet int64  `json:"petal_net"`
}

type adminPoolDistribution struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Orders int64  `json:"orders"`
	Draws  int64  `json:"draws"`
	Petals int64  `json:"petals"`
}

type adminRewardDistribution struct {
	ItemCode string `json:"item_code"`
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
}

type adminFlowerDistribution struct {
	LitFlowerCount uint8 `json:"lit_flower_count"`
	Users          int64 `json:"users"`
}

type adminDashboardData struct {
	Users              int64                     `json:"users"`
	ActiveUsers        int64                     `json:"active_users"`
	Orders             int64                     `json:"orders"`
	Draws              int64                     `json:"draws"`
	Rewards            int64                     `json:"rewards"`
	Petals             int64                     `json:"petals"`
	CoinBalance        int64                     `json:"coin_balance"`
	PetalBalance       int64                     `json:"petal_balance"`
	TodayNewUsers      int64                     `json:"today_new_users"`
	TodayOrders        int64                     `json:"today_orders"`
	PendingOrders      int64                     `json:"pending_orders"`
	Activity           *adminDashboardActivity   `json:"activity"`
	DailyLottery       []adminDailyLotteryPoint  `json:"daily_lottery"`
	DailyAssets        []adminDailyAssetPoint    `json:"daily_assets"`
	PoolDistribution   []adminPoolDistribution   `json:"pool_distribution"`
	RewardDistribution []adminRewardDistribution `json:"reward_distribution"`
	FlowerDistribution []adminFlowerDistribution `json:"flower_distribution"`
}

func (a *AdminController) Dashboard(c *gin.Context) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	since := today.AddDate(0, 0, -13)
	data := adminDashboardData{
		DailyLottery:       make([]adminDailyLotteryPoint, 0, 14),
		DailyAssets:        make([]adminDailyAssetPoint, 0, 14),
		PoolDistribution:   make([]adminPoolDistribution, 0),
		RewardDistribution: make([]adminRewardDistribution, 0),
		FlowerDistribution: make([]adminFlowerDistribution, 0),
	}

	if err := a.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&data.Users).Error; err != nil {
		writeError(c, err)
		return
	}
	if err := a.db.Table("lottery_orders").Where("status=1 AND created_at>=?", since).Distinct("user_id").Count(&data.ActiveUsers).Error; err != nil {
		writeError(c, err)
		return
	}

	var orderTotals struct {
		Orders int64 `gorm:"column:orders"`
		Draws  int64 `gorm:"column:draws"`
		Petals int64 `gorm:"column:petals"`
	}
	if err := a.db.Table("lottery_orders").
		Select("COUNT(*) AS orders,COALESCE(SUM(executed_draw_count),0) AS draws,COALESCE(SUM(petal_cost-petal_refund),0) AS petals").
		Where("status=1").Scan(&orderTotals).Error; err != nil {
		writeError(c, err)
		return
	}
	data.Orders = orderTotals.Orders
	data.Draws = orderTotals.Draws
	data.Petals = orderTotals.Petals

	if err := a.db.Model(&model.UserReward{}).Where("status IN (1,2)").Count(&data.Rewards).Error; err != nil {
		writeError(c, err)
		return
	}
	var walletTotals struct {
		CoinBalance  int64 `gorm:"column:coin_balance"`
		PetalBalance int64 `gorm:"column:petal_balance"`
	}
	if err := a.db.Table("user_wallets AS w").
		Select("COALESCE(SUM(w.coin_balance),0) AS coin_balance,COALESCE(SUM(w.petal_balance),0) AS petal_balance").
		Joins("JOIN users AS u ON u.id=w.user_id AND u.deleted_at IS NULL").
		Scan(&walletTotals).Error; err != nil {
		writeError(c, err)
		return
	}
	data.CoinBalance = walletTotals.CoinBalance
	data.PetalBalance = walletTotals.PetalBalance

	if err := a.db.Model(&model.User{}).Where("deleted_at IS NULL AND created_at>=?", today).Count(&data.TodayNewUsers).Error; err != nil {
		writeError(c, err)
		return
	}
	if err := a.db.Model(&model.LotteryOrder{}).Where("status=1 AND created_at>=?", today).Count(&data.TodayOrders).Error; err != nil {
		writeError(c, err)
		return
	}
	if err := a.db.Model(&model.LotteryOrder{}).Where("status IN (0,2)").Count(&data.PendingOrders).Error; err != nil {
		writeError(c, err)
		return
	}

	var activity adminDashboardActivity
	activityQuery := a.db.Table("activities").
		Select("id,name,status,starts_at,ends_at").
		Where("deleted_at IS NULL")
	if err := activityQuery.
		Order("CASE WHEN status=2 AND starts_at<=NOW() AND ends_at>NOW() THEN 0 ELSE 1 END,id DESC").
		Limit(1).Scan(&activity).Error; err != nil {
		writeError(c, err)
		return
	}
	if activity.ID > 0 {
		data.Activity = &activity
	}

	var lotteryRows []adminDailyLotteryPoint
	if err := a.db.Table("lottery_orders").
		Select(`DATE_FORMAT(created_at,'%Y-%m-%d') AS day,COUNT(*) AS orders,
			COALESCE(SUM(executed_draw_count),0) AS draws,
			COALESCE(SUM(petal_cost-petal_refund),0) AS petals,
			COALESCE(SUM(CASE WHEN flowers_after>=flowers_before THEN flowers_after-flowers_before ELSE 0 END),0) AS flowers`).
		Where("status=1 AND created_at>=?", since).
		Group("DATE_FORMAT(created_at,'%Y-%m-%d')").
		Order("day").Scan(&lotteryRows).Error; err != nil {
		writeError(c, err)
		return
	}
	lotteryByDay := make(map[string]adminDailyLotteryPoint, len(lotteryRows))
	for _, row := range lotteryRows {
		lotteryByDay[row.Day] = row
	}

	var assetRows []adminDailyAssetPoint
	if err := a.db.Table("asset_transactions").
		Select(`DATE_FORMAT(created_at,'%Y-%m-%d') AS day,
			COALESCE(SUM(CASE WHEN asset_type='coin' AND change_amount>0 THEN change_amount ELSE 0 END),0) AS coin_in,
			COALESCE(SUM(CASE WHEN asset_type='coin' AND change_amount<0 THEN -change_amount ELSE 0 END),0) AS coin_out,
			COALESCE(SUM(CASE WHEN asset_type='coin' THEN change_amount ELSE 0 END),0) AS coin_net,
			COALESCE(SUM(CASE WHEN asset_type='petal' AND change_amount>0 THEN change_amount ELSE 0 END),0) AS petal_in,
			COALESCE(SUM(CASE WHEN asset_type='petal' AND change_amount<0 THEN -change_amount ELSE 0 END),0) AS petal_out,
			COALESCE(SUM(CASE WHEN asset_type='petal' THEN change_amount ELSE 0 END),0) AS petal_net`).
		Where("created_at>=?", since).
		Group("DATE_FORMAT(created_at,'%Y-%m-%d')").
		Order("day").Scan(&assetRows).Error; err != nil {
		writeError(c, err)
		return
	}
	assetsByDay := make(map[string]adminDailyAssetPoint, len(assetRows))
	for _, row := range assetRows {
		assetsByDay[row.Day] = row
	}

	for day := since; !day.After(today); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		lotteryPoint := lotteryByDay[key]
		lotteryPoint.Day = key
		data.DailyLottery = append(data.DailyLottery, lotteryPoint)
		assetPoint := assetsByDay[key]
		assetPoint.Day = key
		data.DailyAssets = append(data.DailyAssets, assetPoint)
	}

	if err := a.db.Table("lottery_orders AS o").
		Select("p.code AS code,p.name AS name,COUNT(*) AS orders,COALESCE(SUM(o.executed_draw_count),0) AS draws,COALESCE(SUM(o.petal_cost-o.petal_refund),0) AS petals").
		Joins("JOIN prize_pools AS p ON p.id=o.prize_pool_id").
		Where("o.status=1").
		Group("p.id,p.code,p.name").
		Order("orders DESC").Scan(&data.PoolDistribution).Error; err != nil {
		writeError(c, err)
		return
	}
	if err := a.db.Table("lottery_draws AS d").
		Select("i.item_code AS item_code,i.name AS name,COALESCE(SUM(d.reward_quantity),0) AS quantity").
		Joins("JOIN lottery_orders AS o ON o.id=d.lottery_order_id AND o.status=1").
		Joins("JOIN reward_items AS i ON i.id=d.reward_item_id").
		Group("i.id,i.item_code,i.name").
		Order("quantity DESC").
		Limit(8).Scan(&data.RewardDistribution).Error; err != nil {
		writeError(c, err)
		return
	}
	if err := a.db.Table("user_activity_rounds AS r").
		Select("r.lit_flower_count AS lit_flower_count,COUNT(*) AS users").
		Where(`r.id=(SELECT r2.id FROM user_activity_rounds AS r2
			WHERE r2.user_id=r.user_id AND r2.activity_id=r.activity_id
			ORDER BY r2.round_no DESC LIMIT 1)`).
		Group("r.lit_flower_count").
		Order("r.lit_flower_count").Scan(&data.FlowerDistribution).Error; err != nil {
		writeError(c, err)
		return
	}

	response.Success(c, data)
}

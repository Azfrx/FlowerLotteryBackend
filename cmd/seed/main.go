package main

import (
	"flower-lottery-backend/initialize"
	"flower-lottery-backend/model"
	"flower-lottery-backend/utils"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

func main() {
	cfg, e := initialize.Config()
	if e != nil {
		log.Fatal(e)
	}
	db, e := initialize.Database(cfg.Database)
	if e != nil {
		log.Fatal(e)
	}
	hash, _ := utils.HashPassword("123456")
	e = db.Transaction(func(tx *gorm.DB) error {
		a := model.Activity{Code: "flower-wish", Name: "花愿奇遇", Status: 2, StartsAt: time.Now().Add(-24 * time.Hour), EndsAt: time.Now().AddDate(0, 1, 0), LeaderboardFreezesAt: time.Now().AddDate(0, 1, 0), Timezone: "Asia/Shanghai"}
		if e := tx.Where("code=?", a.Code).FirstOrCreate(&a).Error; e != nil {
			return e
		}
		u := model.User{UserNo: "demo", Nickname: "体验用户", PasswordHash: hash, Status: 1}
		if e := tx.Where("user_no=?", u.UserNo).FirstOrCreate(&u).Error; e != nil {
			return e
		}
		admin := model.AdminUser{Username: "admin", DisplayName: "系统管理员", PasswordHash: hash, Status: 1}
		if e := tx.Where("username=?", admin.Username).FirstOrCreate(&admin).Error; e != nil {
			return e
		}
		w := model.UserWallet{UserID: u.ID, CoinBalance: 10000000}
		if e := tx.Where("user_id=?", u.ID).FirstOrCreate(&w).Error; e != nil {
			return e
		}
		opts := [][2]uint64{{5, 300}, {50, 3000}, {150, 9000}, {100, 6000}, {1000, 60000}, {3000, 180000}}
		for i, v := range opts {
			o := model.ExchangeOption{ActivityID: a.ID, PetalAmount: v[0], CoinCost: v[1], SortNo: i + 1, Status: 1}
			if e := tx.Where("activity_id=? AND petal_amount=?", a.ID, v[0]).FirstOrCreate(&o).Error; e != nil {
				return e
			}
		}
		pools := []model.PrizePool{{ActivityID: a.ID, Code: "day", Name: "白昼许愿", PetalCostPerDraw: 5, CoinValuePerDraw: 300, SupportedDrawCounts: []byte(`[1,10,30]`), Status: 1, SortNo: 1}, {ActivityID: a.ID, Code: "night", Name: "星夜许愿", PetalCostPerDraw: 100, CoinValuePerDraw: 6000, SupportedDrawCounts: []byte(`[1,10,30]`), Status: 1, SortNo: 2}}
		items := []model.RewardItem{{ItemCode: "1001", Name: "金币", ItemType: "coin", Status: 1}, {ItemCode: "1002", Name: "花瓣", ItemType: "petal", Status: 1}, {ItemCode: "1207244", Name: "烟花之恋戒指", ItemType: "item", Status: 1}, {ItemCode: "1203743", Name: "闪闪心蝶戒指", ItemType: "item", Status: 1}, {ItemCode: "1205251", Name: "月亮游记戒指", ItemType: "item", Status: 1}, {ItemCode: "1205470", Name: "爱意翩跹戒指", ItemType: "item", Status: 1}, {ItemCode: "1207751", Name: "真爱无敌戒指", ItemType: "item", Status: 1}}
		for i := range items {
			if e := tx.Where("item_code=?", items[i].ItemCode).FirstOrCreate(&items[i]).Error; e != nil {
				return e
			}
		}
		for i := range pools {
			if e := tx.Where("activity_id=? AND code=?", a.ID, pools[i].Code).FirstOrCreate(&pools[i]).Error; e != nil {
				return e
			}
			v := model.PrizePoolVersion{PrizePoolID: pools[i].ID, VersionNo: 1, Status: 1, TotalWeight: 1000000}
			now := time.Now().Add(-time.Hour)
			v.EffectiveAt = &now
			if e := tx.Where("prize_pool_id=? AND version_no=1", pools[i].ID).FirstOrCreate(&v).Error; e != nil {
				return e
			}
			weights := [][]uint64{{350000, 250000, 200000, 100000, 70000, 25000, 5000}, {250000, 200000, 200000, 150000, 100000, 70000, 30000}}[i]
			qty := [][]uint64{{100, 300, 1, 1, 1, 1, 1}, {1000, 10, 1, 1, 1, 1, 1}}[i]
			for j := range weights {
				pr := model.PrizePoolReward{VersionID: v.ID, RewardItemID: items[j].ID, Quantity: qty[j], Weight: weights[j], SortNo: j + 1}
				if e := tx.Where("version_id=? AND reward_item_id=? AND quantity=?", v.ID, items[j].ID, qty[j]).FirstOrCreate(&pr).Error; e != nil {
					return e
				}
			}
		}
		day := []uint{1000000, 25000, 5770, 2310, 750, 650, 5350, 160, 150, 150, 100, 70, 170, 80, 80, 80, 50, 40}
		night := []uint{1000000, 1000000, 1000000, 46150, 15000, 12930, 107060, 3190, 3000, 3000, 2000, 1430, 3350, 1670, 1580, 1520, 990, 730}
		guarantee := []uint64{300, 900, 3500, 10000, 30000, 53198, 56000, 150000, 250000, 350000, 500000, 710399, 800000, 980000, 1170000, 1367702, 1672100, 2082499}
		for i := 0; i < 18; i++ {
			r := model.FlowerLightRule{ActivityID: a.ID, FlowerPosition: uint8(i + 1), DayProbabilityPPM: day[i], NightProbabilityPPM: night[i], GuaranteeCoinTotal: guarantee[i], Status: 1}
			if e := tx.Where("activity_id=? AND flower_position=?", a.ID, i+1).FirstOrCreate(&r).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println("seed completed: demo / 123456")
}

package main

import (
	"encoding/json"
	"flag"
	"flower-lottery-backend/initialize"
	"flower-lottery-backend/model"
	"flower-lottery-backend/utils"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

func main() {
	contentOnly := flag.Bool("content-only", false, "only update activities.rules_json")
	flag.Parse()

	cfg, e := initialize.Config()
	if e != nil {
		log.Fatal(e)
	}
	db, e := initialize.Database(cfg.Database)
	if e != nil {
		log.Fatal(e)
	}
	rulesJSON, e := json.Marshal(activityContentSeed())
	if e != nil {
		log.Fatal(e)
	}
	if *contentOnly {
		var activity model.Activity
		if e := db.Where("code=? AND deleted_at IS NULL", "flower-wish").First(&activity).Error; e != nil {
			log.Fatal(e)
		}
		if e := db.Model(&activity).Update("rules_json", rulesJSON).Error; e != nil {
			log.Fatal(e)
		}
		fmt.Println("activity content seed completed: flower-wish")
		return
	}

	hash, _ := utils.HashPassword("123456")
	e = db.Transaction(func(tx *gorm.DB) error {
		a := model.Activity{Code: "flower-wish", Name: "花愿奇遇", Status: 2, StartsAt: time.Now().Add(-24 * time.Hour), EndsAt: time.Now().AddDate(0, 1, 0), LeaderboardFreezesAt: time.Now().AddDate(0, 1, 0), Timezone: "Asia/Shanghai", RulesJSON: rulesJSON}
		if e := tx.Where("code=?", a.Code).Assign(map[string]any{"rules_json": rulesJSON}).FirstOrCreate(&a).Error; e != nil {
			return e
		}
		demoAvatar := "https://api.dicebear.com/9.x/thumbs/svg?seed=FlowerDemo"
		u := model.User{UserNo: "demo", Nickname: "体验用户", AvatarURL: demoAvatar, PasswordHash: hash, Status: 1}
		if e := tx.Where("user_no=?", u.UserNo).Assign(map[string]any{
			"nickname":      u.Nickname,
			"avatar_url":    u.AvatarURL,
			"password_hash": hash,
			"status":        1,
		}).FirstOrCreate(&u).Error; e != nil {
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
		type mockUserSeed struct {
			userNo       string
			nickname     string
			avatarSeed   string
			coinBalance  int64
			petalBalance int64
			score        uint64
		}
		mockUsers := []mockUserSeed{
			{userNo: "flower_001", nickname: "绮梦花园", avatarSeed: "QiMeng", coinBalance: 286000, petalBalance: 1380, score: 12860},
			{userNo: "flower_002", nickname: "星河入梦", avatarSeed: "XingHe", coinBalance: 252000, petalBalance: 1120, score: 10520},
			{userNo: "flower_003", nickname: "月下铃兰", avatarSeed: "YueXia", coinBalance: 218000, petalBalance: 980, score: 9880},
			{userNo: "flower_004", nickname: "云端来信", avatarSeed: "YunDuan", coinBalance: 196000, petalBalance: 860, score: 8460},
			{userNo: "flower_005", nickname: "花影微光", avatarSeed: "HuaYing", coinBalance: 173000, petalBalance: 790, score: 7930},
			{userNo: "flower_006", nickname: "晚风拾光", avatarSeed: "WanFeng", coinBalance: 156000, petalBalance: 720, score: 7280},
			{userNo: "flower_007", nickname: "银河邮差", avatarSeed: "YinHe", coinBalance: 148000, petalBalance: 650, score: 6650},
			{userNo: "flower_008", nickname: "桃枝小满", avatarSeed: "TaoZhi", coinBalance: 132000, petalBalance: 590, score: 5920},
			{userNo: "flower_009", nickname: "雾岛听潮", avatarSeed: "WuDao", coinBalance: 118000, petalBalance: 520, score: 5310},
			{userNo: "flower_010", nickname: "夏夜萤火", avatarSeed: "XiaYe", coinBalance: 105000, petalBalance: 460, score: 4760},
			{userNo: "flower_011", nickname: "晴空花信", avatarSeed: "QingKong", coinBalance: 92000, petalBalance: 380, score: 4180},
			{userNo: "flower_012", nickname: "蔷薇小筑", avatarSeed: "QiangWei", coinBalance: 86000, petalBalance: 320, score: 3620},
		}
		for index, mock := range mockUsers {
			lastLogin := time.Now().Add(-time.Duration(index+1) * 2 * time.Hour)
			avatarURL := fmt.Sprintf("https://api.dicebear.com/9.x/thumbs/svg?seed=%s", mock.avatarSeed)
			user := model.User{
				UserNo:       mock.userNo,
				Nickname:     mock.nickname,
				AvatarURL:    avatarURL,
				PasswordHash: hash,
				Status:       1,
				LastLoginAt:  &lastLogin,
			}
			if e := tx.Where("user_no=?", user.UserNo).Assign(map[string]any{
				"nickname":      user.Nickname,
				"avatar_url":    user.AvatarURL,
				"password_hash": hash,
				"status":        1,
				"last_login_at": lastLogin,
			}).FirstOrCreate(&user).Error; e != nil {
				return e
			}
			wallet := model.UserWallet{UserID: user.ID}
			if e := tx.Where("user_id=?", user.ID).Assign(map[string]any{
				"coin_balance":  mock.coinBalance,
				"petal_balance": mock.petalBalance,
			}).FirstOrCreate(&wallet).Error; e != nil {
				return e
			}
			reachedAt := time.Now().Add(-time.Duration(len(mockUsers)-index) * time.Minute)
			entry := model.LeaderboardEntry{ActivityID: a.ID, UserID: user.ID}
			if e := tx.Where("activity_id=? AND user_id=?", a.ID, user.ID).Assign(map[string]any{
				"score":      mock.score,
				"reached_at": reachedAt,
			}).FirstOrCreate(&entry).Error; e != nil {
				return e
			}
		}
		opts := [][2]uint64{{5, 300}, {50, 3000}, {150, 9000}, {100, 6000}, {1000, 60000}, {3000, 180000}}
		for i, v := range opts {
			o := model.ExchangeOption{ActivityID: a.ID, PetalAmount: v[0], CoinCost: v[1], SortNo: i + 1, Status: 1}
			if e := tx.Where("activity_id=? AND petal_amount=?", a.ID, v[0]).FirstOrCreate(&o).Error; e != nil {
				return e
			}
		}
		pools := []model.PrizePool{{ActivityID: a.ID, Code: "day", Name: "白昼许愿", PetalCostPerDraw: 5, CoinValuePerDraw: 300, SupportedDrawCounts: []byte(`[1,10,30]`), Status: 1, SortNo: 1}, {ActivityID: a.ID, Code: "night", Name: "星夜许愿", PetalCostPerDraw: 100, CoinValuePerDraw: 6000, SupportedDrawCounts: []byte(`[1,10,30]`), Status: 1, SortNo: 2}}
		items := []model.RewardItem{
			{ItemCode: "1001", Name: "金币", ItemType: "coin", ImageURL: "http://wespynextpic.afunapp.com/wespy_game_1652434382.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", Status: 1},
			{ItemCode: "1002", Name: "花瓣", ItemType: "petal", ImageURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/4UDCBWqC.png?_w=170&_h=170", Status: 1},
			{ItemCode: "1207244", Name: "烟花之恋戒指", ItemType: "item", ImageURL: "https://resource.afunapp.com/admin/1CD327FE-4ED9-46BA-8AF8-859BD5B875E0.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/MSWiUiXo.svg?_w=150&_h=150", Status: 1},
			{ItemCode: "1203743", Name: "闪闪心蝶戒指", ItemType: "item", ImageURL: "https://resource.afunapp.com/admin/9124A617-5445-42FA-95C5-70D5130D6B6B.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/NFvT38a9.svg?_w=150&_h=150", Status: 1},
			{ItemCode: "1205251", Name: "月亮游记戒指", ItemType: "item", ImageURL: "https://resource.afunapp.com/admin/047A106A-1227-4152-9C04-7071D9B37BD8.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/q4pbcEb0.svg?_w=150&_h=150", Status: 1},
			{ItemCode: "1205470", Name: "爱意翩跹戒指", ItemType: "item", ImageURL: "https://resource.afunapp.com/admin/0E2FCAD9-B8D0-4E5C-83B0-2C08AE7562EB.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/pjADZhm7.svg?_w=150&_h=150", Status: 1},
			{ItemCode: "1207751", Name: "真爱无敌戒指", ItemType: "item", ImageURL: "https://resource-inner.afunapp.com/admin/B693FB21-96A4-45FB-B5C4-50EE7FD2F53B.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/YsRArrZN.svg?_w=150&_h=150", Extra: []byte(`{"choice_group_code":"true-love-invincible","candidate_item_codes":["1207751","1207752","1207753"]}`), Status: 1},
			{ItemCode: "1207752", Name: "真爱无敌·铭文版戒指", ItemType: "item", ImageURL: "https://resource-inner.afunapp.com/admin/5984EC45-2B1D-4B64-A56E-DD37B04E38F7.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/oRc6Nwt5.svg?_w=150&_h=150", Status: 1},
			{ItemCode: "1207753", Name: "真爱无敌·铭文版戒指", ItemType: "item", ImageURL: "https://resource-inner.afunapp.com/admin/67EDC309-9680-458F-99D5-90D1F524637C.png?x-oss-process=image%2Fresize%2Cm_lfit%2Ch_200%2Cw_200", AnimationURL: "https://fe-center.afunapp.com/page-center/assets/zBvuCpRS/EZZaxmK2.svg?_w=150&_h=150", Status: 1},
		}
		itemByCode := make(map[string]model.RewardItem, len(items))
		for i := range items {
			attrs := map[string]any{
				"name":          items[i].Name,
				"item_type":     items[i].ItemType,
				"image_url":     items[i].ImageURL,
				"animation_url": items[i].AnimationURL,
				"extra":         items[i].Extra,
				"status":        items[i].Status,
			}
			if e := tx.Where("item_code=?", items[i].ItemCode).Assign(attrs).FirstOrCreate(&items[i]).Error; e != nil {
				return e
			}
			itemByCode[items[i].ItemCode] = items[i]
		}
		type poolRewardSeed struct {
			itemCode        string
			quantity        uint64
			weight          uint64
			choiceGroupCode string
		}
		poolRewardSeeds := map[string][]poolRewardSeed{
			"day": {
				{itemCode: "1001", quantity: 100, weight: 350000},
				{itemCode: "1001", quantity: 300, weight: 250000},
				{itemCode: "1002", quantity: 1, weight: 200000},
				{itemCode: "1207244", quantity: 1, weight: 100000},
				{itemCode: "1203743", quantity: 1, weight: 70000},
				{itemCode: "1205251", quantity: 1, weight: 25000},
				{itemCode: "1207751", quantity: 1, weight: 5000, choiceGroupCode: "true-love-invincible"},
			},
			"night": {
				{itemCode: "1001", quantity: 1000, weight: 250000},
				{itemCode: "1002", quantity: 10, weight: 200000},
				{itemCode: "1207244", quantity: 1, weight: 200000},
				{itemCode: "1203743", quantity: 1, weight: 150000},
				{itemCode: "1205251", quantity: 1, weight: 100000},
				{itemCode: "1205470", quantity: 1, weight: 70000},
				{itemCode: "1207751", quantity: 1, weight: 30000, choiceGroupCode: "true-love-invincible"},
			},
		}
		for i := range pools {
			poolAttrs := map[string]any{
				"name":                  pools[i].Name,
				"petal_cost_per_draw":   pools[i].PetalCostPerDraw,
				"coin_value_per_draw":   pools[i].CoinValuePerDraw,
				"supported_draw_counts": pools[i].SupportedDrawCounts,
				"status":                pools[i].Status,
				"sort_no":               pools[i].SortNo,
			}
			if e := tx.Where("activity_id=? AND code=?", a.ID, pools[i].Code).Assign(poolAttrs).FirstOrCreate(&pools[i]).Error; e != nil {
				return e
			}
			v := model.PrizePoolVersion{PrizePoolID: pools[i].ID, VersionNo: 1, Status: 1, TotalWeight: 1000000}
			now := time.Now().Add(-time.Hour)
			v.EffectiveAt = &now
			if e := tx.Where("prize_pool_id=? AND version_no=1", pools[i].ID).Assign(map[string]any{"status": 1, "total_weight": 1000000}).FirstOrCreate(&v).Error; e != nil {
				return e
			}
			if e := tx.Where("version_id=?", v.ID).Delete(&model.PrizePoolReward{}).Error; e != nil {
				return e
			}
			var totalWeight uint64
			for j, rewardSeed := range poolRewardSeeds[pools[i].Code] {
				item := itemByCode[rewardSeed.itemCode]
				pr := model.PrizePoolReward{
					VersionID:       v.ID,
					RewardItemID:    item.ID,
					Quantity:        rewardSeed.quantity,
					Weight:          rewardSeed.weight,
					ChoiceGroupCode: rewardSeed.choiceGroupCode,
					SortNo:          j + 1,
				}
				if e := tx.Create(&pr).Error; e != nil {
					return e
				}
				totalWeight += rewardSeed.weight
			}
			if totalWeight != v.TotalWeight {
				return fmt.Errorf("pool %s weight total is %d, expected %d", pools[i].Code, totalWeight, v.TotalWeight)
			}
		}
		stageRewardSeeds := []struct {
			requiredFlowers uint8
			itemCode        string
			quantity        uint64
		}{
			{requiredFlowers: 3, itemCode: "1001", quantity: 500},
			{requiredFlowers: 5, itemCode: "1002", quantity: 10},
			{requiredFlowers: 8, itemCode: "1207244", quantity: 1},
			{requiredFlowers: 11, itemCode: "1001", quantity: 2000},
			{requiredFlowers: 15, itemCode: "1205251", quantity: 1},
			{requiredFlowers: 17, itemCode: "1207751", quantity: 1},
		}
		if e := tx.Model(&model.StageRewardRule{}).Where("activity_id=?", a.ID).Update("status", 0).Error; e != nil {
			return e
		}
		for i, stageSeed := range stageRewardSeeds {
			item := itemByCode[stageSeed.itemCode]
			rule := model.StageRewardRule{
				ActivityID:      a.ID,
				RequiredFlowers: stageSeed.requiredFlowers,
				RewardItemID:    item.ID,
				Quantity:        stageSeed.quantity,
				Status:          1,
				SortNo:          i + 1,
			}
			attrs := map[string]any{
				"reward_item_id": item.ID,
				"quantity":       stageSeed.quantity,
				"status":         1,
				"sort_no":        i + 1,
			}
			if e := tx.Where("activity_id=? AND required_flowers=?", a.ID, stageSeed.requiredFlowers).Assign(attrs).FirstOrCreate(&rule).Error; e != nil {
				return e
			}
		}
		if e := tx.Model(&model.ChestRewardRule{}).Where("activity_id=?", a.ID).Update("status", 0).Error; e != nil {
			return e
		}
		chestRewardCodes := []string{"1205251", "1205470", "1207751"}
		for chestNo := uint8(1); chestNo <= 3; chestNo++ {
			for _, itemCode := range chestRewardCodes {
				item := itemByCode[itemCode]
				rule := model.ChestRewardRule{
					ActivityID:   a.ID,
					ChestNo:      chestNo,
					RewardItemID: item.ID,
					Quantity:     1,
					Weight:       1,
					Status:       1,
				}
				attrs := map[string]any{"quantity": 1, "weight": 1, "status": 1}
				if e := tx.Where("activity_id=? AND chest_no=? AND reward_item_id=?", a.ID, chestNo, item.ID).Assign(attrs).FirstOrCreate(&rule).Error; e != nil {
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

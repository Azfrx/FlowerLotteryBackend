package service

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flower-lottery-backend/common"
	"flower-lottery-backend/model"
	"flower-lottery-backend/repository"
	"gorm.io/gorm"
	"time"
)

type GameService struct{ repo *repository.GameRepository }

func NewGameService(r *repository.GameRepository) *GameService { return &GameService{repo: r} }

type HomeData struct {
	Activity     *model.Activity          `json:"activity"`
	Wallet       *model.UserWallet        `json:"wallet"`
	Pools        []model.PrizePool        `json:"pools"`
	Round        *model.UserActivityRound `json:"round"`
	StageRewards []model.StageRewardRule  `json:"stage_rewards"`
	ShowPreview  bool                     `json:"show_late_to_pay"`
}

func (s *GameService) Home(userID uint64) (*HomeData, error) {
	a, err := s.repo.CurrentActivity()
	if err != nil {
		return nil, err
	}
	w, err := repository.NewWalletRepository(s.repo.DB).Get(userID)
	if err != nil {
		return nil, err
	}
	p, err := s.repo.Pools(a.ID)
	if err != nil {
		return nil, err
	}
	round, err := s.repo.Round(userID, a.ID)
	if err == gorm.ErrRecordNotFound {
		round = &model.UserActivityRound{UserID: userID, ActivityID: a.ID, RoundNo: 1, Status: 1}
	} else if err != nil {
		return nil, err
	}
	stages, err := s.repo.StageRules(a.ID)
	return &HomeData{Activity: a, Wallet: w, Pools: p, Round: round, StageRewards: stages, ShowPreview: round.LitFlowerCount == 0}, err
}
func (s *GameService) Catalog() ([]model.PrizePoolReward, error) {
	a, err := s.repo.CurrentActivity()
	if err != nil {
		return nil, err
	}
	pools, err := s.repo.Pools(a.ID)
	if err != nil {
		return nil, err
	}
	all := []model.PrizePoolReward{}
	for _, p := range pools {
		v, e := s.repo.Version(p.ID)
		if e != nil {
			return nil, e
		}
		items, e := s.repo.PoolRewards(v.ID)
		if e != nil {
			return nil, e
		}
		all = append(all, items...)
	}
	return all, nil
}
func (s *GameService) Draw(userID uint64, poolCode string, count uint, requestID string) (*model.LotteryOrder, error) {
	var result *model.LotteryOrder
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		if old, e := r.ExistingOrder(userID, requestID); e == nil {
			result = old
			return nil
		} else if e != gorm.ErrRecordNotFound {
			return e
		}
		a, e := r.CurrentActivity()
		if e != nil {
			return e
		}
		pool, e := r.Pool(a.ID, poolCode)
		if e != nil {
			return e
		}
		if count != 1 && count != 10 && count != 30 {
			return common.NewError(400, 13003, "抽奖次数无效")
		}
		version, e := r.Version(pool.ID)
		if e != nil {
			return e
		}
		rewards, e := r.PoolRewards(version.ID)
		if e != nil {
			return e
		}
		rules, e := r.FlowerRules(a.ID)
		if e != nil {
			return e
		}
		wallet, e := r.WalletForUpdate(userID)
		if e != nil {
			return e
		}
		round, e := r.RoundForUpdate(userID, a.ID)
		if e != nil {
			return e
		}
		cost := pool.PetalCostPerDraw * uint64(count)
		if wallet.PetalBalance < int64(cost) {
			return common.NewError(409, 12003, "花瓣余额不足")
		}
		beforePetal := wallet.PetalBalance
		wallet.PetalBalance -= int64(cost)
		order := &model.LotteryOrder{OrderNo: newOrderNo("LT"), UserID: userID, ActivityID: a.ID, PrizePoolID: pool.ID, PoolVersionID: version.ID, RoundID: round.ID, OrderType: "normal", RequestedDrawCount: count, PetalCost: cost, FlowersBefore: round.LitFlowerCount, Status: 0, RequestID: requestID}
		if e = r.Create(order); e != nil {
			return e
		}
		for i := uint(1); i <= count && round.LitFlowerCount < 18; i++ {
			chosen, rv := pickReward(rewards)
			draw := model.LotteryDraw{LotteryOrderID: order.ID, DrawIndex: i, RewardItemID: chosen.RewardItemID, RewardQuantity: chosen.Quantity, RewardSnapshot: rewardJSON(chosen.RewardItem, chosen.Quantity), RandomValue: rv}
			round.CumulativeCoinValue += pool.CoinValuePerDraw
			lit := applyFlower(round, rules, poolCode, &draw)
			if e = r.Create(&draw); e != nil {
				return e
			}
			if lit {
				record := model.FlowerLightRecord{UserID: userID, ActivityID: a.ID, RoundID: round.ID, LotteryDrawID: draw.ID, FlowerPosition: *draw.FlowerPosition, TriggerType: map[bool]string{true: "guarantee", false: "probability"}[draw.GuaranteeTriggered == 1], CumulativeCoinValue: round.CumulativeCoinValue}
				if e = r.Create(&record); e != nil {
					return e
				}
				if round.LitFlowerCount%6 == 0 {
					round.ChestGrantedCount++
					ch := model.UserChestOpportunity{UserID: userID, ActivityID: a.ID, RoundID: round.ID, ChestNo: round.ChestGrantedCount, UnlockFlowerCount: round.LitFlowerCount, Status: 0}
					if e = r.Create(&ch); e != nil {
						return e
					}
				}
			}
			if e = grantReward(r, userID, a.ID, chosen.RewardItem, chosen.Quantity, "lottery", draw.ID, wallet); e != nil {
				return e
			}
			order.ExecutedDrawCount++
		}
		refund := pool.PetalCostPerDraw * uint64(count-order.ExecutedDrawCount)
		wallet.PetalBalance += int64(refund)
		order.PetalRefund = refund
		order.PetalCost -= refund
		order.FlowersAfter = round.LitFlowerCount
		order.LeaderboardScoreAdded = order.PetalCost
		order.Status = 1
		if round.LitFlowerCount == 18 {
			round.Status = 2
		}
		wallet.Version++
		if e = r.Save(wallet); e != nil {
			return e
		}
		if e = r.Save(round); e != nil {
			return e
		}
		if e = r.Save(order); e != nil {
			return e
		}
		biz := order.ID
		rows := []model.AssetTransaction{{UserID: userID, ActivityID: &a.ID, AssetType: "petal", ChangeAmount: -int64(cost), BalanceBefore: beforePetal, BalanceAfter: beforePetal - int64(cost), ReasonCode: "lottery_cost", BizType: "lottery", BizID: &biz, RequestID: requestID}}
		if refund > 0 {
			rows = append(rows, model.AssetTransaction{UserID: userID, ActivityID: &a.ID, AssetType: "petal", ChangeAmount: int64(refund), BalanceBefore: beforePetal - int64(cost), BalanceAfter: wallet.PetalBalance, ReasonCode: "lottery_refund", BizType: "lottery", BizID: &biz, RequestID: requestID})
		}
		if e = r.Create(&rows); e != nil {
			return e
		}
		if e = r.AddLeaderboard(a.ID, userID, order.PetalCost); e != nil {
			return e
		}
		result = order
		return nil
	})
	return result, err
}
func applyFlower(round *model.UserActivityRound, rules []model.FlowerLightRule, pool string, draw *model.LotteryDraw) bool {
	lit := false
	for round.LitFlowerCount < 18 {
		rule := rules[round.LitFlowerCount]
		prob := rule.DayProbabilityPPM
		if pool == "night" {
			prob = rule.NightProbabilityPPM
		}
		random := uint(randomN(1000000))
		guarantee := round.CumulativeCoinValue >= rule.GuaranteeCoinTotal
		if !guarantee && random >= prob {
			if !lit {
				draw.FlowerRandomValue = &random
				draw.FlowerProbabilityPPM = &prob
			}
			break
		}
		round.LitFlowerCount++
		pos := round.LitFlowerCount
		draw.FlowerLit = 1
		draw.FlowerPosition = &pos
		draw.FlowerRandomValue = &random
		draw.FlowerProbabilityPPM = &prob
		if guarantee {
			draw.GuaranteeTriggered = 1
		}
		lit = true
		if !guarantee {
			break
		}
	}
	return lit
}
func pickReward(items []model.PrizePoolReward) (model.PrizePoolReward, uint64) {
	var total uint64
	for _, v := range items {
		total += v.Weight
	}
	n := randomN(total)
	var sum uint64
	for _, v := range items {
		sum += v.Weight
		if n < sum {
			return v, n
		}
	}
	return items[len(items)-1], n
}
func randomN(max uint64) uint64 {
	if max == 0 {
		return 0
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:]) % max
}
func rewardJSON(item model.RewardItem, q uint64) []byte {
	v, _ := json.Marshal(map[string]any{"item_code": item.ItemCode, "name": item.Name, "quantity": q, "image_url": item.ImageURL})
	return v
}
func grantReward(r *repository.GameRepository, userID, activityID uint64, item model.RewardItem, q uint64, source string, sourceID uint64, wallet *model.UserWallet) error {
	if item.ItemType == "coin" {
		wallet.CoinBalance += int64(q)
		return nil
	}
	if item.ItemType == "petal" {
		wallet.PetalBalance += int64(q)
		return nil
	}
	now := timeNow()
	reward := model.UserReward{UserID: userID, ActivityID: activityID, RewardItemID: item.ID, Quantity: q, SourceType: source, SourceID: &sourceID, Status: 1, RewardSnapshot: rewardJSON(item, q), GrantedAt: &now}
	return r.Create(&reward)
}

var timeNow = time.Now

func (s *GameService) Orders(userID uint64, page, pageSize int) ([]model.LotteryOrder, int64, error) {
	return s.repo.Orders(userID, page, pageSize)
}
func (s *GameService) Rewards(userID uint64, page, pageSize int) ([]model.UserReward, int64, error) {
	return s.repo.Rewards(userID, page, pageSize)
}
func (s *GameService) Leaderboard(userID uint64) ([]model.LeaderboardEntry, *model.LeaderboardEntry, int64, error) {
	a, e := s.repo.CurrentActivity()
	if e != nil {
		return nil, nil, 0, e
	}
	return s.repo.Leaderboard(a.ID, userID)
}

var _ = errors.Is

func (s *GameService) OpenChest(userID, id uint64) ([]model.UserChestCandidate, error) {
	var out []model.UserChestCandidate
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		ch, e := r.ChestForUpdate(id, userID)
		if e != nil {
			return e
		}
		if ch.Status == 1 {
			returnCandidates(r, ch, &out)
		}
		if ch.Status != 0 {
			return common.NewError(409, 14001, "宝箱不可开启")
		}
		rules, e := r.ChestRules(ch.ActivityID, ch.ChestNo)
		if e != nil {
			return e
		}
		limit := 3
		if len(rules) < limit {
			limit = len(rules)
		}
		for i := 0; i < limit; i++ {
			candidate := model.UserChestCandidate{OpportunityID: ch.ID, RewardItemID: rules[i].RewardItemID, Quantity: rules[i].Quantity, RewardSnapshot: rewardJSON(rules[i].RewardItem, rules[i].Quantity)}
			if e = r.Create(&candidate); e != nil {
				return e
			}
		}
		now := time.Now()
		ch.Status = 1
		ch.OpenedAt = &now
		if e = r.Save(ch); e != nil {
			return e
		}
		return returnCandidates(r, ch, &out)
	})
	return out, err
}
func returnCandidates(r *repository.GameRepository, ch *model.UserChestOpportunity, out *[]model.UserChestCandidate) error {
	v, e := r.ChestCandidates(ch.ID)
	*out = v
	return e
}
func (s *GameService) SelectChest(userID, id, candidateID uint64) (*model.UserReward, error) {
	var out *model.UserReward
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		ch, e := r.ChestForUpdate(id, userID)
		if e != nil {
			return e
		}
		if ch.Status != 1 {
			return common.NewError(409, 14002, "宝箱不在待选择状态")
		}
		candidates, e := r.ChestCandidates(id)
		if e != nil {
			return e
		}
		var selected *model.UserChestCandidate
		for i := range candidates {
			if candidates[i].ID == candidateID {
				selected = &candidates[i]
				break
			}
		}
		if selected == nil {
			return common.NewError(400, 14003, "候选奖励无效")
		}
		selected.Selected = 1
		if e = r.Save(selected); e != nil {
			return e
		}
		now := time.Now()
		reward := &model.UserReward{UserID: userID, ActivityID: ch.ActivityID, RewardItemID: selected.RewardItemID, Quantity: selected.Quantity, SourceType: "chest", SourceID: &ch.ID, Status: 1, RewardSnapshot: selected.RewardSnapshot, GrantedAt: &now}
		if e = r.Create(reward); e != nil {
			return e
		}
		ch.Status = 2
		ch.SelectedAt = &now
		if e = r.Save(ch); e != nil {
			return e
		}
		var round model.UserActivityRound
		if e = r.DB.Where("id=?", ch.RoundID).First(&round).Error; e != nil {
			return e
		}
		round.ChestProcessedCount++
		if e = r.Save(&round); e != nil {
			return e
		}
		out = reward
		return nil
	})
	return out, err
}
func (s *GameService) ClaimStage(userID, ruleID uint64, requestID string) (*model.UserReward, error) {
	var out *model.UserReward
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		a, e := r.CurrentActivity()
		if e != nil {
			return e
		}
		round, e := r.RoundForUpdate(userID, a.ID)
		if e != nil {
			return e
		}
		rule, e := r.StageRule(ruleID, a.ID)
		if e != nil {
			return e
		}
		if round.LitFlowerCount < rule.RequiredFlowers {
			return common.NewError(409, 14004, "尚未达到领取条件")
		}
		claim := model.UserStageRewardClaim{UserID: userID, ActivityID: a.ID, RoundID: round.ID, StageRewardRuleID: rule.ID, Status: 1, RequestID: requestID}
		if e = r.Create(&claim); e != nil {
			return common.ErrDuplicateRequest
		}
		now := time.Now()
		reward := &model.UserReward{UserID: userID, ActivityID: a.ID, RewardItemID: rule.RewardItemID, Quantity: rule.Quantity, SourceType: "stage", SourceID: &claim.ID, Status: 1, RewardSnapshot: rewardJSON(rule.RewardItem, rule.Quantity), GrantedAt: &now}
		if e = r.Create(reward); e != nil {
			return e
		}
		out = reward
		return nil
	})
	return out, err
}
func (s *GameService) NextRound(userID uint64) (*model.UserActivityRound, error) {
	var out *model.UserActivityRound
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		a, e := r.CurrentActivity()
		if e != nil {
			return e
		}
		round, e := r.RoundForUpdate(userID, a.ID)
		if e != nil {
			return e
		}
		pendingChest, e := r.PendingChests(round.ID)
		if e != nil {
			return e
		}
		pendingStage, e := r.PendingStageClaims(round.ID, round.LitFlowerCount)
		if e != nil {
			return e
		}
		if round.LitFlowerCount < 18 || pendingChest > 0 || pendingStage > 0 {
			return common.NewError(409, 14005, "请先完成本轮宝箱和阶段奖励")
		}
		now := time.Now()
		round.Status = 3
		round.CompletedAt = &now
		if e = r.Save(round); e != nil {
			return e
		}
		next := &model.UserActivityRound{UserID: userID, ActivityID: a.ID, RoundNo: round.RoundNo + 1, Status: 1}
		if e = r.Create(next); e != nil {
			return e
		}
		out = next
		return nil
	})
	return out, err
}

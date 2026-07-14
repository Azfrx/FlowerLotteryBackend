package service

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flower-lottery-backend/common"
	"flower-lottery-backend/model"
	"flower-lottery-backend/repository"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

type GameService struct{ repo *repository.GameRepository }

func NewGameService(r *repository.GameRepository) *GameService { return &GameService{repo: r} }

type StageRewardResult struct {
	model.StageRewardRule
	Claimed bool
}

type HomeData struct {
	Activity                  *model.Activity          `json:"activity"`
	ActivityContent           model.ActivityContent    `json:"activity_content"`
	Wallet                    *model.UserWallet        `json:"wallet"`
	Pools                     []model.PrizePool        `json:"pools"`
	Round                     *model.UserActivityRound `json:"round"`
	StageRound                *model.UserActivityRound `json:"stage_round"`
	StageRewards              []StageRewardResult      `json:"stage_rewards"`
	FeaturedRewards           []model.RewardItem       `json:"featured_rewards"`
	OwnedRewardItemCodes      []string                 `json:"owned_reward_item_codes"`
	ClaimedStageRewardRuleIDs []uint64                 `json:"claimed_stage_reward_rule_ids"`
	PendingChests             []ChestOpportunityResult `json:"pending_chests"`
	PendingPreview            *LotteryResult           `json:"pending_preview"`
	LatestLeaderboardRewardID uint64                   `json:"latest_leaderboard_reward_id"`
	ShowPreview               bool                     `json:"show_late_to_pay"`
}

type LotteryRewardResult struct {
	DrawIndex    uint
	ItemCode     string
	Name         string
	Quantity     uint64
	ImageURL     string
	AnimationURL string
}

type LotteryResult struct {
	*model.LotteryOrder
	Rewards        []LotteryRewardResult
	UnlockedChests []ChestOpportunityResult
}

type ChestRewardResult struct {
	ItemCode     string
	Name         string
	Quantity     uint64
	ImageURL     string
	AnimationURL string
}

type ChestOpportunityResult struct {
	ID                         uint64
	RoundID                    uint64
	ChestNo                    uint8
	UnlockFlowerCount          uint8
	Status                     uint8
	SelectedReward             *ChestRewardResult
	RequiresChoice             bool
	UnavailableRewardItemCodes []string
}

type ChestSummonResult struct {
	OpportunityID  uint64
	ChestNo        uint8
	Status         uint8
	Reward         ChestRewardResult
	RequiresChoice bool
}

type HistoryRecordResult struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Quantity  uint64    `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

type LotteryRewardInventoryResult struct {
	ItemCode     string
	Name         string
	Quantity     uint64
	ImageURL     string
	AnimationURL string
}

type LotteryHistoryResult struct {
	Day    []HistoryRecordResult `json:"day"`
	Night  []HistoryRecordResult `json:"night"`
	Flower []HistoryRecordResult `json:"flower"`
}

func (s *GameService) Home(userID uint64) (*HomeData, error) {
	a, err := s.repo.CurrentActivity()
	if err != nil {
		return nil, err
	}
	activityContent, err := parseActivityContent(a.RulesJSON)
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
	if round.ID > 0 {
		if err = reconcileUnlockedChestOpportunities(s.repo, round, 0); err != nil {
			return nil, err
		}
	}
	stageRound := round
	if pendingStageRound, stageErr := s.repo.StageDisplayRound(userID, a.ID); stageErr == nil {
		stageRound = pendingStageRound
	} else if stageErr != gorm.ErrRecordNotFound {
		return nil, stageErr
	}
	stages, err := s.repo.StageRules(a.ID)
	if err != nil {
		return nil, err
	}
	featured, err := s.repo.RewardItemsByCodes([]string{"1205251", "1207751", "1205470"})
	if err != nil {
		return nil, err
	}
	ownedCodes, err := s.repo.OwnedRewardItemCodes(userID, a.ID)
	if err != nil {
		return nil, err
	}
	claimedStageIDs := make([]uint64, 0)
	if stageRound.ID > 0 {
		claimedStageIDs, err = s.repo.ClaimedStageRewardRuleIDs(stageRound.ID)
		if err != nil {
			return nil, err
		}
	}
	claimedStageIDSet := make(map[uint64]struct{}, len(claimedStageIDs))
	for _, id := range claimedStageIDs {
		claimedStageIDSet[id] = struct{}{}
	}
	stageResults := make([]StageRewardResult, 0, len(stages))
	for _, stage := range stages {
		_, claimed := claimedStageIDSet[stage.ID]
		stageResults = append(stageResults, StageRewardResult{
			StageRewardRule: stage,
			Claimed:         claimed,
		})
	}
	latestLeaderboardRewardID, err := s.repo.LatestLeaderboardRewardID(userID, a.ID)
	if err != nil {
		return nil, err
	}
	pendingChests, err := s.repo.PendingChestOpportunities(userID, a.ID)
	if err != nil {
		return nil, err
	}
	pendingChestResults, err := chestOpportunityResults(s.repo, pendingChests)
	if err != nil {
		return nil, err
	}
	normalOrderCount, err := s.repo.NormalOrderCount(userID, a.ID)
	if err != nil {
		return nil, err
	}
	previewOrderCount, err := s.repo.PreviewOrderCount(userID, a.ID)
	if err != nil {
		return nil, err
	}
	var pendingPreview *LotteryResult
	pendingPreviewOrder, previewErr := s.repo.PendingPreviewOrder(userID, a.ID)
	if previewErr == nil {
		pendingPreview, err = lotteryResultForOrder(s.repo, pendingPreviewOrder, false)
		if err != nil {
			return nil, err
		}
	} else if previewErr != gorm.ErrRecordNotFound {
		return nil, previewErr
	}
	return &HomeData{
		Activity:                  a,
		ActivityContent:           activityContent,
		Wallet:                    w,
		Pools:                     p,
		Round:                     round,
		StageRound:                stageRound,
		StageRewards:              stageResults,
		FeaturedRewards:           featured,
		OwnedRewardItemCodes:      ownedCodes,
		ClaimedStageRewardRuleIDs: claimedStageIDs,
		PendingChests:             pendingChestResults,
		PendingPreview:            pendingPreview,
		LatestLeaderboardRewardID: latestLeaderboardRewardID,
		ShowPreview:               normalOrderCount == 0 && previewOrderCount == 0,
	}, nil
}

func parseActivityContent(raw []byte) (model.ActivityContent, error) {
	var content model.ActivityContent
	if len(raw) == 0 {
		return content, nil
	}
	if err := json.Unmarshal(raw, &content); err != nil {
		return content, fmt.Errorf("parse activity rules_json: %w", err)
	}
	return content, nil
}

func (s *GameService) ActivityContent() (model.ActivityContent, error) {
	activity, err := s.repo.CurrentActivity()
	if err != nil {
		return model.ActivityContent{}, err
	}
	return parseActivityContent(activity.RulesJSON)
}

func chestRewardResult(item model.RewardItem, quantity uint64) ChestRewardResult {
	name := item.Name
	if item.ItemCode == trueLoveChoiceRewardCode {
		name = "真爱无敌戒指"
	}
	return ChestRewardResult{
		ItemCode:     item.ItemCode,
		Name:         name,
		Quantity:     quantity,
		ImageURL:     item.ImageURL,
		AnimationURL: item.AnimationURL,
	}
}

func chestOpportunityResults(r *repository.GameRepository, opportunities []model.UserChestOpportunity) ([]ChestOpportunityResult, error) {
	results := make([]ChestOpportunityResult, 0, len(opportunities))
	for _, opportunity := range opportunities {
		unavailableCodes, err := r.SelectedChestRewardItemCodes(
			opportunity.RoundID,
			opportunity.ID,
		)
		if err != nil {
			return nil, err
		}
		result := ChestOpportunityResult{
			ID:                         opportunity.ID,
			RoundID:                    opportunity.RoundID,
			ChestNo:                    opportunity.ChestNo,
			UnlockFlowerCount:          opportunity.UnlockFlowerCount,
			Status:                     opportunity.Status,
			UnavailableRewardItemCodes: unavailableCodes,
		}
		if opportunity.Status == 1 {
			candidates, err := r.ChestCandidates(opportunity.ID)
			if err != nil {
				return nil, err
			}
			for _, candidate := range candidates {
				if candidate.Selected != 1 {
					continue
				}
				reward := chestRewardResult(candidate.RewardItem, candidate.Quantity)
				result.SelectedReward = &reward
				result.RequiresChoice = candidate.RewardItem.ItemCode == "1207751"
				break
			}
		}
		results = append(results, result)
	}
	return results, nil
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

func advanceCompletedRoundForDraw(r *repository.GameRepository, round *model.UserActivityRound) (*model.UserActivityRound, error) {
	if round.LitFlowerCount < 18 {
		return round, nil
	}
	pendingChests, err := r.PendingChests(round.ID)
	if err != nil {
		return nil, err
	}
	if pendingChests > 0 {
		return nil, common.NewError(409, 14005, "请先完成本轮戒指宝箱召唤")
	}

	now := time.Now()
	round.Status = 3
	round.CompletedAt = &now
	if err := r.Save(round); err != nil {
		return nil, err
	}

	next := &model.UserActivityRound{
		UserID:     round.UserID,
		ActivityID: round.ActivityID,
		RoundNo:    round.RoundNo + 1,
		Status:     1,
	}
	if err := r.Create(next); err != nil {
		return nil, err
	}
	return next, nil
}

func chestUnlockThresholds(before, after uint8) []uint8 {
	if after > 18 {
		after = 18
	}
	thresholds := make([]uint8, 0, 3)
	for threshold := uint8(6); threshold <= 18; threshold += 6 {
		if threshold > before && threshold <= after {
			thresholds = append(thresholds, threshold)
		}
	}
	return thresholds
}

func reconcileUnlockedChestOpportunities(r *repository.GameRepository, round *model.UserActivityRound, before uint8) error {
	if round.ID == 0 {
		return nil
	}

	roundChanged := false
	for _, unlockFlowerCount := range chestUnlockThresholds(before, round.LitFlowerCount) {
		chestNo := unlockFlowerCount / 6
		opportunity := model.UserChestOpportunity{
			UserID:            round.UserID,
			ActivityID:        round.ActivityID,
			RoundID:           round.ID,
			ChestNo:           chestNo,
			UnlockFlowerCount: unlockFlowerCount,
			Status:            0,
		}
		if err := r.DB.Where("round_id=? AND chest_no=?", round.ID, chestNo).
			FirstOrCreate(&opportunity).Error; err != nil {
			return err
		}
		if round.ChestGrantedCount < chestNo {
			round.ChestGrantedCount = chestNo
			roundChanged = true
		}
	}
	if roundChanged {
		return r.Save(round)
	}
	return nil
}

func (s *GameService) Draw(userID uint64, poolCode string, count uint, requestID string) (*LotteryResult, error) {
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
		if _, pendingPreviewErr := r.PendingPreviewOrder(userID, a.ID); pendingPreviewErr == nil {
			return common.NewError(409, 13012, "请先处理先抽后付预览奖励")
		} else if pendingPreviewErr != gorm.ErrRecordNotFound {
			return pendingPreviewErr
		}
		round, e = advanceCompletedRoundForDraw(r, round)
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
			flowersBeforeDraw := round.LitFlowerCount
			lit := applyFlower(round, rules, poolCode, &draw)
			if e = r.Create(&draw); e != nil {
				return e
			}
			if lit {
				record := model.FlowerLightRecord{UserID: userID, ActivityID: a.ID, RoundID: round.ID, LotteryDrawID: draw.ID, FlowerPosition: *draw.FlowerPosition, TriggerType: map[bool]string{true: "guarantee", false: "probability"}[draw.GuaranteeTriggered == 1], CumulativeCoinValue: round.CumulativeCoinValue}
				if e = r.Create(&record); e != nil {
					return e
				}
				if e = reconcileUnlockedChestOpportunities(r, round, flowersBeforeDraw); e != nil {
					return e
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
	if err != nil {
		return nil, err
	}
	return lotteryResultForOrder(s.repo, result, true)
}

func lotteryResultForOrder(r *repository.GameRepository, order *model.LotteryOrder, includeUnlockedChests bool) (*LotteryResult, error) {
	draws, err := r.Draws(order.ID)
	if err != nil {
		return nil, err
	}
	rewards := make([]LotteryRewardResult, 0, len(draws))
	for _, draw := range draws {
		rewards = append(rewards, LotteryRewardResult{
			DrawIndex:    draw.DrawIndex,
			ItemCode:     draw.RewardItem.ItemCode,
			Name:         draw.RewardItem.Name,
			Quantity:     draw.RewardQuantity,
			ImageURL:     draw.RewardItem.ImageURL,
			AnimationURL: draw.RewardItem.AnimationURL,
		})
	}

	unlockedChestResults := make([]ChestOpportunityResult, 0)
	if includeUnlockedChests {
		unlockedChests, chestErr := r.ChestsUnlockedBetween(order.RoundID, order.FlowersBefore, order.FlowersAfter)
		if chestErr != nil {
			return nil, chestErr
		}
		unlockedChestResults, chestErr = chestOpportunityResults(r, unlockedChests)
		if chestErr != nil {
			return nil, chestErr
		}
	}

	return &LotteryResult{
		LotteryOrder:   order,
		Rewards:        rewards,
		UnlockedChests: unlockedChestResults,
	}, nil
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
	v, _ := json.Marshal(map[string]any{"item_code": item.ItemCode, "name": item.Name, "quantity": q, "image_url": item.ImageURL, "animation_url": item.AnimationURL})
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

func isDuplicateDatabaseError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (s *GameService) Orders(userID uint64, page, pageSize int) ([]model.LotteryOrder, int64, error) {
	return s.repo.Orders(userID, page, pageSize)
}
func (s *GameService) LotteryHistory(userID uint64) (*LotteryHistoryResult, error) {
	activity, err := s.repo.CurrentActivity()
	if err != nil {
		return nil, err
	}
	const historyLimit = 200
	dayRows, err := s.repo.LotteryHistory(userID, activity.ID, "day", historyLimit)
	if err != nil {
		return nil, err
	}
	nightRows, err := s.repo.LotteryHistory(userID, activity.ID, "night", historyLimit)
	if err != nil {
		return nil, err
	}
	flowerRows, err := s.repo.FlowerHistory(userID, activity.ID, historyLimit)
	if err != nil {
		return nil, err
	}

	result := &LotteryHistoryResult{
		Day:    make([]HistoryRecordResult, 0, len(dayRows)),
		Night:  make([]HistoryRecordResult, 0, len(nightRows)),
		Flower: make([]HistoryRecordResult, 0, len(flowerRows)),
	}
	for _, row := range dayRows {
		result.Day = append(result.Day, HistoryRecordResult{
			ID:        row.ID,
			Name:      row.RewardName,
			Quantity:  row.Quantity,
			CreatedAt: row.CreatedAt,
		})
	}
	for _, row := range nightRows {
		result.Night = append(result.Night, HistoryRecordResult{
			ID:        row.ID,
			Name:      row.RewardName,
			Quantity:  row.Quantity,
			CreatedAt: row.CreatedAt,
		})
	}
	for _, row := range flowerRows {
		result.Flower = append(result.Flower, HistoryRecordResult{
			ID:        row.ID,
			Name:      "点亮花朵",
			Quantity:  row.Quantity,
			CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}
func (s *GameService) ChestHistory(userID uint64) ([]HistoryRecordResult, error) {
	activity, err := s.repo.CurrentActivity()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ChestHistory(userID, activity.ID, 0)
	if err != nil {
		return nil, err
	}
	result := make([]HistoryRecordResult, 0, len(rows))
	for _, row := range rows {
		result = append(result, HistoryRecordResult{
			ID:        row.ID,
			Name:      row.RewardName,
			Quantity:  row.Quantity,
			CreatedAt: row.CreatedAt,
		})
	}
	return result, nil
}

func (s *GameService) LotteryRewardInventory(userID uint64) ([]LotteryRewardInventoryResult, error) {
	activity, err := s.repo.CurrentActivity()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.LotteryRewardInventory(userID, activity.ID)
	if err != nil {
		return nil, err
	}
	result := make([]LotteryRewardInventoryResult, 0, len(rows))
	for _, row := range rows {
		result = append(result, LotteryRewardInventoryResult{
			ItemCode:     row.ItemCode,
			Name:         row.Name,
			Quantity:     row.Quantity,
			ImageURL:     row.ImageURL,
			AnimationURL: row.AnimationURL,
		})
	}
	return result, nil
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

const trueLoveChoiceRewardCode = "1207751"

var chestRewardItemCodes = []string{"1205251", "1205470", trueLoveChoiceRewardCode}

var trueLoveChoiceItemCodes = map[string]struct{}{
	"1207751": {},
	"1207752": {},
	"1207753": {},
}

func selectedChestCandidate(candidates []model.UserChestCandidate) *model.UserChestCandidate {
	for i := range candidates {
		if candidates[i].Selected == 1 {
			return &candidates[i]
		}
	}
	return nil
}

func chestRulesWithFallback(r *repository.GameRepository, activityID uint64, chestNo uint8) ([]model.ChestRewardRule, error) {
	rules, err := r.ChestRules(activityID, chestNo)
	if err != nil || len(rules) > 0 {
		return rules, err
	}
	items, err := r.RewardItemsByCodes(chestRewardItemCodes)
	if err != nil {
		return nil, err
	}
	rules = make([]model.ChestRewardRule, 0, len(items))
	for _, item := range items {
		rules = append(rules, model.ChestRewardRule{
			ActivityID:   activityID,
			ChestNo:      chestNo,
			RewardItemID: item.ID,
			Quantity:     1,
			Weight:       1,
			Status:       1,
			RewardItem:   item,
		})
	}
	return rules, nil
}

func summonChestCandidate(r *repository.GameRepository, ch *model.UserChestOpportunity) (*model.UserChestCandidate, error) {
	candidates, err := r.ChestCandidates(ch.ID)
	if err != nil {
		return nil, err
	}
	if selected := selectedChestCandidate(candidates); selected != nil {
		return selected, nil
	}

	rules, err := chestRulesWithFallback(r, ch.ActivityID, ch.ChestNo)
	if err != nil {
		return nil, err
	}
	usedRewardIDs, err := r.SelectedChestRewardItemIDs(ch.RoundID, ch.ID)
	if err != nil {
		return nil, err
	}
	used := make(map[uint64]struct{}, len(usedRewardIDs))
	for _, rewardID := range usedRewardIDs {
		used[rewardID] = struct{}{}
	}
	available := make([]model.ChestRewardRule, 0, len(rules))
	for _, rule := range rules {
		if _, exists := used[rule.RewardItemID]; !exists {
			available = append(available, rule)
		}
	}
	if len(available) == 0 {
		return nil, common.NewError(500, 14007, "戒指宝箱奖励配置不足")
	}

	chosen := available[randomN(uint64(len(available)))]
	for i := range candidates {
		if candidates[i].RewardItemID != chosen.RewardItemID {
			continue
		}
		candidates[i].Selected = 1
		if err = r.Save(&candidates[i]); err != nil {
			return nil, err
		}
		return &candidates[i], nil
	}

	candidate := &model.UserChestCandidate{
		OpportunityID:  ch.ID,
		RewardItemID:   chosen.RewardItemID,
		Quantity:       chosen.Quantity,
		RewardSnapshot: rewardJSON(chosen.RewardItem, chosen.Quantity),
		Selected:       1,
		RewardItem:     chosen.RewardItem,
	}
	if err = r.Create(candidate); err != nil {
		return nil, err
	}
	return candidate, nil
}

func completeChest(r *repository.GameRepository, ch *model.UserChestOpportunity) error {
	now := time.Now()
	ch.Status = 2
	ch.SelectedAt = &now
	if err := r.Save(ch); err != nil {
		return err
	}
	round, err := r.RoundByIDForUpdate(ch.RoundID)
	if err != nil {
		return err
	}
	round.ChestProcessedCount++
	if round.LitFlowerCount < 18 || round.ChestProcessedCount < round.ChestGrantedCount {
		return r.Save(round)
	}

	round.Status = 3
	round.CompletedAt = &now
	if err = r.Save(round); err != nil {
		return err
	}
	next := model.UserActivityRound{
		UserID:     round.UserID,
		ActivityID: round.ActivityID,
		RoundNo:    round.RoundNo + 1,
		Status:     1,
	}
	return r.DB.Where(
		"user_id=? AND activity_id=? AND round_no=?",
		round.UserID,
		round.ActivityID,
		next.RoundNo,
	).FirstOrCreate(&next).Error
}

func createChestReward(r *repository.GameRepository, userID uint64, ch *model.UserChestOpportunity, item model.RewardItem, quantity uint64) error {
	now := time.Now()
	reward := &model.UserReward{
		UserID:         userID,
		ActivityID:     ch.ActivityID,
		RewardItemID:   item.ID,
		Quantity:       quantity,
		SourceType:     "chest",
		SourceID:       &ch.ID,
		Status:         1,
		RewardSnapshot: rewardJSON(item, quantity),
		GrantedAt:      &now,
	}
	return r.Create(reward)
}

func completedChestResult(r *repository.GameRepository, userID uint64, ch *model.UserChestOpportunity) (*ChestSummonResult, error) {
	reward, err := r.ChestGrantedReward(userID, ch.ID)
	if err != nil {
		return nil, err
	}
	return &ChestSummonResult{
		OpportunityID: ch.ID,
		ChestNo:       ch.ChestNo,
		Status:        ch.Status,
		Reward:        chestRewardResult(reward.RewardItem, reward.Quantity),
	}, nil
}

func (s *GameService) OpenChest(userID, id uint64, requestID string) (*ChestSummonResult, error) {
	_ = requestID
	var out *ChestSummonResult
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		ch, err := r.ChestForUpdate(id, userID)
		if err != nil {
			return err
		}
		if ch.Status == 2 {
			out, err = completedChestResult(r, userID, ch)
			return err
		}
		if ch.Status != 0 && ch.Status != 1 {
			return common.NewError(409, 14001, "宝箱不可开启")
		}

		candidate, err := summonChestCandidate(r, ch)
		if err != nil {
			return err
		}
		if ch.OpenedAt == nil {
			now := time.Now()
			ch.OpenedAt = &now
		}
		reward := chestRewardResult(candidate.RewardItem, candidate.Quantity)
		if candidate.RewardItem.ItemCode == trueLoveChoiceRewardCode {
			ch.Status = 1
			if err = r.Save(ch); err != nil {
				return err
			}
			out = &ChestSummonResult{
				OpportunityID:  ch.ID,
				ChestNo:        ch.ChestNo,
				Status:         ch.Status,
				Reward:         reward,
				RequiresChoice: true,
			}
			return nil
		}

		if err = createChestReward(r, userID, ch, candidate.RewardItem, candidate.Quantity); err != nil {
			return err
		}
		if err = completeChest(r, ch); err != nil {
			return err
		}
		out = &ChestSummonResult{
			OpportunityID: ch.ID,
			ChestNo:       ch.ChestNo,
			Status:        ch.Status,
			Reward:        reward,
		}
		return nil
	})
	return out, err
}

func (s *GameService) SelectChest(userID, id uint64, itemCode, requestID string) (*ChestSummonResult, error) {
	_ = requestID
	var out *ChestSummonResult
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		ch, err := r.ChestForUpdate(id, userID)
		if err != nil {
			return err
		}
		if ch.Status == 2 {
			out, err = completedChestResult(r, userID, ch)
			return err
		}
		if ch.Status != 1 {
			return common.NewError(409, 14002, "宝箱不在待选择状态")
		}
		candidates, err := r.ChestCandidates(id)
		if err != nil {
			return err
		}
		selected := selectedChestCandidate(candidates)
		if selected == nil || selected.RewardItem.ItemCode != trueLoveChoiceRewardCode {
			return common.NewError(409, 14002, "当前宝箱无需选择戒指")
		}
		if _, allowed := trueLoveChoiceItemCodes[itemCode]; !allowed {
			return common.NewError(400, 14003, "戒指款式无效")
		}
		item, err := r.RewardItemByCode(itemCode)
		if err != nil {
			return err
		}
		selectedItem := *item
		if selectedItem.ItemCode == trueLoveChoiceRewardCode {
			selectedItem.Name = "真爱无敌戒指"
		}
		if err = createChestReward(r, userID, ch, selectedItem, selected.Quantity); err != nil {
			return err
		}
		if err = completeChest(r, ch); err != nil {
			return err
		}
		out = &ChestSummonResult{
			OpportunityID: ch.ID,
			ChestNo:       ch.ChestNo,
			Status:        ch.Status,
			Reward:        chestRewardResult(selectedItem, selected.Quantity),
		}
		return nil
	})
	return out, err
}

type StageClaimResult struct {
	RuleID       uint64
	ItemCode     string
	Name         string
	Quantity     uint64
	ImageURL     string
	AnimationURL string
}

func (s *GameService) ClaimStage(userID, ruleID uint64, requestID string) (*StageClaimResult, error) {
	var out *StageClaimResult
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		a, e := r.CurrentActivity()
		if e != nil {
			return e
		}
		rule, e := r.StageRule(ruleID, a.ID)
		if e != nil {
			return e
		}
		round, e := r.StageRoundForClaimForUpdate(userID, a.ID, rule.ID, rule.RequiredFlowers)
		if e == gorm.ErrRecordNotFound {
			claimed, claimErr := r.StageClaimExistsForUser(userID, a.ID, rule.ID)
			if claimErr != nil {
				return claimErr
			}
			if claimed {
				return common.NewError(409, 14006, "该奖励已领取，请勿重复操作")
			}
			return common.NewError(409, 14004, "尚未达到领取条件")
		}
		if e != nil {
			return e
		}
		claim := model.UserStageRewardClaim{
			UserID:            userID,
			ActivityID:        a.ID,
			RoundID:           round.ID,
			StageRewardRuleID: rule.ID,
			Status:            1,
			RequestID:         requestID,
			ClaimedAt:         time.Now(),
		}
		if e = r.Create(&claim); e != nil {
			if isDuplicateDatabaseError(e) {
				return common.NewError(409, 14006, "该奖励已领取，请勿重复操作")
			}
			return e
		}
		wallet, e := r.WalletForUpdate(userID)
		if e != nil {
			return e
		}
		coinBefore := wallet.CoinBalance
		petalBefore := wallet.PetalBalance
		if e = grantReward(r, userID, a.ID, rule.RewardItem, rule.Quantity, "stage", claim.ID, wallet); e != nil {
			return e
		}
		if rule.RewardItem.ItemType == "coin" || rule.RewardItem.ItemType == "petal" {
			wallet.Version++
			if e = r.Save(wallet); e != nil {
				return e
			}
			assetType := rule.RewardItem.ItemType
			balanceBefore := coinBefore
			balanceAfter := wallet.CoinBalance
			if assetType == "petal" {
				balanceBefore = petalBefore
				balanceAfter = wallet.PetalBalance
			}
			biz := claim.ID
			transaction := model.AssetTransaction{
				UserID:        userID,
				ActivityID:    &a.ID,
				AssetType:     assetType,
				ChangeAmount:  int64(rule.Quantity),
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceAfter,
				ReasonCode:    "stage_reward",
				BizType:       "stage",
				BizID:         &biz,
				RequestID:     requestID,
			}
			if e = r.Create(&transaction); e != nil {
				return e
			}
		}
		out = &StageClaimResult{
			RuleID:       rule.ID,
			ItemCode:     rule.RewardItem.ItemCode,
			Name:         rule.RewardItem.Name,
			Quantity:     rule.Quantity,
			ImageURL:     rule.RewardItem.ImageURL,
			AnimationURL: rule.RewardItem.AnimationURL,
		}
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
		if round.LitFlowerCount < 18 || pendingChest > 0 {
			return common.NewError(409, 14005, "请先完成本轮宝箱召唤")
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

const previewDrawCount uint = 180
const previewPaymentPetals uint64 = 1800

func (s *GameService) Preview180(userID uint64, requestID string) (*LotteryResult, error) {
	var out *model.LotteryOrder
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		if old, existingErr := r.ExistingPreviewOrder(userID, requestID); existingErr == nil {
			out = old
			return nil
		} else if existingErr != gorm.ErrRecordNotFound {
			return existingErr
		}
		a, e := r.CurrentActivity()
		if e != nil {
			return e
		}
		round, e := r.RoundForUpdate(userID, a.ID)
		if e != nil {
			return e
		}
		n, e := r.NormalOrderCount(userID, a.ID)
		if e != nil {
			return e
		}
		if n > 0 {
			return common.NewError(409, 13010, "仅首次抽奖前可使用先抽后付")
		}
		previewCount, e := r.PreviewOrderCount(userID, a.ID)
		if e != nil {
			return e
		}
		if previewCount > 0 {
			return common.NewError(409, 13010, "先抽后付机会已使用")
		}
		pool, e := r.Pool(a.ID, "day")
		if e != nil {
			return e
		}
		version, e := r.Version(pool.ID)
		if e != nil {
			return e
		}
		items, e := r.PoolRewards(version.ID)
		if e != nil {
			return e
		}
		rules, e := r.FlowerRules(a.ID)
		if e != nil {
			return e
		}
		previewRound := *round
		order := &model.LotteryOrder{
			OrderNo:            newOrderNo("PV"),
			UserID:             userID,
			ActivityID:         a.ID,
			PrizePoolID:        pool.ID,
			PoolVersionID:      version.ID,
			RoundID:            round.ID,
			OrderType:          "preview",
			RequestedDrawCount: previewDrawCount,
			ExecutedDrawCount:  previewDrawCount,
			PetalCost:          previewPaymentPetals,
			FlowersBefore:      round.LitFlowerCount,
			Status:             2,
			RequestID:          requestID,
		}
		if e = r.Create(order); e != nil {
			return e
		}
		for i := uint(1); i <= previewDrawCount; i++ {
			chosen, rv := pickReward(items)
			d := model.LotteryDraw{LotteryOrderID: order.ID, DrawIndex: i, RewardItemID: chosen.RewardItemID, RewardQuantity: chosen.Quantity, RewardSnapshot: rewardJSON(chosen.RewardItem, chosen.Quantity), RandomValue: rv}
			previewRound.CumulativeCoinValue += pool.CoinValuePerDraw
			applyFlower(&previewRound, rules, "day", &d)
			if e = r.Create(&d); e != nil {
				return e
			}
		}
		order.FlowersAfter = previewRound.LitFlowerCount
		if e = r.Save(order); e != nil {
			return e
		}
		out = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lotteryResultForOrder(s.repo, out, false)
}
func (s *GameService) ConfirmPreview(userID, orderID uint64) (*LotteryResult, error) {
	var out *model.LotteryOrder
	err := s.repo.Tx(func(r *repository.GameRepository) error {
		o, e := r.OrderByIDForUpdate(orderID, userID)
		if e != nil {
			return e
		}
		if o.OrderType != "preview" || o.Status != 2 {
			return common.NewError(409, 13011, "预览订单状态无效")
		}
		wallet, e := r.WalletForUpdate(userID)
		if e != nil {
			return e
		}
		petalCost := o.PetalCost
		if petalCost == 0 {
			petalCost = previewPaymentPetals
		}
		if wallet.PetalBalance < int64(petalCost) {
			return common.NewError(409, 12003, "花瓣余额不足")
		}
		round, e := r.RoundByIDForUpdate(o.RoundID)
		if e != nil {
			return e
		}
		if round.UserID != userID || round.LitFlowerCount != o.FlowersBefore {
			return common.NewError(409, 13013, "当前花朵进度已变化，无法支付该预览订单")
		}
		var pool model.PrizePool
		if e = r.DB.Where("id=? AND activity_id=?", o.PrizePoolID, o.ActivityID).First(&pool).Error; e != nil {
			return e
		}
		petalBefore := wallet.PetalBalance
		wallet.PetalBalance -= int64(petalCost)
		draws, e := r.Draws(o.ID)
		if e != nil {
			return e
		}
		flowersBefore := round.LitFlowerCount
		cumulativeCoinValue := round.CumulativeCoinValue
		for _, d := range draws {
			cumulativeCoinValue += pool.CoinValuePerDraw
			if d.FlowerLit == 1 && d.FlowerPosition != nil {
				if *d.FlowerPosition > round.LitFlowerCount {
					round.LitFlowerCount = *d.FlowerPosition
				}
				record := model.FlowerLightRecord{
					UserID:              userID,
					ActivityID:          o.ActivityID,
					RoundID:             round.ID,
					LotteryDrawID:       d.ID,
					FlowerPosition:      *d.FlowerPosition,
					TriggerType:         map[bool]string{true: "guarantee", false: "probability"}[d.GuaranteeTriggered == 1],
					CumulativeCoinValue: cumulativeCoinValue,
				}
				if e = r.Create(&record); e != nil {
					return e
				}
			}
			if e = grantReward(r, userID, o.ActivityID, d.RewardItem, d.RewardQuantity, "preview", d.ID, wallet); e != nil {
				return e
			}
		}
		if o.FlowersAfter > round.LitFlowerCount {
			round.LitFlowerCount = o.FlowersAfter
		}
		round.CumulativeCoinValue = cumulativeCoinValue
		if e = reconcileUnlockedChestOpportunities(r, round, flowersBefore); e != nil {
			return e
		}
		if round.LitFlowerCount >= 18 {
			round.Status = 2
		}
		if e = r.Save(round); e != nil {
			return e
		}
		wallet.Version++
		if e = r.Save(wallet); e != nil {
			return e
		}
		o.Status = 1
		o.PetalCost = petalCost
		o.LeaderboardScoreAdded = petalCost
		o.PaidAt = &[]time.Time{time.Now()}[0]
		if e = r.Save(o); e != nil {
			return e
		}
		biz := o.ID
		row := model.AssetTransaction{UserID: userID, ActivityID: &o.ActivityID, AssetType: "petal", ChangeAmount: -int64(petalCost), BalanceBefore: petalBefore, BalanceAfter: petalBefore - int64(petalCost), ReasonCode: "preview_payment", BizType: "preview", BizID: &biz, RequestID: o.RequestID}
		if e = r.Create(&row); e != nil {
			return e
		}
		if e = r.AddLeaderboard(o.ActivityID, userID, petalCost); e != nil {
			return e
		}
		out = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lotteryResultForOrder(s.repo, out, true)
}
func (s *GameService) CancelPreview(userID, orderID uint64) error {
	return s.repo.Tx(func(r *repository.GameRepository) error {
		o, e := r.OrderByIDForUpdate(orderID, userID)
		if e != nil {
			return e
		}
		if o.OrderType != "preview" || o.Status != 2 {
			return common.NewError(409, 13011, "预览订单状态无效")
		}
		o.Status = 3
		if e = r.Save(o); e != nil {
			return e
		}
		return nil
	})
}

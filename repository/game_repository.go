package repository

import (
	"flower-lottery-backend/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GameRepository struct{ DB *gorm.DB }

type LotteryHistoryRow struct {
	ID         uint64
	RewardName string
	Quantity   uint64
	CreatedAt  time.Time
}

type FlowerHistoryRow struct {
	ID        uint64
	Quantity  uint64
	CreatedAt time.Time
}

type ChestHistoryRow struct {
	ID         uint64
	RewardName string
	Quantity   uint64
	CreatedAt  time.Time
}

type LotteryRewardInventoryRow struct {
	ItemCode     string
	Name         string
	Quantity     uint64
	ImageURL     string
	AnimationURL string
	LastWonAt    time.Time
}

func NewGameRepository(db *gorm.DB) *GameRepository { return &GameRepository{DB: db} }
func (r *GameRepository) Tx(fn func(*GameRepository) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error { return fn(&GameRepository{DB: tx}) })
}
func (r *GameRepository) CurrentActivity() (*model.Activity, error) {
	var v model.Activity
	err := r.DB.Where("status = 2 AND starts_at <= ? AND ends_at > ? AND deleted_at IS NULL", time.Now(), time.Now()).First(&v).Error
	return &v, err
}
func (r *GameRepository) Pools(activityID uint64) ([]model.PrizePool, error) {
	var v []model.PrizePool
	err := r.DB.Where("activity_id=? AND status=1 AND deleted_at IS NULL", activityID).Order("sort_no").Find(&v).Error
	return v, err
}
func (r *GameRepository) Pool(activityID uint64, code string) (*model.PrizePool, error) {
	var v model.PrizePool
	err := r.DB.Where("activity_id=? AND code=? AND status=1 AND deleted_at IS NULL", activityID, code).First(&v).Error
	return &v, err
}
func (r *GameRepository) Version(poolID uint64) (*model.PrizePoolVersion, error) {
	var v model.PrizePoolVersion
	err := r.DB.Where("prize_pool_id=? AND status=1 AND effective_at<=?", poolID, time.Now()).Order("version_no DESC").First(&v).Error
	return &v, err
}
func (r *GameRepository) PoolRewards(versionID uint64) ([]model.PrizePoolReward, error) {
	var v []model.PrizePoolReward
	err := r.DB.Preload("RewardItem").Where("version_id=?", versionID).Order("sort_no").Find(&v).Error
	return v, err
}
func (r *GameRepository) FlowerRules(activityID uint64) ([]model.FlowerLightRule, error) {
	var v []model.FlowerLightRule
	err := r.DB.Where("activity_id=? AND status=1", activityID).Order("flower_position").Find(&v).Error
	return v, err
}
func (r *GameRepository) StageRules(activityID uint64) ([]model.StageRewardRule, error) {
	var v []model.StageRewardRule
	err := r.DB.Preload("RewardItem").Where("activity_id=? AND status=1", activityID).Order("sort_no").Find(&v).Error
	return v, err
}
func (r *GameRepository) RewardItemsByCodes(codes []string) ([]model.RewardItem, error) {
	var v []model.RewardItem
	err := r.DB.Where("item_code IN ? AND status=1 AND deleted_at IS NULL", codes).Find(&v).Error
	return v, err
}
func (r *GameRepository) OwnedRewardItemCodes(userID, activityID uint64) ([]string, error) {
	var codes []string
	err := r.DB.Table("user_rewards AS ur").
		Joins("JOIN reward_items AS ri ON ri.id=ur.reward_item_id").
		Where("ur.user_id=? AND ur.activity_id=? AND ur.status IN (1,2) AND ri.deleted_at IS NULL", userID, activityID).
		Order("ri.item_code").
		Distinct("ri.item_code").
		Pluck("ri.item_code", &codes).Error
	return codes, err
}
func (r *GameRepository) ClaimedStageRewardRuleIDs(roundID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.DB.Model(&model.UserStageRewardClaim{}).
		Where("round_id=? AND status=1", roundID).
		Order("stage_reward_rule_id").
		Pluck("stage_reward_rule_id", &ids).Error
	return ids, err
}
func (r *GameRepository) StageDisplayRound(userID, activityID uint64) (*model.UserActivityRound, error) {
	var v model.UserActivityRound
	result := r.DB.Where(
		"user_id=? AND activity_id=? AND EXISTS (SELECT 1 FROM stage_reward_rules s WHERE s.activity_id=user_activity_rounds.activity_id AND s.status=1 AND s.required_flowers<=user_activity_rounds.lit_flower_count AND NOT EXISTS (SELECT 1 FROM user_stage_reward_claims c WHERE c.round_id=user_activity_rounds.id AND c.stage_reward_rule_id=s.id AND c.status=1))",
		userID,
		activityID,
	).Order("round_no ASC").Limit(1).Find(&v)
	if result.Error != nil {
		return &v, result.Error
	}
	if result.RowsAffected == 0 {
		return &v, gorm.ErrRecordNotFound
	}
	return &v, nil
}
func (r *GameRepository) LatestLeaderboardRewardID(userID, activityID uint64) (uint64, error) {
	var id uint64
	err := r.DB.Model(&model.UserReward{}).
		Where("user_id=? AND activity_id=? AND source_type='leaderboard' AND status IN (1,2)", userID, activityID).
		Select("COALESCE(MAX(id),0)").
		Scan(&id).Error
	return id, err
}
func (r *GameRepository) WalletForUpdate(userID uint64) (*model.UserWallet, error) {
	var v model.UserWallet
	err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id=?", userID).First(&v).Error
	return &v, err
}
func (r *GameRepository) RoundForUpdate(userID, activityID uint64) (*model.UserActivityRound, error) {
	var v model.UserActivityRound
	err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id=? AND activity_id=? AND status IN (1,2)", userID, activityID).Order("round_no DESC").First(&v).Error
	if err == gorm.ErrRecordNotFound {
		v = model.UserActivityRound{UserID: userID, ActivityID: activityID, RoundNo: 1, Status: 1}
		err = r.DB.Create(&v).Error
	}
	return &v, err
}
func (r *GameRepository) ExistingOrder(userID uint64, requestID string) (*model.LotteryOrder, error) {
	var v model.LotteryOrder
	err := r.DB.Where("user_id=? AND order_type='normal' AND request_id=?", userID, requestID).First(&v).Error
	return &v, err
}
func (r *GameRepository) ExistingPreviewOrder(userID uint64, requestID string) (*model.LotteryOrder, error) {
	var v model.LotteryOrder
	err := r.DB.Where("user_id=? AND order_type='preview' AND request_id=?", userID, requestID).First(&v).Error
	return &v, err
}
func (r *GameRepository) Save(v any) error   { return r.DB.Save(v).Error }
func (r *GameRepository) Create(v any) error { return r.DB.Create(v).Error }
func (r *GameRepository) AddLeaderboard(activityID, userID, score uint64) error {
	now := time.Now()
	entry := model.LeaderboardEntry{ActivityID: activityID, UserID: userID, Score: score, ReachedAt: now}
	return r.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "activity_id"}, {Name: "user_id"}}, DoUpdates: clause.Assignments(map[string]any{"score": gorm.Expr("score + ?", score), "reached_at": now, "updated_at": now})}).Create(&entry).Error
}
func (r *GameRepository) Round(userID, activityID uint64) (*model.UserActivityRound, error) {
	var v model.UserActivityRound
	err := r.DB.Where("user_id=? AND activity_id=? AND status IN (1,2)", userID, activityID).Order("round_no DESC").First(&v).Error
	return &v, err
}
func (r *GameRepository) RoundByIDForUpdate(id uint64) (*model.UserActivityRound, error) {
	var v model.UserActivityRound
	err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).First(&v).Error
	return &v, err
}
func (r *GameRepository) Orders(userID uint64, page, pageSize int) ([]model.LotteryOrder, int64, error) {
	var v []model.LotteryOrder
	var n int64
	q := r.DB.Model(&model.LotteryOrder{}).Where("user_id=?", userID)
	if err := q.Count(&n).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&v).Error
	return v, n, err
}
func (r *GameRepository) Rewards(userID uint64, page, pageSize int) ([]model.UserReward, int64, error) {
	var v []model.UserReward
	var n int64
	q := r.DB.Model(&model.UserReward{}).Where("user_id=?", userID)
	if err := q.Count(&n).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("RewardItem").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&v).Error
	return v, n, err
}
func (r *GameRepository) LotteryHistory(userID, activityID uint64, poolCode string, limit int) ([]LotteryHistoryRow, error) {
	var v []LotteryHistoryRow
	err := r.DB.Table("lottery_draws AS draw").
		Select("draw.id AS id, item.name AS reward_name, draw.reward_quantity AS quantity, draw.created_at AS created_at").
		Joins("JOIN lottery_orders AS lottery_order ON lottery_order.id=draw.lottery_order_id").
		Joins("JOIN prize_pools AS pool ON pool.id=lottery_order.prize_pool_id").
		Joins("JOIN reward_items AS item ON item.id=draw.reward_item_id").
		Where("lottery_order.user_id=? AND lottery_order.activity_id=? AND lottery_order.status=1 AND pool.code=?", userID, activityID, poolCode).
		Order("draw.id DESC").
		Limit(limit).
		Scan(&v).Error
	return v, err
}
func (r *GameRepository) FlowerHistory(userID, activityID uint64, limit int) ([]FlowerHistoryRow, error) {
	var v []FlowerHistoryRow
	err := r.DB.Table("lottery_orders AS lottery_order").
		Select("lottery_order.id AS id, lottery_order.flowers_after - lottery_order.flowers_before AS quantity, lottery_order.created_at AS created_at").
		Where("lottery_order.user_id=? AND lottery_order.activity_id=? AND lottery_order.status=1 AND lottery_order.flowers_after>lottery_order.flowers_before", userID, activityID).
		Order("lottery_order.id DESC").
		Limit(limit).
		Scan(&v).Error
	return v, err
}
func (r *GameRepository) ChestHistory(userID, activityID uint64, limit int) ([]ChestHistoryRow, error) {
	var v []ChestHistoryRow
	query := r.DB.Table("user_rewards AS user_reward").
		Select("user_reward.id AS id, item.name AS reward_name, user_reward.quantity AS quantity, COALESCE(user_reward.granted_at,user_reward.created_at) AS created_at").
		Joins("JOIN reward_items AS item ON item.id=user_reward.reward_item_id").
		Where("user_reward.user_id=? AND user_reward.activity_id=? AND user_reward.status IN (1,2) AND item.name LIKE ?", userID, activityID, "%戒指%").
		Order("user_reward.id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Scan(&v).Error
	return v, err
}

func (r *GameRepository) LotteryRewardInventory(userID, activityID uint64) ([]LotteryRewardInventoryRow, error) {
	var v []LotteryRewardInventoryRow
	err := r.DB.Table("lottery_draws AS draw").
		Select("item.item_code AS item_code, item.name AS name, SUM(draw.reward_quantity) AS quantity, item.image_url AS image_url, item.animation_url AS animation_url, MAX(lottery_order.created_at) AS last_won_at").
		Joins("JOIN lottery_orders AS lottery_order ON lottery_order.id=draw.lottery_order_id").
		Joins("JOIN reward_items AS item ON item.id=draw.reward_item_id").
		Where("lottery_order.user_id=? AND lottery_order.activity_id=? AND lottery_order.status=1 AND lottery_order.order_type IN ?", userID, activityID, []string{"normal", "preview"}).
		Group("item.id, item.item_code, item.name, item.image_url, item.animation_url").
		Order("last_won_at DESC, item.item_code ASC").
		Scan(&v).Error
	return v, err
}
func (r *GameRepository) Leaderboard(activityID, userID uint64) ([]model.LeaderboardEntry, *model.LeaderboardEntry, int64, error) {
	var top []model.LeaderboardEntry
	err := r.DB.Preload("User").Where("activity_id=?", activityID).Order("score DESC,reached_at ASC,user_id ASC").Limit(20).Find(&top).Error
	if err != nil {
		return nil, nil, 0, err
	}
	var self model.LeaderboardEntry
	if err = r.DB.Preload("User").Where("activity_id=? AND user_id=?", activityID, userID).First(&self).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, 0, err
	}
	var rank int64
	if self.ID > 0 {
		r.DB.Model(&model.LeaderboardEntry{}).Where("activity_id=? AND (score>? OR (score=? AND reached_at<?))", activityID, self.Score, self.Score, self.ReachedAt).Count(&rank)
		rank++
	}
	return top, &self, rank, nil
}

func (r *GameRepository) ChestForUpdate(id, userID uint64) (*model.UserChestOpportunity, error) {
	var v model.UserChestOpportunity
	err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND user_id=?", id, userID).First(&v).Error
	return &v, err
}
func (r *GameRepository) ChestRules(activityID uint64, chestNo uint8) ([]model.ChestRewardRule, error) {
	var v []model.ChestRewardRule
	err := r.DB.Preload("RewardItem").Where("activity_id=? AND chest_no=? AND status=1", activityID, chestNo).Order("id").Find(&v).Error
	return v, err
}
func (r *GameRepository) ChestCandidates(id uint64) ([]model.UserChestCandidate, error) {
	var v []model.UserChestCandidate
	err := r.DB.Preload("RewardItem").Where("opportunity_id=?", id).Order("id").Find(&v).Error
	return v, err
}
func (r *GameRepository) PendingChestOpportunities(userID, activityID uint64) ([]model.UserChestOpportunity, error) {
	var v []model.UserChestOpportunity
	err := r.DB.Where("user_id=? AND activity_id=? AND status IN (0,1)", userID, activityID).
		Order("round_id ASC, chest_no ASC").
		Find(&v).Error
	return v, err
}
func (r *GameRepository) ChestsUnlockedBetween(roundID uint64, before, after uint8) ([]model.UserChestOpportunity, error) {
	var v []model.UserChestOpportunity
	err := r.DB.Where("round_id=? AND unlock_flower_count>? AND unlock_flower_count<=? AND status IN (0,1)", roundID, before, after).
		Order("chest_no ASC").
		Find(&v).Error
	return v, err
}
func (r *GameRepository) SelectedChestRewardItemIDs(roundID, excludeOpportunityID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.DB.Table("user_chest_candidates AS candidate").
		Joins("JOIN user_chest_opportunities AS opportunity ON opportunity.id=candidate.opportunity_id").
		Where("opportunity.round_id=? AND opportunity.id<>? AND candidate.selected=1", roundID, excludeOpportunityID).
		Distinct("candidate.reward_item_id").
		Pluck("candidate.reward_item_id", &ids).Error
	return ids, err
}
func (r *GameRepository) SelectedChestRewardItemCodes(roundID, excludeOpportunityID uint64) ([]string, error) {
	var codes []string
	err := r.DB.Table("user_chest_candidates AS candidate").
		Joins("JOIN user_chest_opportunities AS opportunity ON opportunity.id=candidate.opportunity_id").
		Joins("JOIN reward_items AS reward ON reward.id=candidate.reward_item_id").
		Where("opportunity.round_id=? AND opportunity.id<>? AND candidate.selected=1", roundID, excludeOpportunityID).
		Order("reward.item_code").
		Distinct("reward.item_code").
		Pluck("reward.item_code", &codes).Error
	return codes, err
}
func (r *GameRepository) ChestGrantedReward(userID, opportunityID uint64) (*model.UserReward, error) {
	var v model.UserReward
	err := r.DB.Preload("RewardItem").
		Where("user_id=? AND source_type='chest' AND source_id=? AND status IN (1,2)", userID, opportunityID).
		Order("id DESC").
		First(&v).Error
	return &v, err
}
func (r *GameRepository) RewardItemByCode(code string) (*model.RewardItem, error) {
	var v model.RewardItem
	err := r.DB.Where("item_code=? AND status=1 AND deleted_at IS NULL", code).First(&v).Error
	return &v, err
}
func (r *GameRepository) StageRule(id, activityID uint64) (*model.StageRewardRule, error) {
	var v model.StageRewardRule
	err := r.DB.Preload("RewardItem").Where("id=? AND activity_id=? AND status=1", id, activityID).First(&v).Error
	return &v, err
}
func (r *GameRepository) StageClaimExists(roundID, ruleID uint64) (bool, error) {
	var count int64
	err := r.DB.Model(&model.UserStageRewardClaim{}).
		Where("round_id=? AND stage_reward_rule_id=? AND status=1", roundID, ruleID).
		Count(&count).Error
	return count > 0, err
}
func (r *GameRepository) StageClaimExistsForUser(userID, activityID, ruleID uint64) (bool, error) {
	var count int64
	err := r.DB.Model(&model.UserStageRewardClaim{}).
		Where("user_id=? AND activity_id=? AND stage_reward_rule_id=? AND status=1", userID, activityID, ruleID).
		Count(&count).Error
	return count > 0, err
}
func (r *GameRepository) StageRoundForClaimForUpdate(userID, activityID, ruleID uint64, requiredFlowers uint8) (*model.UserActivityRound, error) {
	var v model.UserActivityRound
	err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
		"user_id=? AND activity_id=? AND lit_flower_count>=? AND NOT EXISTS (SELECT 1 FROM user_stage_reward_claims c WHERE c.round_id=user_activity_rounds.id AND c.stage_reward_rule_id=? AND c.status=1)",
		userID,
		activityID,
		requiredFlowers,
		ruleID,
	).Order("round_no ASC").First(&v).Error
	return &v, err
}
func (r *GameRepository) PendingStageClaims(roundID, activityID uint64, lit uint8) (int64, error) {
	var eligible int64
	err := r.DB.Table("stage_reward_rules s").Where("s.activity_id=? AND s.required_flowers<=? AND s.status=1 AND NOT EXISTS (SELECT 1 FROM user_stage_reward_claims c WHERE c.round_id=? AND c.stage_reward_rule_id=s.id AND c.status=1)", activityID, lit, roundID).Count(&eligible).Error
	return eligible, err
}
func (r *GameRepository) PendingChests(roundID uint64) (int64, error) {
	var n int64
	err := r.DB.Model(&model.UserChestOpportunity{}).Where("round_id=? AND status IN (0,1)", roundID).Count(&n).Error
	return n, err
}

func (r *GameRepository) OrderByID(id, userID uint64) (*model.LotteryOrder, error) {
	var v model.LotteryOrder
	err := r.DB.Where("id=? AND user_id=?", id, userID).First(&v).Error
	return &v, err
}
func (r *GameRepository) OrderByIDForUpdate(id, userID uint64) (*model.LotteryOrder, error) {
	var v model.LotteryOrder
	err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND user_id=?", id, userID).First(&v).Error
	return &v, err
}
func (r *GameRepository) Draws(orderID uint64) ([]model.LotteryDraw, error) {
	var v []model.LotteryDraw
	err := r.DB.Preload("RewardItem").Where("lottery_order_id=?", orderID).Order("draw_index").Find(&v).Error
	return v, err
}
func (r *GameRepository) NormalOrderCount(userID, activityID uint64) (int64, error) {
	var n int64
	err := r.DB.Model(&model.LotteryOrder{}).Where("user_id=? AND activity_id=? AND order_type='normal' AND status=1", userID, activityID).Count(&n).Error
	return n, err
}
func (r *GameRepository) PreviewOrderCount(userID, activityID uint64) (int64, error) {
	var n int64
	err := r.DB.Model(&model.LotteryOrder{}).
		Where("user_id=? AND activity_id=? AND order_type='preview' AND status IN (1,2,3,4)", userID, activityID).
		Count(&n).Error
	return n, err
}
func (r *GameRepository) PendingPreviewOrder(userID, activityID uint64) (*model.LotteryOrder, error) {
	var v model.LotteryOrder
	err := r.DB.Where("user_id=? AND activity_id=? AND order_type='preview' AND status=2", userID, activityID).
		Order("id DESC").
		First(&v).Error
	return &v, err
}

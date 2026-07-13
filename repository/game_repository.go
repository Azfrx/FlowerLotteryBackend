package repository

import (
	"flower-lottery-backend/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GameRepository struct{ DB *gorm.DB }

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
	err := r.DB.Preload("RewardItem").Where("activity_id=? AND chest_no=? AND status=1", activityID, chestNo).Find(&v).Error
	return v, err
}
func (r *GameRepository) ChestCandidates(id uint64) ([]model.UserChestCandidate, error) {
	var v []model.UserChestCandidate
	err := r.DB.Preload("RewardItem").Where("opportunity_id=?", id).Find(&v).Error
	return v, err
}
func (r *GameRepository) StageRule(id, activityID uint64) (*model.StageRewardRule, error) {
	var v model.StageRewardRule
	err := r.DB.Preload("RewardItem").Where("id=? AND activity_id=? AND status=1", id, activityID).First(&v).Error
	return &v, err
}
func (r *GameRepository) PendingStageClaims(roundID uint64, lit uint8) (int64, error) {
	var eligible int64
	err := r.DB.Table("stage_reward_rules s").Where("s.required_flowers<=? AND s.status=1 AND NOT EXISTS (SELECT 1 FROM user_stage_reward_claims c WHERE c.round_id=? AND c.stage_reward_rule_id=s.id)", lit, roundID).Count(&eligible).Error
	return eligible, err
}
func (r *GameRepository) PendingChests(roundID uint64) (int64, error) {
	var n int64
	err := r.DB.Model(&model.UserChestOpportunity{}).Where("round_id=? AND status<>2", roundID).Count(&n).Error
	return n, err
}

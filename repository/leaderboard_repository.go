package repository

import (
	"errors"
	"flower-lottery-backend/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrLeaderboardFrozen = errors.New("leaderboard is frozen")

func leaderboardFreezeDeadline(activity *model.Activity) time.Time {
	deadline := activity.EndsAt
	if !activity.LeaderboardFreezesAt.IsZero() &&
		(deadline.IsZero() || activity.LeaderboardFreezesAt.Before(deadline)) {
		deadline = activity.LeaderboardFreezesAt
	}
	return deadline
}

func leaderboardWritableAt(activity *model.Activity, now time.Time) bool {
	if !ActivityPlayableAt(activity, now) {
		return false
	}
	deadline := leaderboardFreezeDeadline(activity)
	return deadline.IsZero() || now.Before(deadline)
}

func leaderboardShouldFreezeAt(activity *model.Activity, now time.Time) bool {
	if activity.Status == 3 || activity.Status == 4 {
		return true
	}
	if activity.Status != 2 {
		return false
	}
	deadline := leaderboardFreezeDeadline(activity)
	return !deadline.IsZero() && !now.Before(deadline)
}

func leaderboardSnapshotTime(activity *model.Activity, now time.Time) time.Time {
	deadline := leaderboardFreezeDeadline(activity)
	if (activity.Status == 3 || activity.Status == 4) &&
		(deadline.IsZero() || now.Before(deadline)) {
		return now
	}
	if deadline.IsZero() {
		return now
	}
	return deadline
}

func (r *GameRepository) LeaderboardActivity(now time.Time) (*model.Activity, error) {
	return r.DisplayActivity(now)
}

func (r *GameRepository) AddLeaderboard(activityID, userID, score uint64) error {
	now := time.Now()
	var activity model.Activity
	if err := r.DB.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id=? AND deleted_at IS NULL", activityID).
		First(&activity).Error; err != nil {
		return err
	}

	var snapshotCount int64
	if err := r.DB.Model(&model.LeaderboardSnapshot{}).
		Where("activity_id=?", activityID).
		Count(&snapshotCount).Error; err != nil {
		return err
	}
	if snapshotCount > 0 || !leaderboardWritableAt(&activity, now) {
		return ErrLeaderboardFrozen
	}

	entry := model.LeaderboardEntry{
		ActivityID: activityID,
		UserID:     userID,
		Score:      score,
		ReachedAt:  now,
	}
	return r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "activity_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"score":      gorm.Expr("score + ?", score),
			"reached_at": now,
			"updated_at": now,
		}),
	}).Create(&entry).Error
}

func (r *GameRepository) FreezeLeaderboardIfDue(activityID uint64, now time.Time) (bool, error) {
	var frozen bool
	err := r.Tx(func(tx *GameRepository) error {
		var activity model.Activity
		if err := tx.DB.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id=? AND deleted_at IS NULL", activityID).
			First(&activity).Error; err != nil {
			return err
		}

		var snapshotCount int64
		if err := tx.DB.Model(&model.LeaderboardSnapshot{}).
			Where("activity_id=?", activityID).
			Count(&snapshotCount).Error; err != nil {
			return err
		}
		if snapshotCount > 0 {
			frozen = true
			return nil
		}
		if !leaderboardShouldFreezeAt(&activity, now) {
			return nil
		}

		frozenAt := leaderboardSnapshotTime(&activity, now)
		if err := tx.DB.Exec(`
			INSERT INTO leaderboard_snapshots
				(activity_id,user_id,rank_no,score,reached_at,reward_status,frozen_at,created_at)
			SELECT activity_id,user_id,
				ROW_NUMBER() OVER (ORDER BY score DESC,reached_at ASC,user_id ASC),
				score,reached_at,0,?,?
			FROM leaderboard_entries
			WHERE activity_id=?`, frozenAt, now, activity.ID).Error; err != nil {
			return err
		}
		frozen = true
		return nil
	})
	return frozen, err
}

func (r *GameRepository) FreezeDueLeaderboards(now time.Time) error {
	var activityIDs []uint64
	if err := r.DB.Model(&model.Activity{}).
		Where("deleted_at IS NULL").
		Where("status IN (3,4) OR (status=2 AND (ends_at<=? OR leaderboard_freezes_at<=?))", now, now).
		Order("id").
		Pluck("id", &activityIDs).Error; err != nil {
		return err
	}
	for _, activityID := range activityIDs {
		if _, err := r.FreezeLeaderboardIfDue(activityID, now); err != nil {
			return err
		}
	}
	return nil
}

func (r *GameRepository) ResetLeaderboardSnapshots(activityID uint64) (int64, error) {
	result := r.DB.Where("activity_id=?", activityID).Delete(&model.LeaderboardSnapshot{})
	return result.RowsAffected, result.Error
}

func snapshotLeaderboardEntry(snapshot model.LeaderboardSnapshot) model.LeaderboardEntry {
	return model.LeaderboardEntry{
		ID:         snapshot.ID,
		ActivityID: snapshot.ActivityID,
		UserID:     snapshot.UserID,
		Score:      snapshot.Score,
		ReachedAt:  snapshot.ReachedAt,
		UpdatedAt:  snapshot.FrozenAt,
		User:       snapshot.User,
	}
}

func (r *GameRepository) Leaderboard(activityID, userID uint64, frozen bool) ([]model.LeaderboardEntry, *model.LeaderboardEntry, int64, error) {
	if frozen {
		var snapshots []model.LeaderboardSnapshot
		if err := r.DB.Preload("User").
			Where("activity_id=?", activityID).
			Order("rank_no ASC").
			Limit(20).
			Find(&snapshots).Error; err != nil {
			return nil, nil, 0, err
		}
		top := make([]model.LeaderboardEntry, 0, len(snapshots))
		for _, snapshot := range snapshots {
			top = append(top, snapshotLeaderboardEntry(snapshot))
		}

		var selfSnapshot model.LeaderboardSnapshot
		err := r.DB.Preload("User").
			Where("activity_id=? AND user_id=?", activityID, userID).
			First(&selfSnapshot).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, nil, 0, err
		}
		if err == gorm.ErrRecordNotFound {
			return top, &model.LeaderboardEntry{}, 0, nil
		}
		self := snapshotLeaderboardEntry(selfSnapshot)
		return top, &self, int64(selfSnapshot.RankNo), nil
	}

	var top []model.LeaderboardEntry
	err := r.DB.Preload("User").
		Where("activity_id=?", activityID).
		Order("score DESC,reached_at ASC,user_id ASC").
		Limit(20).
		Find(&top).Error
	if err != nil {
		return nil, nil, 0, err
	}
	var self model.LeaderboardEntry
	if err = r.DB.Preload("User").
		Where("activity_id=? AND user_id=?", activityID, userID).
		First(&self).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, nil, 0, err
	}
	var rank int64
	if self.ID > 0 {
		if err = r.DB.Model(&model.LeaderboardEntry{}).
			Where(`activity_id=? AND (
				score>? OR
				(score=? AND reached_at<?) OR
				(score=? AND reached_at=? AND user_id<?)
			)`, activityID, self.Score, self.Score, self.ReachedAt, self.Score, self.ReachedAt, self.UserID).
			Count(&rank).Error; err != nil {
			return nil, nil, 0, err
		}
		rank++
	}
	return top, &self, rank, nil
}

package repository

import "time"

type RewardAnnouncementRow struct {
	ID         uint64
	Nickname   string
	ItemCode   string
	RewardName string
	GrantedAt  time.Time
}

func (r *GameRepository) RewardAnnouncements(
	afterID uint64,
	since time.Time,
	itemCodes []string,
	limit int,
	latest bool,
) ([]RewardAnnouncementRow, error) {
	if len(itemCodes) == 0 || limit <= 0 {
		return []RewardAnnouncementRow{}, nil
	}
	if limit > 100 {
		limit = 100
	}

	now := time.Now()
	query := r.DB.Table("user_rewards AS user_reward").
		Select(
			"user_reward.id AS id, "+
				"COALESCE(NULLIF(account.nickname, ''), account.user_no) AS nickname, "+
				"item.item_code AS item_code, item.name AS reward_name, "+
				"COALESCE(user_reward.granted_at, user_reward.created_at) AS granted_at",
		).
		Joins("JOIN users AS account ON account.id=user_reward.user_id").
		Joins("JOIN reward_items AS item ON item.id=user_reward.reward_item_id").
		Joins("JOIN activities AS activity ON activity.id=user_reward.activity_id").
		Where("user_reward.status IN (1,2)").
		Where("item.item_code IN ?", itemCodes).
		Where("account.status=1 AND account.deleted_at IS NULL").
		Where("activity.status=2 AND activity.starts_at<=? AND activity.ends_at>? AND activity.deleted_at IS NULL", now, now).
		Where("COALESCE(user_reward.granted_at, user_reward.created_at)>=?", since)
	if afterID > 0 {
		query = query.Where("user_reward.id>?", afterID)
	}

	order := "user_reward.id ASC"
	if latest {
		order = "user_reward.id DESC"
	}
	var rows []RewardAnnouncementRow
	if err := query.Order(order).Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if latest {
		for left, right := 0, len(rows)-1; left < right; left, right = left+1, right-1 {
			rows[left], rows[right] = rows[right], rows[left]
		}
	}
	return rows, nil
}

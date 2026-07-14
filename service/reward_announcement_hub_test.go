package service

import (
	"flower-lottery-backend/repository"
	"testing"
	"time"
)

func TestRewardAnnouncementFromRow(t *testing.T) {
	grantedAt := time.Date(2026, time.July, 14, 18, 0, 0, 0, time.Local)
	tests := []struct {
		name     string
		row      repository.RewardAnnouncementRow
		wantName string
		wantUser string
	}{
		{
			name:     "moon journey",
			row:      repository.RewardAnnouncementRow{ID: 1, Nickname: "小花", ItemCode: "1205251", RewardName: "旧名称", GrantedAt: grantedAt},
			wantName: "月亮游记戒指",
			wantUser: "小花",
		},
		{
			name:     "love flutter",
			row:      repository.RewardAnnouncementRow{ID: 2, Nickname: "小愿", ItemCode: "1205470", RewardName: "旧名称", GrantedAt: grantedAt},
			wantName: "爱意翩跹戒指",
			wantUser: "小愿",
		},
		{
			name:     "true love variants share one family name",
			row:      repository.RewardAnnouncementRow{ID: 3, Nickname: "小奇", ItemCode: "1207753", RewardName: "真爱无敌·铭文版戒指", GrantedAt: grantedAt},
			wantName: "真爱无敌戒指",
			wantUser: "小奇",
		},
		{
			name:     "blank nickname uses public fallback",
			row:      repository.RewardAnnouncementRow{ID: 4, Nickname: "  ", ItemCode: "1207751", GrantedAt: grantedAt},
			wantName: "真爱无敌戒指",
			wantUser: "神秘用户",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := rewardAnnouncementFromRow(test.row)
			if got.RewardName != test.wantName {
				t.Fatalf("RewardName = %q, want %q", got.RewardName, test.wantName)
			}
			if got.Nickname != test.wantUser {
				t.Fatalf("Nickname = %q, want %q", got.Nickname, test.wantUser)
			}
			if got.ID != test.row.ID || !got.GrantedAt.Equal(grantedAt) {
				t.Fatalf("announcement identity changed: %+v", got)
			}
		})
	}
}

func TestRewardAnnouncementSubscribeExcludesExpiredCache(t *testing.T) {
	now := time.Now()
	hub := &RewardAnnouncementHub{
		subscribers: make(map[chan RewardAnnouncement]struct{}),
		recent: []RewardAnnouncement{
			{ID: 1, GrantedAt: now.Add(-rewardAnnouncementLookback - time.Second)},
			{ID: 2, GrantedAt: now.Add(-time.Minute)},
		},
	}
	_, recent, unsubscribe := hub.Subscribe(0)
	defer unsubscribe()
	if len(recent) != 1 || recent[0].ID != 2 {
		t.Fatalf("recent announcements = %+v, want only ID 2", recent)
	}
}

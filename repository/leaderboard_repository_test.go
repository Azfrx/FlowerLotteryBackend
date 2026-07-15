package repository

import (
	"flower-lottery-backend/model"
	"testing"
	"time"
)

func testLeaderboardActivity() model.Activity {
	startsAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	return model.Activity{
		ID:                   1,
		Status:               2,
		StartsAt:             startsAt,
		EndsAt:               startsAt.Add(48 * time.Hour),
		LeaderboardFreezesAt: startsAt.Add(36 * time.Hour),
	}
}

func TestLeaderboardWritableAt(t *testing.T) {
	activity := testLeaderboardActivity()
	if !leaderboardWritableAt(&activity, activity.StartsAt.Add(time.Hour)) {
		t.Fatal("active leaderboard should accept scores before its freeze time")
	}
	if leaderboardWritableAt(&activity, activity.LeaderboardFreezesAt) {
		t.Fatal("leaderboard should reject scores at its freeze time")
	}
	activity.Status = 3
	if leaderboardWritableAt(&activity, activity.StartsAt.Add(time.Hour)) {
		t.Fatal("ended activity should never accept leaderboard scores")
	}
}

func TestActivityPlayableAt(t *testing.T) {
	activity := testLeaderboardActivity()
	if !ActivityPlayableAt(&activity, activity.StartsAt) {
		t.Fatal("activity should be playable at its start time")
	}
	if ActivityPlayableAt(&activity, activity.EndsAt) {
		t.Fatal("activity should be read-only at its end time")
	}
	activity.Status = 3
	if ActivityPlayableAt(&activity, activity.StartsAt.Add(time.Hour)) {
		t.Fatal("manually ended activity should be read-only")
	}
}

func TestLeaderboardShouldFreezeAt(t *testing.T) {
	activity := testLeaderboardActivity()
	if leaderboardShouldFreezeAt(&activity, activity.LeaderboardFreezesAt.Add(-time.Millisecond)) {
		t.Fatal("leaderboard froze before its configured cutoff")
	}
	if !leaderboardShouldFreezeAt(&activity, activity.LeaderboardFreezesAt) {
		t.Fatal("leaderboard did not freeze at its configured cutoff")
	}
	activity.Status = 3
	if !leaderboardShouldFreezeAt(&activity, activity.StartsAt.Add(time.Hour)) {
		t.Fatal("manually ended activity should freeze immediately")
	}
}

func TestSnapshotLeaderboardEntry(t *testing.T) {
	frozenAt := time.Date(2026, 7, 2, 12, 0, 0, 0, time.Local)
	snapshot := model.LeaderboardSnapshot{
		ID:         9,
		ActivityID: 3,
		UserID:     7,
		RankNo:     2,
		Score:      880,
		ReachedAt:  frozenAt.Add(-time.Hour),
		FrozenAt:   frozenAt,
		User:       model.User{ID: 7, Nickname: "test"},
	}
	entry := snapshotLeaderboardEntry(snapshot)
	if entry.ID != snapshot.ID || entry.Score != snapshot.Score || entry.User.ID != snapshot.User.ID {
		t.Fatalf("snapshot entry fields were not preserved: %+v", entry)
	}
	if !entry.UpdatedAt.Equal(frozenAt) {
		t.Fatalf("snapshot entry updated_at = %v, want frozen_at %v", entry.UpdatedAt, frozenAt)
	}
}

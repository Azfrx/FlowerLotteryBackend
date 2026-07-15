package service

import (
	"testing"
	"time"
)

func TestPoolConfigHubPublishesAndReplaysLatestPoolUpdates(t *testing.T) {
	hub := NewPoolConfigHub()
	updates, recent, unsubscribe := hub.Subscribe()
	defer unsubscribe()
	if len(recent) != 0 {
		t.Fatalf("initial recent updates = %+v, want none", recent)
	}

	published := hub.Publish(PoolConfigUpdate{
		PoolID: 1, PoolCode: "day", PetalCostPerDraw: 8,
		CoinValuePerDraw: 480, VersionNo: 2,
	})
	if published.ID == 0 || published.PublishedAt.IsZero() {
		t.Fatalf("published update is missing identity: %+v", published)
	}

	select {
	case got := <-updates:
		if got != published {
			t.Fatalf("subscriber update = %+v, want %+v", got, published)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for pool config update")
	}

	hub.Publish(PoolConfigUpdate{
		PoolID: 2, PoolCode: "night", PetalCostPerDraw: 120,
		CoinValuePerDraw: 7200, VersionNo: 3,
	})
	latestDay := hub.Publish(PoolConfigUpdate{
		PoolID: 1, PoolCode: "day", PetalCostPerDraw: 9,
		CoinValuePerDraw: 540, VersionNo: 4,
	})
	_, replay, stopReplay := hub.Subscribe()
	defer stopReplay()
	if len(replay) != 2 {
		t.Fatalf("replayed updates = %+v, want latest update for both pools", replay)
	}
	if replay[1] != latestDay {
		t.Fatalf("latest day update = %+v, want %+v", replay[1], latestDay)
	}
}

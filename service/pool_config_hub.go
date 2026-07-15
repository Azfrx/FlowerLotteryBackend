package service

import (
	"sort"
	"sync"
	"time"
)

const poolConfigSubscriberBuffer = 16

type PoolConfigUpdate struct {
	ID               uint64    `json:"id"`
	PoolID           uint64    `json:"pool_id"`
	PoolCode         string    `json:"pool_code"`
	PetalCostPerDraw uint64    `json:"petal_cost_per_draw"`
	CoinValuePerDraw uint64    `json:"coin_value_per_draw"`
	VersionNo        uint      `json:"version_no"`
	PublishedAt      time.Time `json:"published_at"`
}

type PoolConfigHub struct {
	mu          sync.RWMutex
	nextID      uint64
	subscribers map[chan PoolConfigUpdate]struct{}
	latest      map[string]PoolConfigUpdate
}

func NewPoolConfigHub() *PoolConfigHub {
	return &PoolConfigHub{
		subscribers: make(map[chan PoolConfigUpdate]struct{}),
		latest:      make(map[string]PoolConfigUpdate),
	}
}

func (h *PoolConfigHub) Publish(update PoolConfigUpdate) PoolConfigUpdate {
	h.mu.Lock()
	h.nextID++
	update.ID = h.nextID
	if update.PublishedAt.IsZero() {
		update.PublishedAt = time.Now()
	}
	h.latest[update.PoolCode] = update
	subscribers := make([]chan PoolConfigUpdate, 0, len(h.subscribers))
	for subscriber := range h.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	h.mu.Unlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- update:
		default:
		}
	}
	return update
}

func (h *PoolConfigHub) Subscribe() (<-chan PoolConfigUpdate, []PoolConfigUpdate, func()) {
	updates := make(chan PoolConfigUpdate, poolConfigSubscriberBuffer)
	h.mu.Lock()
	h.subscribers[updates] = struct{}{}
	recent := make([]PoolConfigUpdate, 0, len(h.latest))
	for _, update := range h.latest {
		recent = append(recent, update)
	}
	h.mu.Unlock()
	sort.Slice(recent, func(i, j int) bool { return recent[i].ID < recent[j].ID })

	unsubscribe := func() {
		h.mu.Lock()
		delete(h.subscribers, updates)
		h.mu.Unlock()
	}
	return updates, recent, unsubscribe
}

package service

import (
	"strings"
	"sync"
	"time"

	"flower-lottery-backend/repository"

	"go.uber.org/zap"
)

const (
	rewardAnnouncementLookback     = 10 * time.Minute
	rewardAnnouncementRecentLimit  = 5
	rewardAnnouncementPollInterval = time.Second
	rewardAnnouncementBatchLimit   = 100
)

var rewardAnnouncementItemCodes = []string{
	"1205251",
	"1205470",
	"1207751",
	"1207752",
	"1207753",
}

type RewardAnnouncement struct {
	ID         uint64    `json:"id"`
	Nickname   string    `json:"nickname"`
	ItemCode   string    `json:"item_code"`
	RewardName string    `json:"reward_name"`
	GrantedAt  time.Time `json:"granted_at"`
}

type RewardAnnouncementHub struct {
	repo *repository.GameRepository
	log  *zap.Logger

	mu           sync.RWMutex
	subscribers  map[chan RewardAnnouncement]struct{}
	recent       []RewardAnnouncement
	lastID       uint64
	initialized  bool
	lastErrorLog time.Time
}

func NewRewardAnnouncementHub(repo *repository.GameRepository, log *zap.Logger) *RewardAnnouncementHub {
	if log == nil {
		log = zap.NewNop()
	}
	hub := &RewardAnnouncementHub{
		repo:        repo,
		log:         log,
		subscribers: make(map[chan RewardAnnouncement]struct{}),
	}
	if err := hub.refresh(true, false); err != nil {
		hub.reportError(err)
	}
	go hub.run()
	return hub
}

func (h *RewardAnnouncementHub) Subscribe(afterID uint64) (<-chan RewardAnnouncement, []RewardAnnouncement, func()) {
	updates := make(chan RewardAnnouncement, rewardAnnouncementBatchLimit)
	cutoff := time.Now().Add(-rewardAnnouncementLookback)
	h.mu.Lock()
	h.subscribers[updates] = struct{}{}
	recent := make([]RewardAnnouncement, 0, len(h.recent))
	for _, announcement := range h.recent {
		if announcement.ID > afterID && !announcement.GrantedAt.Before(cutoff) {
			recent = append(recent, announcement)
		}
	}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		delete(h.subscribers, updates)
		h.mu.Unlock()
	}
	return updates, recent, unsubscribe
}

func (h *RewardAnnouncementHub) run() {
	ticker := time.NewTicker(rewardAnnouncementPollInterval)
	defer ticker.Stop()
	for range ticker.C {
		h.mu.RLock()
		latest := !h.initialized
		h.mu.RUnlock()
		if err := h.refresh(latest, true); err != nil {
			h.reportError(err)
		}
	}
}

func (h *RewardAnnouncementHub) refresh(latest, broadcast bool) error {
	h.mu.RLock()
	afterID := h.lastID
	h.mu.RUnlock()
	if latest {
		afterID = 0
	}

	limit := rewardAnnouncementBatchLimit
	if latest {
		limit = rewardAnnouncementRecentLimit
	}
	cutoff := time.Now().Add(-rewardAnnouncementLookback)
	rows, err := h.repo.RewardAnnouncements(
		afterID,
		cutoff,
		rewardAnnouncementItemCodes,
		limit,
		latest,
	)
	if err != nil {
		return err
	}

	announcements := make([]RewardAnnouncement, 0, len(rows))
	h.mu.Lock()
	retained := h.recent[:0]
	for _, announcement := range h.recent {
		if !announcement.GrantedAt.Before(cutoff) {
			retained = append(retained, announcement)
		}
	}
	h.recent = retained
	for _, row := range rows {
		if row.ID <= h.lastID {
			continue
		}
		announcement := rewardAnnouncementFromRow(row)
		h.lastID = announcement.ID
		h.recent = append(h.recent, announcement)
		if len(h.recent) > rewardAnnouncementRecentLimit {
			h.recent = append([]RewardAnnouncement(nil), h.recent[len(h.recent)-rewardAnnouncementRecentLimit:]...)
		}
		announcements = append(announcements, announcement)
	}
	h.initialized = true
	subscribers := make([]chan RewardAnnouncement, 0, len(h.subscribers))
	if broadcast && len(announcements) > 0 {
		for subscriber := range h.subscribers {
			subscribers = append(subscribers, subscriber)
		}
	}
	h.mu.Unlock()

	for _, announcement := range announcements {
		for _, subscriber := range subscribers {
			select {
			case subscriber <- announcement:
			default:
			}
		}
	}
	return nil
}

func (h *RewardAnnouncementHub) reportError(err error) {
	now := time.Now()
	if now.Sub(h.lastErrorLog) < 30*time.Second {
		return
	}
	h.lastErrorLog = now
	h.log.Warn("reward announcement polling failed", zap.Error(err))
}

func rewardAnnouncementFromRow(row repository.RewardAnnouncementRow) RewardAnnouncement {
	nickname := strings.TrimSpace(row.Nickname)
	if nickname == "" {
		nickname = "神秘用户"
	}
	return RewardAnnouncement{
		ID:         row.ID,
		Nickname:   nickname,
		ItemCode:   row.ItemCode,
		RewardName: rewardAnnouncementName(row.ItemCode, row.RewardName),
		GrantedAt:  row.GrantedAt,
	}
}

func rewardAnnouncementName(itemCode, fallback string) string {
	switch itemCode {
	case "1205251":
		return "月亮游记戒指"
	case "1205470":
		return "爱意翩跹戒指"
	case "1207751", "1207752", "1207753":
		return "真爱无敌戒指"
	default:
		return fallback
	}
}

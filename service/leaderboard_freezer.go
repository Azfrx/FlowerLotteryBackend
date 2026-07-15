package service

import (
	"context"
	"flower-lottery-backend/repository"
	"go.uber.org/zap"
	"time"
)

func StartLeaderboardFreezer(
	repo *repository.GameRepository,
	log *zap.Logger,
	interval time.Duration,
) context.CancelFunc {
	if interval <= 0 {
		interval = time.Minute
	}
	ctx, cancel := context.WithCancel(context.Background())
	run := func() {
		if err := repo.FreezeDueLeaderboards(time.Now()); err != nil {
			log.Warn("leaderboard snapshot freeze failed", zap.Error(err))
		}
	}
	go func() {
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
	return cancel
}

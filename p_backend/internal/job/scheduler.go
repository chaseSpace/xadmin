package job

import (
	"context"
	"monorepo/internal/repo/system"
	"monorepo/internal/support/ipblacklist"
	"monorepo/pkg/logger"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithLocation(time.Local)),
	}
}

func (s *Scheduler) Start() error {
	_, _ = s.cron.AddFunc("@every 30s", flushIPBlacklistHitCounts)
	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) {
	flushIPBlacklistHitCounts()
	if ctx == nil {
		ctx = context.Background()
	}
	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
		logger.Info("Scheduler stopped")
	case <-ctx.Done():
		logger.Warn("Scheduler stop timeout", zap.Error(ctx.Err()))
	}
}

func flushIPBlacklistHitCounts() {
	deltas := ipblacklist.DefaultStore().FlushHitDeltas()
	if len(deltas) == 0 {
		return
	}
	repo := system.NewRepo()
	if err := repo.IncrementHitCounts(context.Background(), deltas); err != nil {
		logger.Error("Failed to flush IP blacklist hit counts", zap.Error(err))
	}
}

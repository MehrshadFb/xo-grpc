package cleanup

import (
	"context"
	"log/slog"
	"time"
)

type Policy struct {
	Waiting    time.Duration
	InProgress time.Duration
	Ended      time.Duration
}

var DefaultPolicy = Policy{
	Waiting:    time.Hour,
	InProgress: 24 * time.Hour,
	Ended:      7 * 24 * time.Hour,
}

type StaleGameDeleter interface {
	DeleteStale(ctx context.Context, waitingBefore, inProgressBefore, endedBefore time.Time) (int64, error)
}

type Service struct {
	games  StaleGameDeleter
	policy Policy
	now    func() time.Time
}

func NewService(games StaleGameDeleter, policy Policy) *Service {
	return &Service{
		games:  games,
		policy: policy,
		now:    time.Now,
	}
}

func (s *Service) RunOnce(ctx context.Context) (int64, error) {
	now := s.now()

	return s.games.DeleteStale(
		ctx,
		now.Add(-s.policy.Waiting),
		now.Add(-s.policy.InProgress),
		now.Add(-s.policy.Ended),
	)
}

func (s *Service) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		deleted, err := s.RunOnce(ctx)
		if err != nil {
			slog.Error("stale game cleanup failed", "error", err)
		} else if deleted > 0 {
			slog.Info("deleted stale games", "count", deleted)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

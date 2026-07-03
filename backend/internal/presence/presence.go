package presence

import (
	"context"
	"time"

	"backend/internal/config"
	"backend/internal/domain"

	"go.uber.org/fx"
)

var Module = fx.Module("presence", fx.Provide(NewStore), fx.Provide(NewService))

type Store interface {
	Touch(ctx context.Context, serverID string, seenAt time.Time) error
	LastSeen(ctx context.Context, serverID string) (time.Time, bool, error)
}

type Service struct {
	cfg   config.Config
	store Store
}

func NewService(cfg config.Config, store Store) *Service {
	return &Service{cfg: cfg, store: store}
}

func (s *Service) Touch(ctx context.Context, serverID string, seenAt time.Time) error {
	if serverID == "" {
		return nil
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}
	return s.store.Touch(ctx, serverID, seenAt.UTC())
}

func (s *Service) ApplySummary(ctx context.Context, summary domain.ServerSummary, now time.Time) domain.ServerSummary {
	seenAt, ok, err := s.store.LastSeen(ctx, summary.ID)
	if err == nil && ok && seenAt.After(summary.LastSeen) {
		summary.LastSeen = seenAt
	}
	summary.Status = statusFor(summary.LastSeen, now, s.cfg.State.OfflineAfter)
	return summary
}

func (s *Service) ApplySummaries(ctx context.Context, summaries []domain.ServerSummary, now time.Time) []domain.ServerSummary {
	out := make([]domain.ServerSummary, len(summaries))
	for i, summary := range summaries {
		out[i] = s.ApplySummary(ctx, summary, now)
	}
	return out
}

func statusFor(lastSeen time.Time, now time.Time, offlineAfter time.Duration) string {
	if lastSeen.IsZero() {
		return "unknown"
	}
	if now.Sub(lastSeen) > offlineAfter {
		return "offline"
	}
	return "online"
}

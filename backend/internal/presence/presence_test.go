package presence

import (
	"context"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/domain"
)

func TestServiceAppliesOnlinePresenceOverOldSnapshot(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute}}
	store := NewMemoryStore()
	service := NewService(cfg, store)
	now := time.Date(2026, 7, 3, 9, 0, 0, 0, time.UTC)
	oldSeen := now.Add(-2 * time.Minute)
	if err := service.Touch(context.Background(), "devbox", now.Add(-10*time.Second)); err != nil {
		t.Fatalf("Touch() error = %v", err)
	}

	summary := service.ApplySummary(context.Background(), domain.ServerSummary{ID: "devbox", LastSeen: oldSeen, Status: "offline"}, now)
	if summary.Status != "online" {
		t.Fatalf("status = %q", summary.Status)
	}
	if !summary.LastSeen.Equal(now.Add(-10 * time.Second)) {
		t.Fatalf("last_seen = %s", summary.LastSeen)
	}
}

func TestServiceFallsBackToSnapshotLastSeen(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute}}
	service := NewService(cfg, NewMemoryStore())
	now := time.Date(2026, 7, 3, 9, 0, 0, 0, time.UTC)
	summary := service.ApplySummary(context.Background(), domain.ServerSummary{ID: "devbox", LastSeen: now.Add(-2 * time.Minute)}, now)
	if summary.Status != "offline" {
		t.Fatalf("status = %q", summary.Status)
	}
}

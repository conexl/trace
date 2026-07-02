package ingest

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"backend/internal/alerts"
	"backend/internal/config"
	"backend/internal/store"
)

func TestServiceIngestsSnapshotEnvelope(t *testing.T) {
	memory := store.NewMemoryStore(config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}})
	service := newTestService(memory)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","host":{"hostname":"arch","platform":"linux"},"system":{"cpu_percent":7,"memory":{"used_percent":30}},"network":{"public_ip":"203.0.113.1"},"collected_at":"2026-07-02T09:00:00Z"}]}`)

	result, err := service.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("Accepted = %d", result.Accepted)
	}
	state, err := memory.GetServer(context.Background(), "devbox", time.Date(2026, 7, 2, 9, 0, 1, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetServer() error = %v", err)
	}
	if state.Summary.PublicIP != "203.0.113.1" || state.Summary.CPUPercent != 7 {
		t.Fatalf("state = %#v", state)
	}
}

func TestServiceRejectsAnonymousSnapshot(t *testing.T) {
	memory := store.NewMemoryStore(config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}})
	service := newTestService(memory)
	payload, _ := json.Marshal(map[string]any{"snapshots": []any{map[string]any{}}})
	if _, err := service.Ingest(context.Background(), payload); err == nil {
		t.Fatal("Ingest() expected validation error")
	}
}

func newTestService(memory *store.MemoryStore) *Service {
	notifier := alerts.NewMemoryNotifier(config.Config{Alerts: config.AlertsConfig{MemoryLimit: 10}})
	dispatcher := alerts.NewDispatcher(alerts.DispatcherParams{Notifiers: []alerts.Notifier{notifier}})
	return NewService(memory, alerts.NewEvaluator(), dispatcher)
}

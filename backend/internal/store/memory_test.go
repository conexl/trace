package store

import (
	"context"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/domain"
)

func TestMemoryStoreUpsertAndListServers(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 2}}
	store := NewMemoryStore(cfg)
	now := time.Now()
	_, err := store.UpsertSnapshot(context.Background(), domain.AgentSnapshot{
		AgentName: "devbox",
		Host:      domain.HostSnapshot{Hostname: "arch", Platform: "linux"},
		Network:   domain.NetworkSnapshot{PublicIP: "203.0.113.10"},
		System:    domain.SystemSnapshot{CPUPercent: 12.5, Memory: domain.Memory{UsedPercent: 40}},
		Events:    []domain.AgentEvent{{Type: "process.down", Severity: "critical", Timestamp: now}},
		Collected: now,
	})
	if err != nil {
		t.Fatalf("UpsertSnapshot() error = %v", err)
	}

	summaries, err := store.ListServers(context.Background(), now)
	if err != nil {
		t.Fatalf("ListServers() error = %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("summaries = %#v", summaries)
	}
	if summaries[0].ID != "devbox" || summaries[0].Status != "online" || summaries[0].EventCount != 1 {
		t.Fatalf("summary = %#v", summaries[0])
	}
}

func TestMemoryStoreMarksOffline(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}}
	store := NewMemoryStore(cfg)
	lastSeen := time.Now().Add(-2 * time.Minute)
	_, err := store.UpsertSnapshot(context.Background(), domain.AgentSnapshot{AgentName: "devbox", Collected: lastSeen})
	if err != nil {
		t.Fatalf("UpsertSnapshot() error = %v", err)
	}

	state, err := store.GetServer(context.Background(), "devbox", time.Now())
	if err != nil {
		t.Fatalf("GetServer() error = %v", err)
	}
	if state.Summary.Status != "offline" {
		t.Fatalf("status = %q", state.Summary.Status)
	}
}

func TestMemoryStoreCapsEvents(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 2}}
	store := NewMemoryStore(cfg)
	for _, eventType := range []string{"one", "two", "three"} {
		_, err := store.UpsertSnapshot(context.Background(), domain.AgentSnapshot{AgentName: "devbox", Events: []domain.AgentEvent{{Type: eventType}}, Collected: time.Now()})
		if err != nil {
			t.Fatalf("UpsertSnapshot() error = %v", err)
		}
	}
	state, err := store.GetServer(context.Background(), "devbox", time.Now())
	if err != nil {
		t.Fatalf("GetServer() error = %v", err)
	}
	if len(state.Events) != 2 || state.Events[0].Type != "two" || state.Events[1].Type != "three" {
		t.Fatalf("events = %#v", state.Events)
	}
}

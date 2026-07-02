package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"backend/internal/config"
	"backend/internal/domain"
)

type MemoryStore struct {
	mu     sync.RWMutex
	cfg    config.Config
	states map[string]domain.ServerState
}

func NewMemoryStore(cfg config.Config) *MemoryStore {
	return &MemoryStore{cfg: cfg, states: make(map[string]domain.ServerState)}
}

func (s *MemoryStore) UpsertSnapshot(ctx context.Context, snapshot domain.AgentSnapshot) (domain.ServerState, error) {
	select {
	case <-ctx.Done():
		return domain.ServerState{}, ctx.Err()
	default:
	}
	state := stateFromSnapshot(snapshot, s.cfg, time.Now())
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.states[state.Summary.ID]; ok {
		state.Events = append(existing.Events, snapshot.Events...)
		if len(state.Events) > s.cfg.State.MaxEvents {
			state.Events = state.Events[len(state.Events)-s.cfg.State.MaxEvents:]
		}
	}
	s.states[state.Summary.ID] = state
	return state, nil
}

func (s *MemoryStore) ListServers(ctx context.Context, now time.Time) ([]domain.ServerSummary, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	summaries := make([]domain.ServerSummary, 0, len(s.states))
	for _, state := range s.states {
		summary := state.Summary
		summary.Status = statusFor(summary.LastSeen, now, s.cfg.State.OfflineAfter)
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})
	return summaries, nil
}

func (s *MemoryStore) GetServer(ctx context.Context, id string, now time.Time) (domain.ServerState, error) {
	select {
	case <-ctx.Done():
		return domain.ServerState{}, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[id]
	if !ok {
		return domain.ServerState{}, ErrNotFound{ID: id}
	}
	state.Summary.Status = statusFor(state.Summary.LastSeen, now, s.cfg.State.OfflineAfter)
	return state, nil
}

func stateFromSnapshot(snapshot domain.AgentSnapshot, cfg config.Config, now time.Time) domain.ServerState {
	if snapshot.Collected.IsZero() {
		snapshot.Collected = now
	}
	id := serverID(snapshot)
	state := domain.ServerState{
		Summary: domain.ServerSummary{
			ID:           id,
			Name:         snapshot.AgentName,
			Hostname:     snapshot.Host.Hostname,
			Platform:     snapshot.Host.Platform,
			PublicIP:     snapshot.Network.PublicIP,
			Status:       statusFor(snapshot.Collected, now, cfg.State.OfflineAfter),
			LastSeen:     snapshot.Collected,
			CPUPercent:   snapshot.System.CPUPercent,
			MemoryUsed:   snapshot.System.Memory.UsedPercent,
			ProcessCount: len(snapshot.Processes),
			EventCount:   len(snapshot.Events),
		},
		Snapshot: snapshot,
		Events:   append([]domain.AgentEvent(nil), snapshot.Events...),
	}
	return state
}

func serverID(snapshot domain.AgentSnapshot) string {
	if snapshot.AgentName != "" {
		return snapshot.AgentName
	}
	if snapshot.Host.Hostname != "" {
		return snapshot.Host.Hostname
	}
	return "unknown"
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

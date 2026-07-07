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
	metrics map[string][]domain.Metric
}

func NewMemoryStore(cfg config.Config) *MemoryStore {
	return &MemoryStore{cfg: cfg, states: make(map[string]domain.ServerState), metrics: make(map[string][]domain.Metric)}
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

func (s *MemoryStore) SaveMetric(ctx context.Context, metric domain.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[metric.ServerID] = append(s.metrics[metric.ServerID], metric)
	// Keep last 1000 metrics
	if len(s.metrics[metric.ServerID]) > 1000 {
		s.metrics[metric.ServerID] = s.metrics[metric.ServerID][len(s.metrics[metric.ServerID])-1000:]
	}
	return nil
}

func (s *MemoryStore) GetMetrics(ctx context.Context, serverID string, from, to time.Time) ([]domain.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.metrics[serverID]
	out := make([]domain.Metric, 0)
	for _, m := range all {
		if (m.Timestamp.After(from) || m.Timestamp.Equal(from)) && (m.Timestamp.Before(to) || m.Timestamp.Equal(to)) {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *MemoryStore) UpdateDesiredRevision(ctx context.Context, serverID string, revision int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if state, ok := s.states[serverID]; ok {
		state.Summary.DesiredConfigRevision = revision
		s.states[serverID] = state
	}
	return nil
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
			Version:      snapshot.Host.Version,
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

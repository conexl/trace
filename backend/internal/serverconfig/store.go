package serverconfig

import (
	"context"
	"errors"
	"sync"

	"backend/internal/domain"
)

var ErrNotFound = errors.New("server config not found")

type Store interface {
	Get(ctx context.Context, serverID string) (domain.AgentDesiredConfig, error)
	Set(ctx context.Context, serverID string, cfg domain.AgentDesiredConfig) error
}

type MemoryStore struct {
	mu     sync.RWMutex
	cfgs   map[string]domain.AgentDesiredConfig
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{cfgs: make(map[string]domain.AgentDesiredConfig)}
}

func (s *MemoryStore) Get(ctx context.Context, serverID string) (domain.AgentDesiredConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.cfgs[serverID]
	if !ok {
		return domain.AgentDesiredConfig{}, ErrNotFound
	}
	return cfg, nil
}

func (s *MemoryStore) Set(ctx context.Context, serverID string, cfg domain.AgentDesiredConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.cfgs[serverID]
	cfg.Revision = current.Revision + 1
	s.cfgs[serverID] = cfg
	return nil
}

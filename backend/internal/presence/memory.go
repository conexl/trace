package presence

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu       sync.RWMutex
	lastSeen map[string]time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{lastSeen: make(map[string]time.Time)}
}

func (s *MemoryStore) Touch(ctx context.Context, serverID string, seenAt time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSeen[serverID] = seenAt.UTC()
	return nil
}

func (s *MemoryStore) LastSeen(ctx context.Context, serverID string) (time.Time, bool, error) {
	select {
	case <-ctx.Done():
		return time.Time{}, false, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	seenAt, ok := s.lastSeen[serverID]
	return seenAt, ok, nil
}

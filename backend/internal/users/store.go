package users

import (
	"context"
	"errors"
	"sync"

	"backend/internal/domain"
)

var ErrExists = errors.New("user already exists")

var ErrNotFound = errors.New("user not found")

type Store interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Count(ctx context.Context) (int, error)
}

type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]domain.User)}
}

func (s *MemoryStore) Create(ctx context.Context, user domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.Email]; ok {
		return ErrExists
	}
	s.users[user.Email] = user
	return nil
}

func (s *MemoryStore) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[email]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	return user, nil
}

func (s *MemoryStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users), nil
}

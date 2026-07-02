package store

import (
	"context"
	"time"

	"backend/internal/domain"
)

type Store interface {
	UpsertSnapshot(ctx context.Context, snapshot domain.AgentSnapshot) (domain.ServerState, error)
	ListServers(ctx context.Context, now time.Time) ([]domain.ServerSummary, error)
	GetServer(ctx context.Context, id string, now time.Time) (domain.ServerState, error)
}

type ErrNotFound struct {
	ID string
}

func (e ErrNotFound) Error() string {
	return "server not found: " + e.ID
}

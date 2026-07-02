package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/domain"
	"backend/internal/store"

	"go.uber.org/fx"
)

var Module = fx.Module("ingest", fx.Provide(NewService))

type Service struct {
	store store.Store
}

type Result struct {
	Accepted int                  `json:"accepted"`
	States   []domain.ServerState `json:"states,omitempty"`
}

func NewService(store store.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Ingest(ctx context.Context, payload []byte) (Result, error) {
	var raw struct {
		Snapshots []json.RawMessage `json:"snapshots"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return Result{}, fmt.Errorf("decode snapshot envelope: %w", err)
	}
	if len(raw.Snapshots) == 0 {
		return Result{}, fmt.Errorf("snapshot envelope is empty")
	}
	states := make([]domain.ServerState, 0, len(raw.Snapshots))
	for _, item := range raw.Snapshots {
		var snapshot domain.AgentSnapshot
		if err := json.Unmarshal(item, &snapshot); err != nil {
			return Result{}, fmt.Errorf("decode snapshot: %w", err)
		}
		snapshot.Raw = append([]byte(nil), item...)
		if snapshot.AgentName == "" && snapshot.Host.Hostname == "" {
			return Result{}, fmt.Errorf("snapshot must include agent_name or host.hostname")
		}
		if snapshot.Collected.IsZero() {
			snapshot.Collected = time.Now()
		}
		state, err := s.store.UpsertSnapshot(ctx, snapshot)
		if err != nil {
			return Result{}, err
		}
		states = append(states, state)
	}
	return Result{Accepted: len(states), States: states}, nil
}

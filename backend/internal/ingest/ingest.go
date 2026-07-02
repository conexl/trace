package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/alerts"
	"backend/internal/domain"
	"backend/internal/store"

	"go.uber.org/fx"
)

var Module = fx.Module("ingest", fx.Provide(NewService))

type Service struct {
	store      store.Store
	evaluator  *alerts.Evaluator
	dispatcher *alerts.Dispatcher
}

type Result struct {
	Accepted int                  `json:"accepted"`
	States   []domain.ServerState `json:"states,omitempty"`
	Alerts   []alerts.Alert       `json:"alerts,omitempty"`
}

func NewService(store store.Store, evaluator *alerts.Evaluator, dispatcher *alerts.Dispatcher) *Service {
	return &Service{store: store, evaluator: evaluator, dispatcher: dispatcher}
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
	allAlerts := make([]alerts.Alert, 0)
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
		snapshotAlerts := s.evaluator.Evaluate(state)
		if err := s.dispatcher.Dispatch(ctx, snapshotAlerts); err != nil {
			return Result{}, err
		}
		allAlerts = append(allAlerts, snapshotAlerts...)
	}
	return Result{Accepted: len(states), States: states, Alerts: allAlerts}, nil
}

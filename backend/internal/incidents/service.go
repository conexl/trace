package incidents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/domain"
	"backend/internal/pubsub"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("incidents",
	fx.Provide(NewMongoStore),
	fx.Provide(NewService),
)

// Service manages incident lifecycle
type Service struct {
	store  Store
	pubsub *pubsub.Service
	logger *zap.Logger
}

// ServiceParams for fx
type ServiceParams struct {
	fx.In
	Store  Store
	Pubsub *pubsub.Service
	Logger *zap.Logger `optional:"true"`
}

func NewService(params ServiceParams) *Service {
	logger := params.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		store:  params.Store,
		pubsub: params.Pubsub,
		logger: logger.Named("incidents"),
	}
}

// ProcessEvent processes agent events and creates/updates incidents
func (s *Service) ProcessEvent(ctx context.Context, serverID string, event domain.AgentEvent) error {
	// Only process process-related events
	if !isProcessEvent(event.Type) {
		return nil
	}

	now := time.Now().UTC()

	// Check for existing open incident
	existing, err := s.store.GetOpen(ctx, serverID, event.Subject)
	if err != nil {
		return fmt.Errorf("get open incident: %w", err)
	}

	if existing != nil {
		// Update existing incident
		return s.updateExistingIncident(ctx, existing, event, now)
	}

	// Create new incident
	incident := CreateFromEvent(serverID, event, now)
	if err := s.store.Save(ctx, *incident); err != nil {
		return fmt.Errorf("save incident: %w", err)
	}

	s.logger.Info("incident created",
		zap.String("incident_id", incident.ID),
		zap.String("server_id", serverID),
		zap.String("service", event.Subject),
	)

	// Publish incident event
	s.publishIncident(ctx, "incident.created", incident)

	return nil
}

func (s *Service) updateExistingIncident(ctx context.Context, incident *Incident, event domain.AgentEvent, now time.Time) error {
	// Add timeline event
	timelineEvent := TimelineEvent{
		ID:        generateEventID(now, len(incident.Timeline)),
		Type:      getTimelineType(event),
		Timestamp: event.Timestamp,
		Title:     formatEventTitle(event),
		Message:   event.Message,
		ExitCode:  event.ExitCode,
	}

	if err := s.store.AddTimelineEvent(ctx, incident.ID, timelineEvent); err != nil {
		return err
	}

	incident.Timeline = append(incident.Timeline, timelineEvent)

	// Auto-resolve incident if service recovered
	if event.Type == "process.up" {
		resolvedAt := event.Timestamp
		if err := s.store.UpdateStatus(ctx, incident.ID, "resolved", &resolvedAt); err != nil {
			return err
		}
		incident.Status = "resolved"
		incident.ResolvedAt = &resolvedAt
		s.publishIncident(ctx, "incident.resolved", incident)
	} else {
		s.publishIncident(ctx, "incident.updated", incident)
	}

	return nil
}

// ResolveIncident marks incident as resolved
func (s *Service) ResolveIncident(ctx context.Context, incidentID string) error {
	now := time.Now().UTC()
	incident, err := s.store.Get(ctx, incidentID)
	if err != nil {
		return err
	}

	// Add resolution event
	timelineEvent := TimelineEvent{
		ID:        generateEventID(now, len(incident.Timeline)),
		Type:      "resolved",
		Timestamp: now,
		Title:     "Service recovered",
		Message:   "Service is now running",
	}

	if err := s.store.AddTimelineEvent(ctx, incidentID, timelineEvent); err != nil {
		return err
	}

	if err := s.store.UpdateStatus(ctx, incidentID, "resolved", &now); err != nil {
		return err
	}

	incident.Status = "resolved"
	incident.ResolvedAt = &now
	s.publishIncident(ctx, "incident.resolved", incident)

	return nil
}

// ExecuteAction executes an incident action
func (s *Service) ExecuteAction(ctx context.Context, incidentID, action, actor string) error {
	incident, err := s.store.Get(ctx, incidentID)
	if err != nil {
		return err
	}

	if incident.Status != "open" {
		return ErrInvalidState
	}

	now := time.Now().UTC()

	// Add action event to timeline
	timelineEvent := TimelineEvent{
		ID:        generateEventID(now, len(incident.Timeline)),
		Type:      "action",
		Timestamp: now,
		Title:     fmt.Sprintf("Action: %s", action),
		Action:    action,
		Actor:     actor,
		Result:    "initiated",
	}

	if err := s.store.AddTimelineEvent(ctx, incidentID, timelineEvent); err != nil {
		return err
	}

	s.publishIncident(ctx, "incident.action", incident)

	return nil
}

// GetIncident retrieves incident by ID
func (s *Service) GetIncident(ctx context.Context, id string) (*Incident, error) {
	return s.store.Get(ctx, id)
}

// ListIncidents lists recent incidents
func (s *Service) ListIncidents(ctx context.Context, serverID string, limit int) ([]Incident, error) {
	return s.store.Recent(ctx, serverID, limit)
}

func (s *Service) publishIncident(ctx context.Context, eventType string, incident *Incident) {
	payload, err := json.Marshal(map[string]any{
		"type": eventType,
		"data": incident,
	})
	if err != nil {
		s.logger.Warn("failed to encode incident event", zap.Error(err))
		return
	}
	_ = s.pubsub.Publish(ctx, "events", payload)
}

func isProcessEvent(eventType string) bool {
	switch eventType {
	case "process.down", "process.restart_failed", "process.restart_suppressed", "process.up":
		return true
	}
	return false
}

func getTimelineType(event domain.AgentEvent) string {
	switch event.Type {
	case "process.up":
		return "resolved"
	case "process.restart_failed", "process.restart_suppressed":
		return "restart"
	default:
		return "crash"
	}
}

func formatEventTitle(event domain.AgentEvent) string {
	switch event.Type {
	case "process.up":
		return "Service recovered"
	case "process.restart_failed":
		return "Restart failed"
	case "process.restart_suppressed":
		return "Restart suppressed (max attempts)"
	default:
		return "Service crashed"
	}
}

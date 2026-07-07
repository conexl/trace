package incidents

import (
	"encoding/json"
	"time"

	"backend/internal/domain"
)

// Incident represents a service failure event with full timeline
type Incident struct {
	ID          string          `json:"id" bson:"_id"`
	ServerID    string          `json:"server_id" bson:"server_id"`
	ServiceName string          `json:"service_name" bson:"service_name"`
	Status      string          `json:"status" bson:"status"`     // open, investigating, resolved, suppressed
	Severity    string          `json:"severity" bson:"severity"` // critical, warning
	Title       string          `json:"title" bson:"title"`
	Summary     string          `json:"summary" bson:"summary"`
	Timeline    []TimelineEvent `json:"timeline" bson:"timeline"`
	CreatedAt   time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" bson:"updated_at"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
}

// TimelineEvent represents a single event in the incident timeline
type TimelineEvent struct {
	ID        string          `json:"id" bson:"id"`
	Type      string          `json:"type" bson:"type"` // crash, restart, action, log, resolved
	Timestamp time.Time       `json:"timestamp" bson:"timestamp"`
	Title     string          `json:"title" bson:"title"`
	Message   string          `json:"message,omitempty" bson:"message,omitempty"`
	ExitCode  int             `json:"exit_code,omitempty" bson:"exit_code,omitempty"`
	Action    string          `json:"action,omitempty" bson:"action,omitempty"`
	Actor     string          `json:"actor,omitempty" bson:"actor,omitempty"`   // who performed action
	Result    string          `json:"result,omitempty" bson:"result,omitempty"` // success, failed
	Metadata  json.RawMessage `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// IncidentAction represents available actions for an incident
type IncidentAction struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	ComingSoon  bool   `json:"coming_soon,omitempty"`
}

// Metrics summarizes incident reliability over a rolling window.
type Metrics struct {
	Window          string                    `json:"window"`
	Total           int                       `json:"total"`
	Open            int                       `json:"open"`
	Resolved        int                       `json:"resolved"`
	Critical        int                       `json:"critical"`
	Warning         int                       `json:"warning"`
	MTTRSeconds     float64                   `json:"mttr_seconds"`
	FrequencyPerDay float64                   `json:"frequency_per_day"`
	ByService       map[string]ServiceMetrics `json:"by_service"`
}

// ServiceMetrics summarizes incident reliability for one service.
type ServiceMetrics struct {
	Total           int     `json:"total"`
	Open            int     `json:"open"`
	Resolved        int     `json:"resolved"`
	Critical        int     `json:"critical"`
	Warning         int     `json:"warning"`
	MTTRSeconds     float64 `json:"mttr_seconds"`
	FrequencyPerDay float64 `json:"frequency_per_day"`
}

// AvailableActions returns list of actions for MVP
func AvailableActions() []IncidentAction {
	return []IncidentAction{
		{
			Name:        "restart",
			Label:       "Restart Service",
			Description: "Restart the failed service",
			Enabled:     true,
		},
		{
			Name:        "disable-watchdog",
			Label:       "Disable Watchdog",
			Description: "Stop automatic restart attempts",
			Enabled:     true,
		},
		{
			Name:        "diagnostics",
			Label:       "Run Diagnostics",
			Description: "Collect diagnostic information",
			Enabled:     false,
			ComingSoon:  true,
		},
		{
			Name:        "rollback-config",
			Label:       "Rollback Config",
			Description: "Restore previous configuration",
			Enabled:     false,
			ComingSoon:  true,
		},
	}
}

// CreateFromEvent creates a new incident from agent event
func CreateFromEvent(serverID string, event domain.AgentEvent, now time.Time) *Incident {
	incident := &Incident{
		ID:          generateID(serverID, event, now),
		ServerID:    serverID,
		ServiceName: event.Subject,
		Status:      "open",
		Severity:    event.Severity,
		Title:       formatTitle(event),
		Summary:     event.Message,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Add initial timeline event
	incident.Timeline = []TimelineEvent{
		{
			ID:        generateEventID(now, 0),
			Type:      "crash",
			Timestamp: event.Timestamp,
			Title:     "Service crashed",
			Message:   event.Message,
			ExitCode:  event.ExitCode,
		},
	}

	// Add restart attempt if watchdog tried
	if event.Action == "restart" {
		incident.Timeline = append(incident.Timeline, TimelineEvent{
			ID:        generateEventID(now, 1),
			Type:      "restart",
			Timestamp: event.Timestamp.Add(100 * time.Millisecond),
			Title:     "Watchdog attempted restart",
			Action:    "restart",
			Result:    "attempted",
		})
	}

	return incident
}

func generateID(serverID string, event domain.AgentEvent, now time.Time) string {
	return serverID + ":" + event.Subject + ":" + now.Format("20060102-150405")
}

func generateEventID(base time.Time, seq int) string {
	return base.Format("20060102-150405") + "-" + string(rune('a'+seq))
}

func formatTitle(event domain.AgentEvent) string {
	switch event.Type {
	case "process.down":
		return event.Subject + " crashed"
	case "process.restart_failed":
		return event.Subject + " restart failed"
	case "process.restart_suppressed":
		return event.Subject + " restart suppressed (max attempts)"
	default:
		return event.Subject + " incident"
	}
}

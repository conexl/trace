package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/internal/domain"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("alerts", fx.Provide(NewEvaluator), NotifierProviders, fx.Provide(NewDispatcher))

type Alert struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Subject   string    `json:"subject,omitempty"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Evaluator struct{}

func NewEvaluator() *Evaluator { return &Evaluator{} }

func (e *Evaluator) Evaluate(state domain.ServerState) []Alert {
	serverID := state.Summary.ID
	createdAt := time.Now().UTC()
	alerts := make([]Alert, 0)
	for _, event := range state.Snapshot.Events {
		if event.Severity == "critical" || event.Type == "process.down" {
			alerts = append(alerts, newAlert(serverID, event.Type, event.Severity, event.Subject, event.Message, createdAt))
		}
	}
	alerts = append(alerts, e.evaluateDNS(serverID, state.Snapshot.Network.PublicIP, state.Snapshot.Network.DNS, createdAt)...)
	alerts = append(alerts, e.evaluatePorts(serverID, state.Snapshot.Network.Ports, createdAt)...)
	return alerts
}

type dnsResult struct {
	Name    string   `json:"name"`
	Domain  string   `json:"domain"`
	Records []string `json:"records"`
	Matches bool     `json:"matches_public_ip"`
	Error   string   `json:"error,omitempty"`
}

func (e *Evaluator) evaluateDNS(serverID string, publicIP string, raw json.RawMessage, createdAt time.Time) []Alert {
	if len(raw) == 0 || publicIP == "" {
		return nil
	}
	var results []dnsResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil
	}
	alerts := make([]Alert, 0)
	for _, result := range results {
		if result.Error != "" {
			alerts = append(alerts, newAlert(serverID, "dns.error", "warning", result.Domain, result.Error, createdAt))
			continue
		}
		if !result.Matches {
			message := fmt.Sprintf("DNS records for %s do not match public IP %s", result.Domain, publicIP)
			alerts = append(alerts, newAlert(serverID, "dns.mismatch", "warning", result.Domain, message, createdAt))
		}
	}
	return alerts
}

type portResult struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Reachable bool   `json:"reachable"`
	Error     string `json:"error,omitempty"`
}

func (e *Evaluator) evaluatePorts(serverID string, raw json.RawMessage, createdAt time.Time) []Alert {
	if len(raw) == 0 {
		return nil
	}
	var results []portResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil
	}
	alerts := make([]Alert, 0)
	for _, result := range results {
		if !result.Reachable {
			message := fmt.Sprintf("Port check %s (%s) is unreachable", result.Name, result.Address)
			if result.Error != "" {
				message += ": " + result.Error
			}
			alerts = append(alerts, newAlert(serverID, "port.unreachable", "warning", result.Name, message, createdAt))
		}
	}
	return alerts
}

func newAlert(serverID string, typ string, severity string, subject string, message string, createdAt time.Time) Alert {
	if severity == "" {
		severity = "warning"
	}
	alert := Alert{ServerID: serverID, Type: typ, Severity: severity, Subject: subject, Message: message, CreatedAt: createdAt}
	alert.ID = fmt.Sprintf("%s:%s:%s:%d", alert.ServerID, alert.Type, alert.Subject, alert.CreatedAt.UnixNano())
	return alert
}

type Notifier interface {
	Notify(ctx context.Context, alert Alert) error
}

type Dispatcher struct {
	notifiers []Notifier
	logger    *zap.Logger
}

type DispatcherParams struct {
	fx.In
	Notifiers []Notifier  `group:"alert_notifiers"`
	Logger    *zap.Logger `optional:"true"`
}

func NewDispatcher(params DispatcherParams) *Dispatcher {
	logger := params.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Dispatcher{notifiers: params.Notifiers, logger: logger.Named("alerts")}
}

func (d *Dispatcher) Dispatch(ctx context.Context, alerts []Alert) error {
	for _, alert := range alerts {
		for _, notifier := range d.notifiers {
			if err := notifier.Notify(ctx, alert); err != nil {
				if ctx.Err() != nil {
					return err
				}
				d.logger.Warn("alert notifier failed", zap.String("alert_id", alert.ID), zap.String("type", alert.Type), zap.Error(err))
			}
		}
	}
	return nil
}

func EnabledTelegram(cfg config.Config) bool {
	return cfg.Alerts.TelegramBotToken != "" && cfg.Alerts.TelegramChatID != ""
}

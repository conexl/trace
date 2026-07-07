package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/incidents"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("notifications",
	fx.Provide(NewTelegramSender),
	fx.Provide(NewService),
	fx.Invoke(RegisterLifecycle),
)

type Sender interface {
	Send(ctx context.Context, text string) error
}

type TelegramSender struct {
	url     string
	chatID  string
	client  *http.Client
	timeout time.Duration
}

func NewTelegramSender(cfg config.Config) (Sender, error) {
	token := strings.TrimSpace(cfg.Notifications.TelegramBotToken)
	chatID := strings.TrimSpace(cfg.Notifications.TelegramChatID)
	if token == "" || chatID == "" {
		return nil, fmt.Errorf("telegram notifications require HOMELYTICS_NOTIFICATIONS_TELEGRAM_BOT_TOKEN and HOMELYTICS_NOTIFICATIONS_TELEGRAM_CHAT_ID")
	}
	return &TelegramSender{
		url:     fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
		chatID:  chatID,
		client:  &http.Client{Timeout: cfg.Notifications.SendTimeout},
		timeout: cfg.Notifications.SendTimeout,
	}, nil
}

func (s *TelegramSender) Send(ctx context.Context, text string) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	payload, err := json.Marshal(map[string]string{
		"chat_id": s.chatID,
		"text":    trimTelegramMessage(text),
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage failed: %s", resp.Status)
	}
	return nil
}

type Service struct {
	client  *redis.Client
	sender  Sender
	channel string
	logger  *zap.Logger
}

func NewService(client *redis.Client, sender Sender, cfg config.Config, logger *zap.Logger) (*Service, error) {
	if client == nil {
		return nil, fmt.Errorf("notification service requires HOMELYTICS_REDIS_ADDR")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		client:  client,
		sender:  sender,
		channel: cfg.Notifications.EventChannel,
		logger:  logger.Named("notifications"),
	}, nil
}

func RegisterLifecycle(lc fx.Lifecycle, svc *Service) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go svc.Run(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}

func (s *Service) Run(ctx context.Context) {
	sub := s.client.Subscribe(ctx, s.channel)
	defer sub.Close()

	s.logger.Info("notification worker subscribed", zap.String("channel", s.channel))
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub.Channel():
			if !ok {
				return
			}
			if err := s.handlePayload(ctx, []byte(msg.Payload)); err != nil {
				s.logger.Warn("notification event failed", zap.Error(err))
			}
		}
	}
}

func (s *Service) handlePayload(ctx context.Context, payload []byte) error {
	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return fmt.Errorf("decode event envelope: %w", err)
	}
	if !shouldNotify(envelope.Type) {
		return nil
	}
	var incident incidents.Incident
	if err := json.Unmarshal(envelope.Data, &incident); err != nil {
		return fmt.Errorf("decode incident event: %w", err)
	}
	return s.sender.Send(ctx, FormatIncidentMessage(envelope.Type, incident))
}

func shouldNotify(eventType string) bool {
	switch eventType {
	case "incident.created", "incident.resolved", "incident.action":
		return true
	default:
		return false
	}
}

func FormatIncidentMessage(eventType string, incident incidents.Incident) string {
	prefix := "Incident update"
	switch eventType {
	case "incident.created":
		prefix = "New incident"
	case "incident.resolved":
		prefix = "Incident resolved"
	case "incident.action":
		prefix = "Incident action"
	}

	lines := []string{
		fmt.Sprintf("%s: %s", prefix, incident.Title),
		fmt.Sprintf("Severity: %s", incident.Severity),
		fmt.Sprintf("Status: %s", incident.Status),
		fmt.Sprintf("Server: %s", incident.ServerID),
		fmt.Sprintf("Service: %s", incident.ServiceName),
	}
	if incident.Summary != "" {
		lines = append(lines, fmt.Sprintf("Summary: %s", incident.Summary))
	}
	if incident.ResolvedAt != nil {
		lines = append(lines, fmt.Sprintf("Resolved at: %s", incident.ResolvedAt.Format(time.RFC3339)))
	}
	lines = append(lines, fmt.Sprintf("Incident ID: %s", incident.ID))
	return trimTelegramMessage(strings.Join(lines, "\n"))
}

func trimTelegramMessage(text string) string {
	const maxTelegramTextLength = 4096
	runes := []rune(text)
	if len(runes) <= maxTelegramTextLength {
		return text
	}
	return string(runes[:maxTelegramTextLength-3]) + "..."
}

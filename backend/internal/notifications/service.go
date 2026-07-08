package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/incidents"
	"backend/internal/telegram"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("notifications",
	fx.Provide(NewTelegramClient),
	fx.Provide(NewService),
	fx.Invoke(RegisterLifecycle),
)

type Sender interface {
	Send(ctx context.Context, chatID string, text string) error
}

type UpdateFetcher interface {
	FetchUpdates(ctx context.Context, offset int, timeout time.Duration) ([]TelegramUpdate, error)
}

type TelegramClient struct {
	baseURL string
	client  *http.Client
	timeout time.Duration
}

func NewTelegramClient(cfg config.Config) (*TelegramClient, error) {
	token := strings.TrimSpace(cfg.Notifications.TelegramBotToken)
	if token == "" {
		return nil, fmt.Errorf("telegram notifications require HOMELYTICS_NOTIFICATIONS_TELEGRAM_BOT_TOKEN")
	}
	return &TelegramClient{
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		client:  &http.Client{Timeout: cfg.Notifications.TelegramPollTimeout + cfg.Notifications.SendTimeout + time.Second},
		timeout: cfg.Notifications.SendTimeout,
	}, nil
}

func (c *TelegramClient) Send(ctx context.Context, chatID string, text string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	payload, err := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    trimTelegramMessage(text),
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sendMessage", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage failed: %s", resp.Status)
	}
	return nil
}

func (c *TelegramClient) FetchUpdates(ctx context.Context, offset int, timeout time.Duration) ([]TelegramUpdate, error) {
	values := url.Values{}
	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}
	values.Set("timeout", strconv.Itoa(int(timeout.Seconds())))
	values.Set("allowed_updates", `["message"]`)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/getUpdates?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", resp.Status)
	}
	var out struct {
		OK     bool             `json:"ok"`
		Result []TelegramUpdate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if !out.OK {
		return nil, fmt.Errorf("telegram getUpdates returned ok=false")
	}
	return out.Result, nil
}

type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

type TelegramMessage struct {
	Text string        `json:"text,omitempty"`
	Chat telegram.Chat `json:"chat"`
}

type Service struct {
	client       *redis.Client
	sender       Sender
	fetcher      UpdateFetcher
	store        telegram.Store
	channel      string
	legacyChatID string
	pollInterval time.Duration
	pollTimeout  time.Duration
	logger       *zap.Logger
}

func NewService(client *redis.Client, telegramClient *TelegramClient, telegramStore telegram.Store, cfg config.Config, logger *zap.Logger) (*Service, error) {
	if client == nil {
		return nil, fmt.Errorf("notification service requires HOMELYTICS_REDIS_ADDR")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		client:       client,
		sender:       telegramClient,
		fetcher:      telegramClient,
		store:        telegramStore,
		channel:      cfg.Notifications.EventChannel,
		legacyChatID: strings.TrimSpace(cfg.Notifications.TelegramChatID),
		pollInterval: cfg.Notifications.TelegramPollInterval,
		pollTimeout:  cfg.Notifications.TelegramPollTimeout,
		logger:       logger.Named("notifications"),
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
	go s.pollTelegram(ctx)

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

func (s *Service) pollTelegram(ctx context.Context) {
	if s.fetcher == nil {
		return
	}
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := s.fetcher.FetchUpdates(ctx, offset, s.pollTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.logger.Warn("telegram update poll failed", zap.Error(err))
			sleepOrDone(ctx, s.pollInterval)
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if err := s.handleTelegramUpdate(ctx, update); err != nil {
				s.logger.Warn("telegram update handling failed", zap.Error(err), zap.Int("update_id", update.UpdateID))
			}
		}
		if len(updates) == 0 {
			sleepOrDone(ctx, s.pollInterval)
		}
	}
}

func (s *Service) handleTelegramUpdate(ctx context.Context, update TelegramUpdate) error {
	if update.Message == nil {
		return nil
	}
	token, ok := startToken(update.Message.Text)
	if !ok {
		return nil
	}
	link, err := s.store.ClaimLink(ctx, token, update.Message.Chat)
	if err != nil {
		_ = s.sender.Send(ctx, fmt.Sprint(update.Message.Chat.ID), "Could not connect Telegram notifications. The link may be expired or already used.")
		return err
	}
	return s.sender.Send(ctx, fmt.Sprint(update.Message.Chat.ID), fmt.Sprintf("Telegram notifications connected for %s.", link.UserEmail))
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
	return s.notifyIncident(ctx, envelope.Type, incident)
}

func (s *Service) notifyIncident(ctx context.Context, eventType string, incident incidents.Incident) error {
	message := FormatIncidentMessage(eventType, incident)
	recipients, err := s.store.ListRecipients(ctx)
	if err != nil {
		return err
	}
	if len(recipients) == 0 && s.legacyChatID != "" {
		return s.sender.Send(ctx, s.legacyChatID, message)
	}
	var sendErr error
	for _, recipient := range recipients {
		if err := s.sender.Send(ctx, fmt.Sprint(recipient.Chat.ID), message); err != nil {
			sendErr = err
			s.logger.Warn("telegram notification send failed", zap.String("user_email", recipient.UserEmail), zap.Error(err))
		}
	}
	return sendErr
}

func shouldNotify(eventType string) bool {
	switch eventType {
	case "incident.created", "incident.resolved", "incident.action":
		return true
	default:
		return false
	}
}

func startToken(text string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) != 2 {
		return "", false
	}
	command := parts[0]
	if command != "/start" && !strings.HasPrefix(command, "/start@") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
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

func sleepOrDone(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

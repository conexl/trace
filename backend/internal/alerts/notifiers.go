package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"backend/internal/config"

	"go.uber.org/fx"
)

type MemoryNotifier struct {
	mu     sync.RWMutex
	limit  int
	alerts []Alert
}

func NewMemoryNotifier(cfg config.Config) *MemoryNotifier {
	limit := cfg.Alerts.MemoryLimit
	if limit <= 0 {
		limit = 200
	}
	return &MemoryNotifier{limit: limit}
}

func (n *MemoryNotifier) Notify(ctx context.Context, alert Alert) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.alerts = append(n.alerts, alert)
	if len(n.alerts) > n.limit {
		n.alerts = n.alerts[len(n.alerts)-n.limit:]
	}
	return nil
}

func (n *MemoryNotifier) Recent(limit int) []Alert {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if limit <= 0 || limit > len(n.alerts) {
		limit = len(n.alerts)
	}
	out := make([]Alert, limit)
	copy(out, n.alerts[len(n.alerts)-limit:])
	return out
}

type TelegramNotifier struct {
	enabled bool
	url     string
	chatID  string
	client  *http.Client
}

func NewTelegramNotifier(cfg config.Config) Notifier {
	if !EnabledTelegram(cfg) {
		return noopNotifier{}
	}
	return &TelegramNotifier{
		enabled: true,
		url:     fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.Alerts.TelegramBotToken),
		chatID:  cfg.Alerts.TelegramChatID,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (n *TelegramNotifier) Notify(ctx context.Context, alert Alert) error {
	if !n.enabled {
		return nil
	}
	payload, err := json.Marshal(map[string]string{
		"chat_id": n.chatID,
		"text":    fmt.Sprintf("[%s] %s: %s", alert.Severity, alert.ServerID, alert.Message),
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram notify failed: %s", resp.Status)
	}
	return nil
}

type noopNotifier struct{}

func (noopNotifier) Notify(context.Context, Alert) error { return nil }

func MemoryAsNotifier(memory *MemoryNotifier) Notifier { return memory }

var NotifierProviders = fx.Options(
	fx.Provide(NewMemoryNotifier),
	fx.Provide(fx.Annotate(MemoryAsNotifier, fx.ResultTags(`group:"alert_notifiers"`))),
	fx.Provide(fx.Annotate(NewTelegramNotifier, fx.As(new(Notifier)), fx.ResultTags(`group:"alert_notifiers"`))),
)

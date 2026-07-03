package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"backend/internal/config"

	"go.uber.org/fx"
)

type Store interface {
	Save(ctx context.Context, alert Alert) error
	Recent(ctx context.Context, limit int) ([]Alert, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	limit  int
	alerts []Alert
}

func NewMemoryStore(cfg config.Config) *MemoryStore {
	limit := cfg.Alerts.MemoryLimit
	if limit <= 0 {
		limit = 200
	}
	return &MemoryStore{limit: limit}
}

func (s *MemoryStore) Save(ctx context.Context, alert Alert) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, alert)
	if len(s.alerts) > s.limit {
		s.alerts = s.alerts[len(s.alerts)-s.limit:]
	}
	return nil
}

func (s *MemoryStore) Recent(ctx context.Context, limit int) ([]Alert, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.alerts) {
		limit = len(s.alerts)
	}
	out := make([]Alert, limit)
	copy(out, s.alerts[len(s.alerts)-limit:])
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

type StoreNotifier struct {
	store Store
}

func NewStoreNotifier(store Store) Notifier { return StoreNotifier{store: store} }

func (n StoreNotifier) Notify(ctx context.Context, alert Alert) error {
	return n.store.Save(ctx, alert)
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

var NotifierProviders = fx.Options(
	fx.Provide(NewStore),
	fx.Provide(fx.Annotate(NewStoreNotifier, fx.ResultTags(`group:"alert_notifiers"`))),
	fx.Provide(fx.Annotate(NewTelegramNotifier, fx.As(new(Notifier)), fx.ResultTags(`group:"alert_notifiers"`))),
)

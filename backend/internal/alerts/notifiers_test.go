package alerts

import (
	"context"
	"errors"
	"testing"

	"backend/internal/config"
)

func TestMemoryNotifierKeepsRecentAlerts(t *testing.T) {
	notifier := NewMemoryNotifier(config.Config{Alerts: config.AlertsConfig{MemoryLimit: 2}})
	_ = notifier.Notify(context.Background(), Alert{ID: "one"})
	_ = notifier.Notify(context.Background(), Alert{ID: "two"})
	_ = notifier.Notify(context.Background(), Alert{ID: "three"})
	recent := notifier.Recent(10)
	if len(recent) != 2 || recent[0].ID != "two" || recent[1].ID != "three" {
		t.Fatalf("recent = %#v", recent)
	}
}

func TestDispatcherTreatsNotifierFailuresAsBestEffort(t *testing.T) {
	memory := NewMemoryNotifier(config.Config{Alerts: config.AlertsConfig{MemoryLimit: 10}})
	dispatcher := NewDispatcher(DispatcherParams{Notifiers: []Notifier{failingNotifier{}, memory}})
	if err := dispatcher.Dispatch(context.Background(), []Alert{{ID: "one", Type: "process.down"}}); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if recent := memory.Recent(10); len(recent) != 1 || recent[0].ID != "one" {
		t.Fatalf("recent = %#v", recent)
	}
}

type failingNotifier struct{}

func (failingNotifier) Notify(context.Context, Alert) error {
	return errors.New("boom")
}

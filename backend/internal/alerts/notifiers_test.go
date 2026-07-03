package alerts

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/config"
)

func TestMemoryStoreKeepsRecentAlerts(t *testing.T) {
	store := NewMemoryStore(config.Config{Alerts: config.AlertsConfig{MemoryLimit: 2}})
	_ = store.Save(context.Background(), Alert{ID: "one", CreatedAt: time.Now().Add(-3 * time.Second)})
	_ = store.Save(context.Background(), Alert{ID: "two", CreatedAt: time.Now().Add(-2 * time.Second)})
	_ = store.Save(context.Background(), Alert{ID: "three", CreatedAt: time.Now().Add(-time.Second)})
	recent, err := store.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Recent() error = %v", err)
	}
	if len(recent) != 2 || recent[0].ID != "three" || recent[1].ID != "two" {
		t.Fatalf("recent = %#v", recent)
	}
}

func TestStoreNotifierPersistsAlerts(t *testing.T) {
	store := NewMemoryStore(config.Config{Alerts: config.AlertsConfig{MemoryLimit: 10}})
	notifier := NewStoreNotifier(store)
	if err := notifier.Notify(context.Background(), Alert{ID: "one", Type: "process.down", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	recent, err := store.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Recent() error = %v", err)
	}
	if len(recent) != 1 || recent[0].ID != "one" {
		t.Fatalf("recent = %#v", recent)
	}
}

func TestDispatcherTreatsNotifierFailuresAsBestEffort(t *testing.T) {
	store := NewMemoryStore(config.Config{Alerts: config.AlertsConfig{MemoryLimit: 10}})
	dispatcher := NewDispatcher(DispatcherParams{Notifiers: []Notifier{failingNotifier{}, NewStoreNotifier(store)}})
	if err := dispatcher.Dispatch(context.Background(), []Alert{{ID: "one", Type: "process.down", CreatedAt: time.Now()}}); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	recent, err := store.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Recent() error = %v", err)
	}
	if len(recent) != 1 || recent[0].ID != "one" {
		t.Fatalf("recent = %#v", recent)
	}
}

type failingNotifier struct{}

func (failingNotifier) Notify(context.Context, Alert) error {
	return errors.New("boom")
}

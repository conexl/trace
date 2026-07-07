package incidents

import (
	"context"
	"math"
	"testing"
	"time"

	"backend/internal/pubsub"

	"go.uber.org/zap"
)

func TestServiceMetricsCalculatesMTTRAndFrequency(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(ServiceParams{Store: store, Pubsub: pubsub.New(nil), Logger: zap.NewNop()})
	now := time.Now().UTC()
	resolvedAt := now.Add(-2 * time.Hour)

	if err := store.Save(context.Background(), Incident{
		ID:          "one",
		ServerID:    "server-a",
		ServiceName: "nginx",
		Status:      "resolved",
		Severity:    "critical",
		CreatedAt:   resolvedAt.Add(-30 * time.Minute),
		UpdatedAt:   resolvedAt,
		ResolvedAt:  &resolvedAt,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), Incident{
		ID:          "two",
		ServerID:    "server-a",
		ServiceName: "nginx",
		Status:      "open",
		Severity:    "warning",
		CreatedAt:   now.Add(-1 * time.Hour),
		UpdatedAt:   now.Add(-1 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), Incident{
		ID:          "outside-window",
		ServerID:    "server-a",
		ServiceName: "postgres",
		Status:      "open",
		Severity:    "critical",
		CreatedAt:   now.Add(-10 * 24 * time.Hour),
		UpdatedAt:   now.Add(-10 * 24 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	metrics, err := service.Metrics(context.Background(), "server-a", 24*time.Hour)
	if err != nil {
		t.Fatalf("Metrics() error = %v", err)
	}

	if metrics.Total != 2 || metrics.Open != 1 || metrics.Resolved != 1 || metrics.Critical != 1 || metrics.Warning != 1 {
		t.Fatalf("metrics = %#v", metrics)
	}
	if math.Abs(metrics.MTTRSeconds-1800) > 0.001 {
		t.Fatalf("MTTRSeconds = %f", metrics.MTTRSeconds)
	}
	if math.Abs(metrics.FrequencyPerDay-2) > 0.001 {
		t.Fatalf("FrequencyPerDay = %f", metrics.FrequencyPerDay)
	}
	nginx := metrics.ByService["nginx"]
	if nginx.Total != 2 || nginx.Open != 1 || nginx.Resolved != 1 {
		t.Fatalf("nginx metrics = %#v", nginx)
	}
}

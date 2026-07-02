package collectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent/internal/config"
)

func TestSpeedTestsMeasureDownloadedBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer server.Close()

	collector := NewNetworkCollector()
	results := collector.speedTests(context.Background(), []config.SpeedTest{{Name: "local", URL: server.URL, MaxBytes: 4, Timeout: time.Second}})
	if len(results) != 1 {
		t.Fatalf("results = %#v", results)
	}
	if results[0].BytesRead != 4 {
		t.Fatalf("BytesRead = %d", results[0].BytesRead)
	}
	if results[0].Mbps <= 0 {
		t.Fatalf("Mbps = %f", results[0].Mbps)
	}
}

package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent/internal/collectors"
	"agent/internal/config"
)

func TestHTTPClientSendsSnapshots(t *testing.T) {
	var gotAuth string
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client, err := NewHTTPClient(config.CloudConfig{Endpoint: server.URL, Token: "pairing-token"})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SendSnapshots(context.Background(), []collectors.Snapshot{{AgentName: "dev", Collected: time.Now()}}); err != nil {
		t.Fatalf("SendSnapshots() error = %v", err)
	}
	if gotPath != "/v1/agent/snapshots" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer pairing-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
}

func TestHTTPClientReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewHTTPClient(config.CloudConfig{Endpoint: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SendSnapshots(context.Background(), []collectors.Snapshot{{AgentName: "dev"}}); err == nil {
		t.Fatal("SendSnapshots() expected status error")
	}
}

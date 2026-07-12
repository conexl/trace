package pairing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agent/internal/config"
)

func TestClientClaimSendsPairingToken(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(Response{AgentID: "agent-1", Certificate: "cert", PrivateKey: "key", CACert: "ca", ExpiresAt: time.Now()})
	}))
	defer server.Close()

	client, err := NewClient(config.CloudConfig{Endpoint: server.URL, Token: "pair-token"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Claim(context.Background(), Request{AgentName: "devbox"})
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if gotAuth != "Bearer pair-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if resp.AgentID != "agent-1" {
		t.Fatalf("AgentID = %q", resp.AgentID)
	}
}

func TestNewClientAllowsFutureCAFileDuringInitialPairing(t *testing.T) {
	missingCA := filepath.Join(t.TempDir(), "certs", "ca.pem")
	client, err := NewClient(config.CloudConfig{
		Endpoint: "https://trace.solen.one",
		MTLS:     config.MTLS{CAFile: missingCA},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
}

func TestSaveCredentialsWritesSecretFiles(t *testing.T) {
	dir := t.TempDir()
	saved, err := SaveCredentials(Response{Certificate: "cert", PrivateKey: "key", CACert: "ca"}, SaveOptions{Dir: dir})
	if err != nil {
		t.Fatalf("SaveCredentials() error = %v", err)
	}
	for _, path := range []string{saved.CAFile, saved.CertFile, saved.KeyFile} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode = %s", path, info.Mode().Perm())
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "agent-key.pem")); string(got) != "key" {
		t.Fatalf("key = %q", got)
	}
}

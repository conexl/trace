package httpapi

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/ingest"
	"backend/internal/security"
	"backend/internal/store"

	"go.uber.org/zap/zaptest"
)

func newTestServer(t *testing.T, cfg config.Config) *Server {
	t.Helper()
	if cfg.State.OfflineAfter == 0 {
		cfg.State.OfflineAfter = time.Minute
	}
	if cfg.State.MaxEvents == 0 {
		cfg.State.MaxEvents = 10
	}
	memory := store.NewMemoryStore(cfg)
	ca, err := security.NewCertificateAuthority(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pairing := security.NewPairingService(cfg, ca)
	return NewServer(cfg, memory, ingest.NewService(memory), pairing, zaptest.NewLogger(t))
}

func TestIngestAndReadServerState(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{IngestTokens: map[string]struct{}{"agent-token": {}}, AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","host":{"hostname":"arch","platform":"linux"},"system":{"cpu_percent":7,"memory":{"used_percent":30}},"network":{"public_ip":"203.0.113.1"},"collected_at":"2026-07-02T09:00:00Z"}]}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer agent-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/servers", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", w.Code, w.Body.String())
	}
	var list struct {
		Servers []map[string]any `json:"servers"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Servers) != 1 || list.Servers[0]["id"] != "devbox" {
		t.Fatalf("list = %#v", list)
	}
}

func TestIngestRequiresConfiguredToken(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{IngestTokens: map[string]struct{}{"agent-token": {}}}}
	server := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader([]byte(`{"snapshots":[]}`)))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAdminTokenProtectsReadAPI(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodGet, "/v1/servers", nil)
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestPairingClaimEndpoint(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Pairing: config.PairingConfig{Tokens: map[string]struct{}{"pair-once": {}}, CertTTL: time.Hour}}
	server := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/pairing/claim", bytes.NewReader([]byte(`{"agent_name":"devbox","hostname":"arch"}`)))
	req.Header.Set("Authorization", "Bearer pair-once")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["agent_id"] == "" || resp["certificate_pem"] == "" || resp["private_key_pem"] == "" {
		t.Fatalf("response = %#v", resp)
	}
}

func TestIngestAllowsVerifiedClientCertificate(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{IngestTokens: map[string]struct{}{"agent-token": {}}}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"mtls-agent","collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	req.TLS = &tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{{}}}, PeerCertificates: []*x509.Certificate{{}}}
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}

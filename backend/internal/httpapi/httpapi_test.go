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

	"backend/internal/ai"
	"backend/internal/alerts"
	"backend/internal/audit"
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/incidents"
	"backend/internal/ingest"
	"backend/internal/presence"
	"backend/internal/pubsub"
	"backend/internal/security"
	"backend/internal/serverconfig"
	"backend/internal/store"
	"backend/internal/tasks"
	"backend/internal/users"

	"go.uber.org/zap"
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
	if cfg.Auth.SessionTTL == 0 {
		cfg.Auth.SessionTTL = 24 * time.Hour
	}
	memory := store.NewMemoryStore(cfg)
	ca, err := security.NewCertificateAuthority(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pairing := security.NewPairingService(cfg, ca)
	alertStore := alerts.NewMemoryStore(cfg)
	dispatcher := alerts.NewDispatcher(alerts.DispatcherParams{Notifiers: []alerts.Notifier{alerts.NewStoreNotifier(alertStore)}})
	presenceService := presence.NewService(cfg, presence.NewMemoryStore())
	incidentStore := incidents.NewMemoryStore()
	pubsubService := pubsub.New(nil)
	incidentService := incidents.NewService(incidents.ServiceParams{
		Store:  incidentStore,
		Pubsub: pubsubService,
		Logger: zaptest.NewLogger(t),
	})
	ingestService := ingest.NewService(memory, alerts.NewEvaluator(), dispatcher, incidentService, presenceService, zaptest.NewLogger(t))
	userStore := users.NewMemoryStore()
	authService := auth.NewService(cfg, userStore, auth.NewMemorySessionStore())
	configStore := serverconfig.NewMemoryStore()
	auditStore := audit.NewMemoryStore()
	aiClient := ai.NewClient(ai.ClientParams{Config: cfg, Logger: zap.NewNop()})
	aiAnalyzer := ai.NewAnalyzer(aiClient, zaptest.NewLogger(t))
	return NewServer(cfg, memory, ingestService, pairing, tasks.NewMemoryStore(), alertStore, incidentService, aiAnalyzer, presenceService, authService, auditStore, pubsubService, configStore, nil, zaptest.NewLogger(t))
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

func TestCORSPreflightAllowsConfiguredOrigin(t *testing.T) {
	cfg := config.Config{HTTP: config.HTTPConfig{AllowedOrigins: []string{"http://localhost:5173"}}, State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}}
	server := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodOptions, "/v1/servers", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow-origin = %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("missing allow-headers")
	}
}

func TestCORSPreflightRejectsUnknownOrigin(t *testing.T) {
	cfg := config.Config{HTTP: config.HTTPConfig{AllowedOrigins: []string{"http://localhost:5173"}}, State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}}
	server := newTestServer(t, cfg)
	req := httptest.NewRequest(http.MethodOptions, "/v1/servers", nil)
	req.Header.Set("Origin", "http://evil.example")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
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

func TestIngestRequiresClientCertificateWhenConfigured(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, TLS: config.TLSConfig{RequireClientCert: true}, Auth: config.AuthConfig{IngestTokens: map[string]struct{}{"agent-token": {}}}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"mtls-agent","collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer agent-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskQueueHTTPLifecycle(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/tasks", bytes.NewReader([]byte(`{"task_name":"disk-usage"}`)))
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("enqueue status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agent/tasks?agent_id=devbox", nil)
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("poll status = %d body=%s", w.Code, w.Body.String())
	}
	var poll struct {
		Tasks []tasks.Task `json:"tasks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &poll); err != nil {
		t.Fatal(err)
	}
	if len(poll.Tasks) != 1 || poll.Tasks[0].Status != tasks.StatusRunning {
		t.Fatalf("poll = %#v", poll)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/agent/tasks/"+poll.Tasks[0].ID+"/result", bytes.NewReader([]byte(`{"exit_code":0,"stdout":"ok","duration_ms":5,"started_at":"2026-07-02T09:00:00Z"}`)))
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("result status = %d body=%s", w.Code, w.Body.String())
	}
	var completed tasks.Task
	if err := json.Unmarshal(w.Body.Bytes(), &completed); err != nil {
		t.Fatal(err)
	}
	if completed.Status != tasks.StatusCompleted || completed.Result.Stdout != "ok" {
		t.Fatalf("completed = %#v", completed)
	}
}

func TestServiceActionQueuesRemoteControllableService(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","processes":[{"name":"nginx","service":"nginx","remote_control":true,"running":true}],"collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/service-actions", bytes.NewReader([]byte(`{"service":"nginx","action":"restart"}`)))
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("action status = %d body=%s", w.Code, w.Body.String())
	}
	var task tasks.Task
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Fatal(err)
	}
	if task.Name != "service-action" || task.Payload.Service != "nginx" || task.Payload.Action != "restart" {
		t.Fatalf("task = %#v", task)
	}
}

func TestServiceActionRejectsNonRemoteControllableService(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","processes":[{"name":"nginx","service":"nginx","running":true}],"collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/service-actions", bytes.NewReader([]byte(`{"service":"nginx","action":"restart"}`)))
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("action status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskPollingUpdatesPresence(t *testing.T) {
	now := time.Now().UTC()
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","collected_at":"` + now.Add(-2*time.Minute).Format(time.RFC3339) + `"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agent/tasks?agent_id=devbox", nil)
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("poll status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/servers/devbox", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", w.Code, w.Body.String())
	}
	var state struct {
		Summary struct {
			Status string `json:"status"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Summary.Status != "online" {
		t.Fatalf("status = %q body=%s", state.Summary.Status, w.Body.String())
	}
}

func TestAlertsCreatedFromIngest(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Alerts: config.AlertsConfig{MemoryLimit: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","events":[{"type":"process.down","severity":"critical","subject":"nginx","message":"critical process is not running","timestamp":"2026-07-02T09:00:00Z"}],"collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/v1/alerts", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("alerts status = %d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		Alerts []alerts.Alert `json:"alerts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Alerts) != 1 || out.Alerts[0].Type != "process.down" || out.Alerts[0].Subject != "nginx" {
		t.Fatalf("alerts = %#v", out.Alerts)
	}
}

func TestIncidentMetricsCreatedFromIngest(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","events":[{"type":"process.down","severity":"critical","subject":"nginx","message":"critical process is not running","timestamp":"2026-07-02T09:00:00Z"}],"collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/incidents/metrics?window=7d", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("metrics status = %d body=%s", w.Code, w.Body.String())
	}
	var out incidents.Metrics
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Total != 1 || out.Open != 1 || out.Critical != 1 {
		t.Fatalf("metrics = %#v", out)
	}
}

func TestIncidentDiagnosticsActionQueuesTask(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	incident := createOpenIncidentForTest(t, server)

	req := httptest.NewRequest(http.MethodPost, "/v1/incidents/"+incident.ID+"/diagnostics", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("diagnostics status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agent/tasks?agent_id=devbox", nil)
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("poll status = %d body=%s", w.Code, w.Body.String())
	}
	var poll struct {
		Tasks []tasks.Task `json:"tasks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &poll); err != nil {
		t.Fatal(err)
	}
	if len(poll.Tasks) != 1 || poll.Tasks[0].Name != "diagnostics" || poll.Tasks[0].Payload.IncidentID != incident.ID || poll.Tasks[0].Payload.Service != "nginx" {
		t.Fatalf("tasks = %#v", poll.Tasks)
	}
}

func TestIncidentRollbackConfigRestoresPreviousDesiredConfig(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)

	setServerConfigForTest(t, server, domain.AgentDesiredConfig{
		Processes: []domain.ProcessConfig{{Name: "nginx", Service: "nginx", Restart: true}},
	})
	setServerConfigForTest(t, server, domain.AgentDesiredConfig{
		Processes: []domain.ProcessConfig{{Name: "nginx", Service: "nginx", Restart: false}},
	})
	incident := createOpenIncidentForTest(t, server)

	req := httptest.NewRequest(http.MethodPost, "/v1/incidents/"+incident.ID+"/rollback-config", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("rollback status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/servers/devbox/config", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get config status = %d body=%s", w.Code, w.Body.String())
	}
	var out domain.AgentDesiredConfig
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Revision != 3 || len(out.Processes) != 1 || !out.Processes[0].Restart {
		t.Fatalf("config = %#v", out)
	}
}

func createOpenIncidentForTest(t *testing.T, server *Server) incidents.Incident {
	t.Helper()
	payload := []byte(`{"snapshots":[{"agent_name":"devbox","events":[{"type":"process.down","severity":"critical","subject":"nginx","message":"critical process is not running","timestamp":"2026-07-02T09:00:00Z"}],"collected_at":"2026-07-02T09:00:00Z"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/agent/snapshots", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("ingest status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/incidents?limit=1", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list incidents status = %d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		Incidents []incidents.Incident `json:"incidents"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Incidents) != 1 {
		t.Fatalf("incidents = %#v", out.Incidents)
	}
	return out.Incidents[0]
}

func setServerConfigForTest(t *testing.T, server *Server, cfg domain.AgentDesiredConfig) domain.AgentDesiredConfig {
	t.Helper()
	body, _ := json.Marshal(cfg)
	req := httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/config", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("set config status = %d body=%s", w.Code, w.Body.String())
	}
	var out domain.AgentDesiredConfig
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func registerUser(t *testing.T, server *Server, email, password string, adminToken string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	if adminToken != "" {
		req.Header.Set("Authorization", "Bearer "+adminToken)
	}
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register status = %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp.Token
}

func TestFirstUserBecomesOwner(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}}
	server := newTestServer(t, cfg)
	token := registerUser(t, server, "owner@example.com", "password123", "")

	req := httptest.NewRequest(http.MethodGet, "/v1/servers", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("owner should access admin endpoint: status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestRegistrationDisabledRejectsSecondUser(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{RegistrationDisabled: true}}
	server := newTestServer(t, cfg)
	registerUser(t, server, "owner@example.com", "password123", "")

	body, _ := json.Marshal(map[string]string{"email": "second@example.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("second registration status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestAdminTokenCanCreateAdminUser(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token", RegistrationDisabled: true}}
	server := newTestServer(t, cfg)
	registerUser(t, server, "owner@example.com", "password123", "")
	adminToken := registerUser(t, server, "admin@example.com", "password123", "admin-token")

	req := httptest.NewRequest(http.MethodGet, "/v1/servers", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("admin token created user should access admin endpoint: status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestViewerCannotAccessAdminEndpoints(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}}
	server := newTestServer(t, cfg)
	registerUser(t, server, "owner@example.com", "password123", "")
	viewerToken := registerUser(t, server, "viewer@example.com", "password123", "")

	req := httptest.NewRequest(http.MethodGet, "/v1/servers", nil)
	req.Header.Set("Authorization", "Bearer "+viewerToken)
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("viewer list servers status = %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+viewerToken)
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("viewer admin endpoint status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestDNSRecheckTaskEnqueueWithDomains(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/tasks", bytes.NewReader([]byte(`{"task_name":"dns-recheck","domains":["example.com","example.org"]}`)))
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("enqueue status = %d body=%s", w.Code, w.Body.String())
	}
	var task tasks.Task
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Fatal(err)
	}
	if task.Name != "dns-recheck" || len(task.Payload.Domains) != 2 {
		t.Fatalf("task = %#v", task)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agent/tasks?agent_id=devbox", nil)
	w = httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("poll status = %d body=%s", w.Code, w.Body.String())
	}
	var poll struct {
		Tasks []tasks.Task `json:"tasks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &poll); err != nil {
		t.Fatal(err)
	}
	if len(poll.Tasks) != 1 || len(poll.Tasks[0].Payload.Domains) != 2 {
		t.Fatalf("poll = %#v", poll)
	}
}

func TestDNSRecheckTaskRejectsMissingDomains(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/tasks", bytes.NewReader([]byte(`{"task_name":"dns-recheck"}`)))
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestLoginRateLimitBlocksExcess(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token", LoginRateLimit: 2, LoginRateWindow: time.Minute}}
	server := newTestServer(t, cfg)

	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "wrong"})
	makeReq := func() int {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
		w := httptest.NewRecorder()
		server.securityHeaders(server.mux).ServeHTTP(w, req)
		return w.Code
	}
	if makeReq() != http.StatusUnauthorized {
		t.Fatalf("first request should be unauthorized")
	}
	if makeReq() != http.StatusUnauthorized {
		t.Fatalf("second request should be unauthorized")
	}
	if makeReq() != http.StatusTooManyRequests {
		t.Fatalf("third request should be rate limited, got %d", makeReq())
	}
}

func TestLoginRateLimitIgnoresForwardedHeadersByDefault(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token", LoginRateLimit: 1, LoginRateWindow: time.Minute}}
	server := newTestServer(t, cfg)

	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "wrong"})
	makeReq := func(xff string) int {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("X-Forwarded-For", xff)
		w := httptest.NewRecorder()
		server.securityHeaders(server.mux).ServeHTTP(w, req)
		return w.Code
	}
	if makeReq("1.2.3.4") != http.StatusUnauthorized {
		t.Fatalf("first request should be unauthorized")
	}
	if makeReq("5.6.7.8") != http.StatusTooManyRequests {
		t.Fatalf("second request with spoofed X-Forwarded-For should still be rate limited, got %d", makeReq("5.6.7.8"))
	}
}

func TestLoginRateLimitUsesForwardedHeadersWhenTrusted(t *testing.T) {
	cfg := config.Config{
		State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10},
		Auth:  config.AuthConfig{AdminToken: "admin-token", LoginRateLimit: 1, LoginRateWindow: time.Minute},
		HTTP:  config.HTTPConfig{TrustForwardedHeaders: true},
	}
	server := newTestServer(t, cfg)

	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "wrong"})
	makeReq := func(xff string) int {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("X-Forwarded-For", xff)
		w := httptest.NewRecorder()
		server.securityHeaders(server.mux).ServeHTTP(w, req)
		return w.Code
	}
	if makeReq("1.2.3.4") != http.StatusUnauthorized {
		t.Fatalf("first request should be unauthorized")
	}
	if makeReq("5.6.7.8") != http.StatusUnauthorized {
		t.Fatalf("second request from different forwarded IP should not be rate limited, got %d", makeReq("5.6.7.8"))
	}
	if makeReq("1.2.3.4") != http.StatusTooManyRequests {
		t.Fatalf("third request from same forwarded IP should be rate limited, got %d", makeReq("1.2.3.4"))
	}
}

func TestRegisterRateLimitBlocksExcess(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{LoginRateLimit: 100, LoginRateWindow: time.Minute, RegisterRateLimit: 1, RegisterRateWindow: time.Minute}}
	server := newTestServer(t, cfg)

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123"})
	makeReq := func() int {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
		w := httptest.NewRecorder()
		server.securityHeaders(server.mux).ServeHTTP(w, req)
		return w.Code
	}
	if makeReq() != http.StatusCreated {
		t.Fatalf("first registration should succeed")
	}
	if makeReq() != http.StatusTooManyRequests {
		t.Fatalf("second registration should be rate limited, got %d", makeReq())
	}
}

func TestServerConfigRejectsShellEnabled(t *testing.T) {
	cfg := config.Config{State: config.StateConfig{OfflineAfter: time.Minute, MaxEvents: 10}, Auth: config.AuthConfig{AdminToken: "admin-token"}}
	server := newTestServer(t, cfg)
	body, _ := json.Marshal(map[string]any{"remote": map[string]any{"shell_enabled": true}})
	req := httptest.NewRequest(http.MethodPost, "/v1/servers/devbox/config", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	server.securityHeaders(server.mux).ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("shell enabled config status = %d body=%s", w.Code, w.Body.String())
	}
}

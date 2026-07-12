package httpapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	"backend/internal/ratelimit"
	"backend/internal/security"
	"backend/internal/serverconfig"
	"backend/internal/store"
	"backend/internal/tasks"
	"backend/internal/telegram"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("httpapi", fx.Provide(NewServer), fx.Invoke(RegisterLifecycle))

type Server struct {
	cfg             config.Config
	store           store.Store
	ingest          *ingest.Service
	pairing         *security.PairingService
	tasks           tasks.Store
	alerts          alerts.Store
	incidents       *incidents.Service
	aiAnalyzer      *ai.Analyzer
	presence        *presence.Service
	auth            *auth.Service
	audit           audit.Store
	pubsub          *pubsub.Service
	configStore     serverconfig.Store
	telegramStore   telegram.Store
	logger          *zap.Logger
	mux             *http.ServeMux
	loginLimiter    ratelimit.Limiter
	registerLimiter ratelimit.Limiter
}

const sessionCookieName = "homelytics_session"

func NewServer(cfg config.Config, store store.Store, ingest *ingest.Service, pairing *security.PairingService, taskStore tasks.Store, alertStore alerts.Store, incidentService *incidents.Service, aiAnalyzer *ai.Analyzer, presenceService *presence.Service, authService *auth.Service, auditStore audit.Store, pubsubService *pubsub.Service, configStore serverconfig.Store, telegramStore telegram.Store, redisClient *redis.Client, logger *zap.Logger) *Server {
	var loginLimiter, registerLimiter ratelimit.Limiter
	if redisClient != nil {
		loginLimiter = ratelimit.NewRedis(redisClient, cfg.Auth.LoginRateLimit, cfg.Auth.LoginRateWindow, "login")
		registerLimiter = ratelimit.NewRedis(redisClient, cfg.Auth.RegisterRateLimit, cfg.Auth.RegisterRateWindow, "register")
	} else {
		memLogin := ratelimit.NewMemory(cfg.Auth.LoginRateLimit, cfg.Auth.LoginRateWindow)
		memRegister := ratelimit.NewMemory(cfg.Auth.RegisterRateLimit, cfg.Auth.RegisterRateWindow)
		go memLogin.Cleanup(5 * time.Minute)
		go memRegister.Cleanup(5 * time.Minute)
		loginLimiter = memLogin
		registerLimiter = memRegister
	}

	server := &Server{
		cfg: cfg, store: store, ingest: ingest, pairing: pairing, tasks: taskStore, alerts: alertStore, incidents: incidentService, aiAnalyzer: aiAnalyzer, presence: presenceService, auth: authService, audit: auditStore, pubsub: pubsubService, configStore: configStore, telegramStore: telegramStore,
		logger:          logger.Named("http"),
		mux:             http.NewServeMux(),
		loginLimiter:    loginLimiter,
		registerLimiter: registerLimiter,
	}
	server.routes()
	return server
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              s.cfg.HTTP.Addr,
		Handler:           s.securityHeaders(s.mux),
		ReadTimeout:       s.cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      s.cfg.HTTP.WriteTimeout,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		TLSConfig:         s.tlsConfig(),
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("POST /v1/auth/login", ratelimit.Middleware(s.loginLimiter, s.cfg.HTTP.TrustForwardedHeaders, s.handleLogin))
	s.mux.HandleFunc("POST /v1/auth/register", ratelimit.Middleware(s.registerLimiter, s.cfg.HTTP.TrustForwardedHeaders, s.handleRegister))
	s.mux.HandleFunc("POST /v1/auth/logout", s.handleLogout)
	s.mux.HandleFunc("GET /v1/auth/me", s.requireAuth(s.handleAuthMe))
	s.mux.HandleFunc("GET /v1/billing/plan", s.requireAuth(s.handleGetBillingPlan))
	s.mux.HandleFunc("POST /v1/billing/plan", s.requireAuth(s.handleUpdateBillingPlan))
	s.mux.HandleFunc("GET /v1/notifications/telegram", s.requireAuth(s.handleGetTelegramNotificationStatus))
	s.mux.HandleFunc("POST /v1/notifications/telegram/link", s.requirePlus(s.handleCreateTelegramNotificationLink))
	s.mux.HandleFunc("DELETE /v1/notifications/telegram", s.requireAuth(s.handleDeleteTelegramNotificationLink))
	s.mux.HandleFunc("POST /v1/pairing/codes", s.requireAuth(s.handleCreatePairingCode))
	s.mux.HandleFunc("POST /v1/pairing/claim", s.handlePairingClaim)
	s.mux.HandleFunc("POST /v1/agent/snapshots", s.requireAgent(s.handleIngest))
	s.mux.HandleFunc("GET /v1/agent/tasks", s.requireAgent(s.handlePollTasks))
	s.mux.HandleFunc("GET /v1/agent/config", s.requireAgent(s.handleGetAgentConfig))
	s.mux.HandleFunc("POST /v1/agent/tasks/", s.requireAgent(s.handleCompleteTask))
	s.mux.HandleFunc("GET /v1/alerts", s.requireAuth(s.handleListAlerts))
	s.mux.HandleFunc("GET /v1/incidents", s.requireAuth(s.handleListIncidents))
	s.mux.HandleFunc("GET /v1/incidents/metrics", s.requireAuth(s.handleIncidentMetrics))
	s.mux.HandleFunc("GET /v1/incidents/actions", s.requireAuth(s.handleGetIncidentActions))
	s.mux.HandleFunc("GET /v1/incidents/", s.requireAuth(s.handleGetIncident))
	s.mux.HandleFunc("POST /v1/incidents/", s.requirePlus(s.handleIncidentAction))
	s.mux.HandleFunc("POST /v1/incidents/{id}/analyze", s.requirePlus(s.handleAnalyzeIncident))
	s.mux.HandleFunc("GET /v1/servers", s.requireAuth(s.handleListServers))
	s.mux.HandleFunc("POST /v1/servers/", s.requirePlus(s.handleServerAction))
	s.mux.HandleFunc("GET /v1/tasks/", s.requireAuth(s.handleGetTask))
	s.mux.HandleFunc("GET /v1/tasks", s.requireAuth(s.handleListTasks))
	s.mux.HandleFunc("GET /v1/servers/", s.requireAuth(s.handleGetServer))
	s.mux.HandleFunc("GET /v1/servers/{id}/config", s.requireAuth(s.handleGetServerConfig))
	s.mux.HandleFunc("GET /v1/servers/{id}/metrics", s.requireAuth(s.handleGetMetrics))
	s.mux.HandleFunc("POST /v1/servers/{id}/config", s.requirePlus(s.handleSetServerConfig))
	s.mux.HandleFunc("GET /v1/audit", s.requirePlus(s.handleListAudit))
	s.mux.HandleFunc("GET /v1/events", s.requireAuth(s.handleEvents))
}

func RegisterLifecycle(lc fx.Lifecycle, api *Server, logger *zap.Logger, cfg config.Config) {
	httpServer := api.HTTPServer()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("http server listening", zap.String("addr", httpServer.Addr))
				if err := listenAndServe(httpServer, cfg); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("http server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeout)
			defer cancel()
			return httpServer.Shutdown(shutdownCtx)
		},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "time": time.Now().UTC()})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid login request")
		return
	}
	token, err := s.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"token": token})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := s.getToken(r)
	if token != "" {
		_ = s.auth.Logout(r.Context(), token)
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	session, ok := s.currentSession(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	subscription := s.subscriptionForSession(r.Context(), session)
	writeJSON(w, http.StatusOK, map[string]any{
		"email":        session.Email,
		"role":         session.Role,
		"subscription": subscription,
	})
}

func (s *Server) handleGetBillingPlan(w http.ResponseWriter, r *http.Request) {
	session, ok := s.currentSession(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	writeJSON(w, http.StatusOK, s.subscriptionForSession(r.Context(), session))
}

func (s *Server) handleUpdateBillingPlan(w http.ResponseWriter, r *http.Request) {
	session, ok := s.currentSession(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	if s.auth.IsAdminToken(s.getToken(r)) {
		writeError(w, http.StatusBadRequest, "dev admin token does not have a billable account")
		return
	}
	defer r.Body.Close()
	var req struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid billing request")
		return
	}
	req.Plan = strings.TrimSpace(strings.ToLower(req.Plan))
	if req.Plan != domain.PlanFree && req.Plan != domain.PlanPlus {
		writeError(w, http.StatusBadRequest, "plan must be free or plus")
		return
	}
	user, err := s.auth.UpdatePlan(r.Context(), session.Email, req.Plan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update billing plan failed")
		return
	}
	subscription := domain.EntitlementsForPlan(user.Plan)
	_ = s.audit.Log(r.Context(), domain.AuditLog{
		UserEmail: session.Email,
		Action:    "billing-plan-change",
		Target:    session.Email,
		Details:   fmt.Sprintf("plan: %s", subscription.Plan),
	})
	writeJSON(w, http.StatusOK, subscription)
}

func (s *Server) handleGetTelegramNotificationStatus(w http.ResponseWriter, r *http.Request) {
	email := s.userEmail(r)
	recipient, err := s.telegramStore.GetRecipient(r.Context(), email)
	if err != nil {
		if errors.Is(err, telegram.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{
				"connected": false,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "get telegram status failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connected": true,
		"chat":      recipient.Chat,
		"linked_at": recipient.LinkedAt,
	})
}

func (s *Server) handleCreateTelegramNotificationLink(w http.ResponseWriter, r *http.Request) {
	botUsername := strings.TrimSpace(s.cfg.Notifications.TelegramBotUsername)
	if botUsername == "" {
		writeError(w, http.StatusServiceUnavailable, "telegram bot username is not configured")
		return
	}

	link, err := s.telegramStore.CreateLink(r.Context(), s.userEmail(r), s.cfg.Notifications.LinkTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create telegram link failed")
		return
	}
	startURL := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, link.Token)
	writeJSON(w, http.StatusCreated, map[string]any{
		"bot_username": botUsername,
		"start_url":    startURL,
		"expires_at":   link.ExpiresAt,
	})
}

func (s *Server) handleDeleteTelegramNotificationLink(w http.ResponseWriter, r *http.Request) {
	if err := s.telegramStore.DeleteRecipient(r.Context(), s.userEmail(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "delete telegram link failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code,omitempty"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid registration request")
		return
	}
	// Email verification code is accepted but ignored until verification is implemented.
	_ = req.Code

	var token string
	var err error
	adminToken := bearerToken(r.Header.Get("Authorization"))
	if s.auth.IsAdminToken(adminToken) {
		token, err = s.auth.RegisterAdmin(r.Context(), adminToken, req.Email, req.Password)
	} else {
		token, err = s.auth.Register(r.Context(), req.Email, req.Password)
	}
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrRegistrationClosed):
			writeError(w, http.StatusForbidden, "registration is disabled")
		case errors.Is(err, auth.ErrUserExists):
			writeError(w, http.StatusConflict, "user already exists")
		case errors.Is(err, auth.ErrWeakPassword):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}
	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]any{"token": token})
}

func (s *Server) handlePairingClaim(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	token := bearerToken(r.Header.Get("Authorization"))
	var req security.PairingRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid pairing request")
		return
	}
	resp, err := s.pairing.Claim(r.Context(), token, req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleCreatePairingCode(w http.ResponseWriter, r *http.Request) {
	code, expiresAt, err := s.pairing.CreateToken(15 * time.Minute)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create pairing code failed")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"code":       code,
		"expires_at": expiresAt,
	})
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2<<20))
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	result, err := s.ingest.Ingest(r.Context(), body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.publishEvent(r.Context(), "snapshot", result)
	writeJSON(w, http.StatusAccepted, map[string]any{"accepted": result.Accepted})
}

func (s *Server) handleServerAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/servers/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "server action not found")
		return
	}
	if parts[1] == "service-actions" {
		s.handleServiceAction(w, r, parts[0])
		return
	}
	if parts[1] != "tasks" {
		writeError(w, http.StatusNotFound, "server action not found")
		return
	}
	var req struct {
		TaskName string   `json:"task_name"`
		Domains  []string `json:"domains"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task request")
		return
	}
	if req.TaskName == "" {
		writeError(w, http.StatusBadRequest, "task_name is required")
		return
	}
	var task tasks.Task
	var err error
	if req.TaskName == "dns-recheck" {
		if len(req.Domains) == 0 {
			writeError(w, http.StatusBadRequest, "domains are required for dns-recheck")
			return
		}
		task, err = s.tasks.EnqueueWithPayload(r.Context(), parts[0], req.TaskName, tasks.TaskPayload{Domains: req.Domains}, s.userEmail(r))
	} else {
		task, err = s.tasks.Enqueue(r.Context(), parts[0], req.TaskName, s.userEmail(r))
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "enqueue task failed")
		return
	}
	writeJSON(w, http.StatusAccepted, task)

	_ = s.audit.Log(r.Context(), domain.AuditLog{
		UserEmail: s.userEmail(r),
		Action:    "task-enqueue",
		Target:    parts[0],
		Details:   fmt.Sprintf("task: %s, domains: %v", req.TaskName, req.Domains),
	})
}

func (s *Server) handleServiceAction(w http.ResponseWriter, r *http.Request, serverID string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Service string `json:"service"`
		Action  string `json:"action"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid service action request")
		return
	}
	req.Service = strings.TrimSpace(req.Service)
	req.Action = strings.TrimSpace(req.Action)
	if req.Service == "" {
		writeError(w, http.StatusBadRequest, "service is required")
		return
	}
	if req.Action != "start" && req.Action != "stop" && req.Action != "restart" {
		writeError(w, http.StatusBadRequest, "action must be start, stop, or restart")
		return
	}
	if !s.serviceAllowsRemoteControl(r.Context(), serverID, req.Service) {
		writeError(w, http.StatusForbidden, "service is not remote-controllable")
		return
	}
	task, err := s.tasks.EnqueueWithPayload(r.Context(), serverID, "service-action", tasks.TaskPayload{Service: req.Service, Action: req.Action}, s.userEmail(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "enqueue service action failed")
		return
	}
	writeJSON(w, http.StatusAccepted, task)

	_ = s.audit.Log(r.Context(), domain.AuditLog{
		UserEmail: s.userEmail(r),
		Action:    "service-action",
		Target:    serverID,
		Details:   fmt.Sprintf("service: %s, action: %s", req.Service, req.Action),
	})
}

func (s *Server) serviceAllowsRemoteControl(ctx context.Context, serverID string, service string) bool {
	state, err := s.store.GetServer(ctx, serverID, time.Now())
	if err != nil {
		return false
	}
	for _, process := range state.Snapshot.Processes {
		if process.Service == service && process.RemoteControl {
			return true
		}
	}
	return false
}

func (s *Server) handlePollTasks(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("agent_id")
	if serverID == "" {
		serverID = r.URL.Query().Get("server_id")
	}
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	if err := s.presence.Touch(r.Context(), serverID, time.Now()); err != nil && r.Context().Err() != nil {
		writeError(w, http.StatusRequestTimeout, "request canceled")
		return
	}
	limit := 1
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 10 {
			limit = parsed
		}
	}
	tasks, err := s.tasks.ClaimPending(r.Context(), serverID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "poll tasks failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": jsonSlice(tasks)})
}

func (s *Server) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/agent/tasks/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "result" {
		writeError(w, http.StatusNotFound, "task action not found")
		return
	}
	var req tasks.TaskResult
	defer r.Body.Close()
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid task result")
		return
	}
	task, err := s.tasks.Complete(r.Context(), parts[0], req)
	if err != nil {
		var notFound tasks.ErrNotFound
		var invalid tasks.ErrInvalidState
		switch {
		case errors.As(err, &notFound):
			writeError(w, http.StatusNotFound, "task not found")
		case errors.As(err, &invalid):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "complete task failed")
		}
		return
	}
	s.publishEvent(r.Context(), "task_completed", task)
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	alerts, err := s.alerts.Recent(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list alerts failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"alerts": jsonSlice(alerts)})
}

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.store.ListServers(r.Context(), time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list servers failed")
		return
	}
	servers = s.presence.ApplySummaries(r.Context(), servers, time.Now())
	writeJSON(w, http.StatusOK, map[string]any{"servers": jsonSlice(servers)})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	tasks, err := s.tasks.List(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list tasks failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": jsonSlice(tasks)})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if taskID == "" || strings.Contains(taskID, "/") {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	task, err := s.tasks.Get(r.Context(), taskID)
	if err != nil {
		var notFound tasks.ErrNotFound
		if errors.As(err, &notFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get task failed")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/servers/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	state, err := s.store.GetServer(r.Context(), id, time.Now())
	if err != nil {
		var notFound store.ErrNotFound
		if errors.As(err, &notFound) {
			writeError(w, http.StatusNotFound, "server not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get server failed")
		return
	}
	state.Summary = s.presence.ApplySummary(r.Context(), state.Summary, time.Now())
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	to := time.Now().UTC()
	if raw := r.URL.Query().Get("to"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			to = t
		}
	}
	from := to.Add(-1 * time.Hour)
	if raw := r.URL.Query().Get("from"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			from = t
		}
	}
	metrics, err := s.store.GetMetrics(r.Context(), id, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get metrics failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"metrics": metrics})
}

func (s *Server) handleGetServerConfig(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	cfg, err := s.configStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, serverconfig.ErrNotFound) {
			writeJSON(w, http.StatusOK, domain.AgentDesiredConfig{})
			return
		}
		writeError(w, http.StatusInternalServerError, "get server config failed")
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleSetServerConfig(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	defer r.Body.Close()
	var cfg domain.AgentDesiredConfig
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid config")
		return
	}
	if err := validateAgentDesiredConfig(cfg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	cfg.UpdatedAt = time.Now().UTC()
	saved, err := s.saveDesiredConfig(r.Context(), id, cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "save server config failed")
		return
	}
	writeJSON(w, http.StatusOK, saved)

	_ = s.audit.Log(r.Context(), domain.AuditLog{
		UserEmail: s.userEmail(r),
		Action:    "config-change",
		Target:    id,
		Details:   "updated agent desired config",
	})
}

func (s *Server) saveDesiredConfig(ctx context.Context, serverID string, cfg domain.AgentDesiredConfig) (domain.AgentDesiredConfig, error) {
	if err := s.configStore.Set(ctx, serverID, cfg); err != nil {
		return domain.AgentDesiredConfig{}, err
	}
	saved, err := s.configStore.Get(ctx, serverID)
	if err != nil {
		return domain.AgentDesiredConfig{}, err
	}
	_ = s.store.UpdateDesiredRevision(ctx, serverID, saved.Revision)
	return saved, nil
}

func validateAgentDesiredConfig(cfg domain.AgentDesiredConfig) error {
	if cfg.Remote.ShellEnabled {
		return errors.New("remote shell is not allowed until cloud-side authorization is implemented")
	}
	if cfg.Agent.Interval < 0 {
		return errors.New("agent interval must not be negative")
	}
	if cfg.Watchdog.PollingSeconds < 0 {
		return errors.New("watchdog polling_seconds must not be negative")
	}
	if cfg.Watchdog.TimeoutSeconds < 0 {
		return errors.New("watchdog timeout_seconds must not be negative")
	}
	for _, proc := range cfg.Processes {
		if proc.Name == "" {
			return errors.New("process name is required")
		}
	}
	for _, check := range cfg.Network.DNSChecks {
		if check.Name == "" || check.Domain == "" {
			return errors.New("dns check needs name and domain")
		}
	}
	for _, check := range cfg.Network.PortChecks {
		if check.Name == "" || check.Address == "" {
			return errors.New("port check needs name and address")
		}
	}
	for _, test := range cfg.Network.SpeedTests {
		if test.Name == "" || test.URL == "" {
			return errors.New("speed test needs name and url")
		}
	}
	for _, stream := range cfg.LogStreams {
		if stream.Name == "" || stream.Path == "" {
			return errors.New("log stream needs name and path")
		}
	}
	return nil
}

func (s *Server) handleGetAgentConfig(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("agent_id")
	if serverID == "" {
		serverID = r.URL.Query().Get("server_id")
	}
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	cfg, err := s.configStore.Get(r.Context(), serverID)
	if err != nil {
		if errors.Is(err, serverconfig.ErrNotFound) {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get agent config failed")
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) requireAgent(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.hasVerifiedClientCertificate(r) {
			next(w, r)
			return
		}
		if s.cfg.TLS.RequireClientCert {
			writeError(w, http.StatusUnauthorized, "verified client certificate required")
			return
		}
		token := s.getToken(r)
		if !s.cfg.Auth.AllowsIngest(token) {
			writeError(w, http.StatusUnauthorized, "invalid ingest token")
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.getToken(r)
		if session, ok := s.auth.ValidateToken(token); ok {
			if s.auth.IsAdmin(session) {
				next(w, r)
				return
			}
			writeError(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		if session, ok, err := s.auth.LoadSession(r.Context(), token); err == nil && ok {
			if s.auth.IsAdmin(session) {
				next(w, r)
				return
			}
			writeError(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		if s.auth.IsAdminToken(token) {
			next(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "authentication required")
	}
}

func (s *Server) requirePlus(next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		token := s.getToken(r)
		if s.auth.IsAdminToken(token) {
			next(w, r)
			return
		}
		session, ok := s.currentSession(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if s.subscriptionForSession(r.Context(), session).Plan != domain.PlanPlus {
			writePaymentRequired(w, "This action requires Trace Plus.")
			return
		}
		next(w, r)
	})
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.getToken(r)
		if _, ok := s.auth.ValidateToken(token); ok {
			next(w, r)
			return
		}
		if _, ok, err := s.auth.LoadSession(r.Context(), token); err == nil && ok {
			next(w, r)
			return
		}
		if s.auth.IsAdminToken(token) {
			next(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "authentication required")
	}
}

func (s *Server) currentSession(r *http.Request) (auth.Session, bool) {
	token := s.getToken(r)
	if session, ok := s.auth.ValidateToken(token); ok {
		session.Plan = domain.NormalizePlan(session.Plan)
		return session, true
	}
	if session, ok, err := s.auth.LoadSession(r.Context(), token); err == nil && ok {
		session.Plan = domain.NormalizePlan(session.Plan)
		return session, true
	}
	if s.auth.IsAdminToken(token) {
		return auth.Session{Email: "admin-token", Role: domain.RoleAdmin, Plan: domain.PlanPlus, ExpiresAt: time.Now().Add(time.Hour)}, true
	}
	return auth.Session{}, false
}

func (s *Server) subscriptionForSession(ctx context.Context, session auth.Session) domain.Subscription {
	if session.Email != "" && session.Email != "admin-token" {
		if user, err := s.auth.User(ctx, session.Email); err == nil {
			return domain.EntitlementsForPlan(user.Plan)
		}
	}
	return domain.EntitlementsForPlan(session.Plan)
}

func (s *Server) requireOwner(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.getToken(r)
		if session, ok := s.auth.ValidateToken(token); ok {
			if s.auth.IsOwner(session) {
				next(w, r)
				return
			}
			writeError(w, http.StatusForbidden, "owner permissions required")
			return
		}
		if session, ok, err := s.auth.LoadSession(r.Context(), token); err == nil && ok {
			if s.auth.IsOwner(session) {
				next(w, r)
				return
			}
			writeError(w, http.StatusForbidden, "owner permissions required")
			return
		}
		writeError(w, http.StatusUnauthorized, "authentication required")
	}
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if s.applyCORS(w, r) && r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) applyCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" || len(s.cfg.HTTP.AllowedOrigins) == 0 {
		return false
	}
	allowedOrigin, ok := s.allowedOrigin(origin)
	if !ok {
		if r.Method == http.MethodOptions {
			writeError(w, http.StatusForbidden, "origin is not allowed")
			return true
		}
		return false
	}
	header := w.Header()
	header.Set("Vary", "Origin")
	header.Set("Access-Control-Allow-Origin", allowedOrigin)
	header.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Set("Access-Control-Max-Age", "600")
	return true
}

func (s *Server) allowedOrigin(origin string) (string, bool) {
	for _, allowed := range s.cfg.HTTP.AllowedOrigins {
		if allowed == "*" {
			return "*", true
		}
		if allowed == origin {
			return origin, true
		}
	}
	return "", false
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.pubsub.Subscribe()
	defer s.pubsub.Unsubscribe(ch)

	for {
		select {
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) publishEvent(ctx context.Context, eventType string, data any) {
	payload, _ := json.Marshal(map[string]any{
		"type": eventType,
		"data": data,
	})
	_ = s.pubsub.Publish(ctx, "events", payload)
}

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	logs, err := s.audit.Recent(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list audit logs failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"audit_logs": jsonSlice(logs)})
}

func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	incidents, err := s.incidents.ListIncidents(r.Context(), serverID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list incidents failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"incidents": jsonSlice(incidents)})
}

func (s *Server) handleGetIncidentActions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"actions": incidents.AvailableActions()})
}

func (s *Server) handleIncidentMetrics(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	window, err := parseIncidentMetricsWindow(r.URL.Query().Get("window"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	metrics, err := s.incidents.Metrics(r.Context(), serverID, window)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "incident metrics failed")
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/incidents/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "incident not found")
		return
	}
	incident, err := s.incidents.GetIncident(r.Context(), id)
	if err != nil {
		if errors.Is(err, incidents.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incident not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get incident failed")
		return
	}
	writeJSON(w, http.StatusOK, incident)
}

func (s *Server) handleIncidentAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/incidents/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, http.StatusNotFound, "incident action not found")
		return
	}
	incidentID := parts[0]
	action := parts[1]

	// Validate action
	validActions := map[string]bool{
		"restart":          true,
		"disable-watchdog": true,
		"diagnostics":      true,
		"rollback-config":  true,
	}
	if !validActions[action] {
		writeError(w, http.StatusBadRequest, "invalid action")
		return
	}

	// Get incident to find server and service
	incident, err := s.incidents.GetIncident(r.Context(), incidentID)
	if err != nil {
		if errors.Is(err, incidents.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incident not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get incident failed")
		return
	}

	actor := s.userEmail(r)

	// Execute action
	switch action {
	case "restart":
		// Enqueue service-action task
		_, err = s.tasks.EnqueueWithPayload(r.Context(), incident.ServerID, "service-action",
			tasks.TaskPayload{Service: incident.ServiceName, Action: "restart"}, actor)
	case "diagnostics":
		_, err = s.tasks.EnqueueWithPayload(r.Context(), incident.ServerID, "diagnostics",
			tasks.TaskPayload{Service: incident.ServiceName, IncidentID: incident.ID}, actor)
	case "disable-watchdog":
		// Update desired config to disable watchdog for this service
		cfg, err := s.configStore.Get(r.Context(), incident.ServerID)
		if err != nil && !errors.Is(err, serverconfig.ErrNotFound) {
			writeError(w, http.StatusInternalServerError, "get config failed")
			return
		}
		// Update process config
		found := false
		for i, proc := range cfg.Processes {
			if proc.Name == incident.ServiceName || proc.Service == incident.ServiceName {
				cfg.Processes[i].Restart = false
				found = true
				break
			}
		}
		if !found {
			writeError(w, http.StatusBadRequest, "service not found in config")
			return
		}
		cfg.UpdatedAt = time.Now().UTC()
		if _, err := s.saveDesiredConfig(r.Context(), incident.ServerID, cfg); err != nil {
			writeError(w, http.StatusInternalServerError, "update config failed")
			return
		}
	case "rollback-config":
		cfg, err := s.configStore.Previous(r.Context(), incident.ServerID)
		if err != nil {
			if errors.Is(err, serverconfig.ErrNotFound) {
				writeError(w, http.StatusBadRequest, "no previous config to rollback")
				return
			}
			writeError(w, http.StatusInternalServerError, "get previous config failed")
			return
		}
		cfg.UpdatedAt = time.Now().UTC()
		if _, err := s.saveDesiredConfig(r.Context(), incident.ServerID, cfg); err != nil {
			writeError(w, http.StatusInternalServerError, "rollback config failed")
			return
		}
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "action execution failed")
		return
	}

	// Record action in incident timeline
	if err := s.incidents.ExecuteAction(r.Context(), incidentID, action, actor); err != nil {
		s.logger.Warn("failed to record incident action", zap.Error(err))
	}

	// Audit log
	_ = s.audit.Log(r.Context(), domain.AuditLog{
		UserEmail: actor,
		Action:    "incident-action",
		Target:    incidentID,
		Details:   fmt.Sprintf("action: %s, service: %s", action, incident.ServiceName),
	})

	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "action": action})
}

func (s *Server) handleAnalyzeIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "incident id is required")
		return
	}

	incident, err := s.incidents.GetIncident(r.Context(), id)
	if err != nil {
		if errors.Is(err, incidents.ErrNotFound) {
			writeError(w, http.StatusNotFound, "incident not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get incident failed")
		return
	}

	analysisContext := ai.IncidentContext{}
	if state, err := s.store.GetServer(r.Context(), incident.ServerID, time.Now()); err == nil {
		analysisContext.Server = &state
	}
	from := incident.CreatedAt.Add(-15 * time.Minute)
	to := time.Now().Add(5 * time.Minute)
	if metrics, err := s.store.GetMetrics(r.Context(), incident.ServerID, from, to); err == nil {
		analysisContext.Metrics = metrics
	}

	analysis, err := s.aiAnalyzer.AnalyzeIncident(r.Context(), incident, analysisContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("AI analysis failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, analysis)
}

func parseIncidentMetricsWindow(raw string) (time.Duration, error) {
	if raw == "" {
		return 7 * 24 * time.Hour, nil
	}
	if strings.HasSuffix(raw, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(raw, "d"))
		if err != nil || days <= 0 || days > 365 {
			return 0, fmt.Errorf("window must be between 1d and 365d")
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	window, err := time.ParseDuration(raw)
	if err != nil || window <= 0 || window > 365*24*time.Hour {
		return 0, fmt.Errorf("window must be a positive duration up to 365d")
	}
	return window, nil
}

func (s *Server) userEmail(r *http.Request) string {
	token := s.getToken(r)
	if session, ok := s.auth.ValidateToken(token); ok {
		return session.Email
	}
	if session, ok, err := s.auth.LoadSession(r.Context(), token); err == nil && ok {
		return session.Email
	}
	if s.auth.IsAdminToken(token) {
		return "admin-token"
	}
	return "unknown"
}

func (s *Server) getToken(r *http.Request) string {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value
	}
	return bearerToken(r.Header.Get("Authorization"))
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.cfg.Auth.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.TLS.Enabled,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.TLS.Enabled,
		SameSite: http.SameSiteLaxMode,
	})
}

func bearerToken(value string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func jsonSlice[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func writePaymentRequired(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusPaymentRequired, map[string]any{
		"error":         message,
		"code":          "plan_required",
		"required_plan": domain.PlanPlus,
	})
}

func URLForServer(id string) string {
	return fmt.Sprintf("/v1/servers/%s", id)
}

func (s *Server) tlsConfig() *tls.Config {
	if !s.cfg.TLS.Enabled {
		return nil
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if s.cfg.TLS.ClientCAFile != "" {
		pool := x509.NewCertPool()
		caPEM, err := os.ReadFile(s.cfg.TLS.ClientCAFile)
		if err == nil && pool.AppendCertsFromPEM(caPEM) {
			cfg.ClientCAs = pool
			cfg.ClientAuth = tls.VerifyClientCertIfGiven
		}
	}
	return cfg
}

func listenAndServe(server *http.Server, cfg config.Config) error {
	if cfg.TLS.Enabled {
		return server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	}
	return server.ListenAndServe()
}

func (s *Server) hasVerifiedClientCertificate(r *http.Request) bool {
	if r.TLS != nil && len(r.TLS.VerifiedChains) > 0 && len(r.TLS.PeerCertificates) > 0 {
		return true
	}
	// This header is written, not forwarded, by the private Nginx reverse proxy.
	// It is enabled only when that proxy is the sole route to the backend.
	return s.cfg.TLS.TrustProxyClientCert && r.Header.Get("X-Trace-Client-Verify") == "SUCCESS"
}

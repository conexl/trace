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

	"backend/internal/config"
	"backend/internal/ingest"
	"backend/internal/security"
	"backend/internal/store"
	"backend/internal/tasks"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("httpapi", fx.Provide(NewServer), fx.Invoke(RegisterLifecycle))

type Server struct {
	cfg     config.Config
	store   store.Store
	ingest  *ingest.Service
	pairing *security.PairingService
	tasks   tasks.Store
	logger  *zap.Logger
	mux     *http.ServeMux
}

func NewServer(cfg config.Config, store store.Store, ingest *ingest.Service, pairing *security.PairingService, taskStore tasks.Store, logger *zap.Logger) *Server {
	server := &Server{cfg: cfg, store: store, ingest: ingest, pairing: pairing, tasks: taskStore, logger: logger.Named("http"), mux: http.NewServeMux()}
	server.routes()
	return server
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:         s.cfg.HTTP.Addr,
		Handler:      s.securityHeaders(s.mux),
		ReadTimeout:  s.cfg.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.HTTP.WriteTimeout,
		TLSConfig:    s.tlsConfig(),
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("POST /v1/pairing/claim", s.handlePairingClaim)
	s.mux.HandleFunc("POST /v1/agent/snapshots", s.requireAgent(s.handleIngest))
	s.mux.HandleFunc("GET /v1/agent/tasks", s.requireAgent(s.handlePollTasks))
	s.mux.HandleFunc("POST /v1/agent/tasks/", s.requireAgent(s.handleCompleteTask))
	s.mux.HandleFunc("GET /v1/servers", s.requireAdmin(s.handleListServers))
	s.mux.HandleFunc("POST /v1/servers/", s.requireAdmin(s.handleServerAction))
	s.mux.HandleFunc("GET /v1/tasks/", s.requireAdmin(s.handleGetTask))
	s.mux.HandleFunc("GET /v1/servers/", s.requireAdmin(s.handleGetServer))
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
	writeJSON(w, http.StatusAccepted, map[string]any{"accepted": result.Accepted})
}

func (s *Server) handleServerAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/servers/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "tasks" {
		writeError(w, http.StatusNotFound, "server action not found")
		return
	}
	var req struct {
		TaskName string `json:"task_name"`
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
	task, err := s.tasks.Enqueue(r.Context(), parts[0], req.TaskName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "enqueue task failed")
		return
	}
	writeJSON(w, http.StatusAccepted, task)
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
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
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
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.store.ListServers(r.Context(), time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list servers failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"servers": servers})
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
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) requireAgent(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hasVerifiedClientCertificate(r) {
			next(w, r)
			return
		}
		if s.cfg.TLS.RequireClientCert {
			writeError(w, http.StatusUnauthorized, "verified client certificate required")
			return
		}
		token := bearerToken(r.Header.Get("Authorization"))
		if !s.cfg.Auth.AllowsIngest(token) {
			writeError(w, http.StatusUnauthorized, "invalid ingest token")
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.cfg.Auth.RequiresAdmin() {
			next(w, r)
			return
		}
		if bearerToken(r.Header.Get("Authorization")) != s.cfg.Auth.AdminToken {
			writeError(w, http.StatusUnauthorized, "invalid admin token")
			return
		}
		next(w, r)
	}
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
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

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
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

func hasVerifiedClientCertificate(r *http.Request) bool {
	return r.TLS != nil && len(r.TLS.VerifiedChains) > 0 && len(r.TLS.PeerCertificates) > 0
}

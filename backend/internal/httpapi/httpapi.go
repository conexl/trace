package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/ingest"
	"backend/internal/store"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("httpapi", fx.Provide(NewServer), fx.Invoke(RegisterLifecycle))

type Server struct {
	cfg    config.Config
	store  store.Store
	ingest *ingest.Service
	logger *zap.Logger
	mux    *http.ServeMux
}

func NewServer(cfg config.Config, store store.Store, ingest *ingest.Service, logger *zap.Logger) *Server {
	server := &Server{cfg: cfg, store: store, ingest: ingest, logger: logger.Named("http"), mux: http.NewServeMux()}
	server.routes()
	return server
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:         s.cfg.HTTP.Addr,
		Handler:      s.securityHeaders(s.mux),
		ReadTimeout:  s.cfg.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.HTTP.WriteTimeout,
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("POST /v1/agent/snapshots", s.requireIngest(s.handleIngest))
	s.mux.HandleFunc("GET /v1/servers", s.requireAdmin(s.handleListServers))
	s.mux.HandleFunc("GET /v1/servers/", s.requireAdmin(s.handleGetServer))
}

func RegisterLifecycle(lc fx.Lifecycle, api *Server, logger *zap.Logger, cfg config.Config) {
	httpServer := api.HTTPServer()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("http server listening", zap.String("addr", httpServer.Addr))
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.store.ListServers(r.Context(), time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list servers failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"servers": servers})
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

func (s *Server) requireIngest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

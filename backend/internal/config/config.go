package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/fx"
)

var Module = fx.Module("config", fx.Provide(Load))

type Config struct {
	HTTP        HTTPConfig
	Auth        AuthConfig
	State       StateConfig
	Mongo       MongoConfig
	Environment string
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type AuthConfig struct {
	IngestTokens map[string]struct{}
	AdminToken   string
}

type StateConfig struct {
	OfflineAfter time.Duration
	MaxEvents    int
}

type MongoConfig struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Addr:            env("HOMELYTICS_HTTP_ADDR", ":8080"),
			ReadTimeout:     envDuration("HOMELYTICS_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("HOMELYTICS_HTTP_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: envDuration("HOMELYTICS_HTTP_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		Auth: AuthConfig{
			IngestTokens: parseTokenSet(os.Getenv("HOMELYTICS_INGEST_TOKENS")),
			AdminToken:   os.Getenv("HOMELYTICS_ADMIN_TOKEN"),
		},
		State: StateConfig{
			OfflineAfter: envDuration("HOMELYTICS_OFFLINE_AFTER", 3*time.Minute),
			MaxEvents:    envInt("HOMELYTICS_MAX_EVENTS", 200),
		},
		Mongo: MongoConfig{
			URI:            os.Getenv("HOMELYTICS_MONGO_URI"),
			Database:       env("HOMELYTICS_MONGO_DATABASE", "homelytics"),
			ConnectTimeout: envDuration("HOMELYTICS_MONGO_CONNECT_TIMEOUT", 5*time.Second),
		},
		Environment: env("HOMELYTICS_ENV", "development"),
	}
	if cfg.State.MaxEvents <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_MAX_EVENTS must be positive")
	}
	if cfg.State.OfflineAfter <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_OFFLINE_AFTER must be positive")
	}
	if cfg.Mongo.Database == "" {
		return Config{}, fmt.Errorf("HOMELYTICS_MONGO_DATABASE must not be empty")
	}
	if cfg.Mongo.ConnectTimeout <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_MONGO_CONNECT_TIMEOUT must be positive")
	}
	return cfg, nil
}

func (c AuthConfig) AllowsIngest(token string) bool {
	if len(c.IngestTokens) == 0 {
		return true
	}
	_, ok := c.IngestTokens[token]
	return ok
}

func (c AuthConfig) RequiresAdmin() bool {
	return c.AdminToken != ""
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseTokenSet(raw string) map[string]struct{} {
	values := strings.Split(raw, ",")
	tokens := make(map[string]struct{})
	for _, value := range values {
		token := strings.TrimSpace(value)
		if token != "" {
			tokens[token] = struct{}{}
		}
	}
	return tokens
}

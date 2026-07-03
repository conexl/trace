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
	TLS         TLSConfig
	Auth        AuthConfig
	Pairing     PairingConfig
	State       StateConfig
	Mongo       MongoConfig
	Redis       RedisConfig
	Alerts      AlertsConfig
	Environment string
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type TLSConfig struct {
	Enabled           bool
	CertFile          string
	KeyFile           string
	ClientCAFile      string
	RequireClientCert bool
}

type AuthConfig struct {
	IngestTokens map[string]struct{}
	AdminToken   string
}

type PairingConfig struct {
	Tokens     map[string]struct{}
	CACertFile string
	CAKeyFile  string
	CertTTL    time.Duration
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

type RedisConfig struct {
	Addr           string
	Password       string
	DB             int
	KeyPrefix      string
	ConnectTimeout time.Duration
}

type AlertsConfig struct {
	MemoryLimit      int
	TelegramBotToken string
	TelegramChatID   string
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Addr:            env("HOMELYTICS_HTTP_ADDR", ":8080"),
			ReadTimeout:     envDuration("HOMELYTICS_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("HOMELYTICS_HTTP_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: envDuration("HOMELYTICS_HTTP_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		TLS: TLSConfig{
			Enabled:           envBool("HOMELYTICS_TLS_ENABLED", false),
			CertFile:          os.Getenv("HOMELYTICS_TLS_CERT_FILE"),
			KeyFile:           os.Getenv("HOMELYTICS_TLS_KEY_FILE"),
			ClientCAFile:      os.Getenv("HOMELYTICS_TLS_CLIENT_CA_FILE"),
			RequireClientCert: envBool("HOMELYTICS_TLS_REQUIRE_CLIENT_CERT", false),
		},
		Auth: AuthConfig{
			IngestTokens: parseTokenSet(os.Getenv("HOMELYTICS_INGEST_TOKENS")),
			AdminToken:   os.Getenv("HOMELYTICS_ADMIN_TOKEN"),
		},
		Pairing: PairingConfig{
			Tokens:     parseTokenSet(os.Getenv("HOMELYTICS_PAIRING_TOKENS")),
			CACertFile: os.Getenv("HOMELYTICS_PAIRING_CA_CERT_FILE"),
			CAKeyFile:  os.Getenv("HOMELYTICS_PAIRING_CA_KEY_FILE"),
			CertTTL:    envDuration("HOMELYTICS_PAIRING_CERT_TTL", 24*time.Hour),
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
		Redis: RedisConfig{
			Addr:           os.Getenv("HOMELYTICS_REDIS_ADDR"),
			Password:       os.Getenv("HOMELYTICS_REDIS_PASSWORD"),
			DB:             envInt("HOMELYTICS_REDIS_DB", 0),
			KeyPrefix:      env("HOMELYTICS_REDIS_KEY_PREFIX", "homelytics"),
			ConnectTimeout: envDuration("HOMELYTICS_REDIS_CONNECT_TIMEOUT", 3*time.Second),
		},
		Alerts: AlertsConfig{
			MemoryLimit:      envInt("HOMELYTICS_ALERT_MEMORY_LIMIT", 200),
			TelegramBotToken: os.Getenv("HOMELYTICS_TELEGRAM_BOT_TOKEN"),
			TelegramChatID:   os.Getenv("HOMELYTICS_TELEGRAM_CHAT_ID"),
		},
		Environment: env("HOMELYTICS_ENV", "development"),
	}
	if cfg.TLS.Enabled && (cfg.TLS.CertFile == "" || cfg.TLS.KeyFile == "") {
		return Config{}, fmt.Errorf("HOMELYTICS_TLS_CERT_FILE and HOMELYTICS_TLS_KEY_FILE are required when TLS is enabled")
	}
	if cfg.TLS.RequireClientCert && cfg.TLS.ClientCAFile == "" {
		return Config{}, fmt.Errorf("HOMELYTICS_TLS_CLIENT_CA_FILE is required when client certs are required")
	}
	if cfg.Pairing.CertTTL <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_PAIRING_CERT_TTL must be positive")
	}
	if cfg.Environment == "production" && len(cfg.Pairing.Tokens) > 0 && (cfg.Pairing.CACertFile == "" || cfg.Pairing.CAKeyFile == "") {
		return Config{}, fmt.Errorf("pairing CA files are required in production")
	}
	if cfg.State.MaxEvents <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_MAX_EVENTS must be positive")
	}
	if cfg.State.OfflineAfter <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_OFFLINE_AFTER must be positive")
	}
	if cfg.Alerts.MemoryLimit <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_ALERT_MEMORY_LIMIT must be positive")
	}
	if cfg.Mongo.Database == "" {
		return Config{}, fmt.Errorf("HOMELYTICS_MONGO_DATABASE must not be empty")
	}
	if cfg.Mongo.ConnectTimeout <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_MONGO_CONNECT_TIMEOUT must be positive")
	}
	if cfg.Redis.DB < 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_REDIS_DB must not be negative")
	}
	if cfg.Redis.KeyPrefix == "" {
		return Config{}, fmt.Errorf("HOMELYTICS_REDIS_KEY_PREFIX must not be empty")
	}
	if cfg.Redis.ConnectTimeout <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_REDIS_CONNECT_TIMEOUT must be positive")
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

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
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

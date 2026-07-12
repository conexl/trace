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
	HTTP          HTTPConfig
	TLS           TLSConfig
	Auth          AuthConfig
	Pairing       PairingConfig
	State         StateConfig
	Mongo         MongoConfig
	Redis         RedisConfig
	Alerts        AlertsConfig
	Notifications NotificationsConfig
	AI            AIConfig
	Environment   string
}

type HTTPConfig struct {
	Addr                  string
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	ShutdownTimeout       time.Duration
	AllowedOrigins        []string
	TrustForwardedHeaders bool
}

type TLSConfig struct {
	Enabled              bool
	CertFile             string
	KeyFile              string
	ClientCAFile         string
	RequireClientCert    bool
	TrustProxyClientCert bool
}

type AuthConfig struct {
	IngestTokens         map[string]struct{}
	AdminToken           string
	RegistrationDisabled bool
	BootstrapAdminEmail  string
	LoginRateLimit       int
	LoginRateWindow      time.Duration
	RegisterRateLimit    int
	RegisterRateWindow   time.Duration
	SessionTTL           time.Duration
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

type NotificationsConfig struct {
	EventChannel         string
	SendTimeout          time.Duration
	LinkTTL              time.Duration
	TelegramBotUsername  string
	TelegramBotToken     string
	TelegramChatID       string
	TelegramPollInterval time.Duration
	TelegramPollTimeout  time.Duration
}

type AIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Addr:                  env("HOMELYTICS_HTTP_ADDR", ":8080"),
			ReadTimeout:           envDuration("HOMELYTICS_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:          envDuration("HOMELYTICS_HTTP_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout:       envDuration("HOMELYTICS_HTTP_SHUTDOWN_TIMEOUT", 5*time.Second),
			AllowedOrigins:        parseCSV(os.Getenv("HOMELYTICS_CORS_ALLOWED_ORIGINS")),
			TrustForwardedHeaders: envBool("HOMELYTICS_TRUST_FORWARDED_HEADERS", false),
		},
		TLS: TLSConfig{
			Enabled:              envBool("HOMELYTICS_TLS_ENABLED", false),
			CertFile:             os.Getenv("HOMELYTICS_TLS_CERT_FILE"),
			KeyFile:              os.Getenv("HOMELYTICS_TLS_KEY_FILE"),
			ClientCAFile:         os.Getenv("HOMELYTICS_TLS_CLIENT_CA_FILE"),
			RequireClientCert:    envBool("HOMELYTICS_TLS_REQUIRE_CLIENT_CERT", false),
			TrustProxyClientCert: envBool("HOMELYTICS_TRUST_PROXY_CLIENT_CERT", false),
		},
		Auth: AuthConfig{
			IngestTokens:         parseTokenSet(os.Getenv("HOMELYTICS_INGEST_TOKENS")),
			AdminToken:           os.Getenv("HOMELYTICS_ADMIN_TOKEN"),
			RegistrationDisabled: envBool("HOMELYTICS_REGISTRATION_DISABLED", false),
			BootstrapAdminEmail:  strings.ToLower(strings.TrimSpace(os.Getenv("HOMELYTICS_BOOTSTRAP_ADMIN_EMAIL"))),
			LoginRateLimit:       envInt("HOMELYTICS_LOGIN_RATE_LIMIT", 10),
			LoginRateWindow:      envDuration("HOMELYTICS_LOGIN_RATE_WINDOW", time.Minute),
			RegisterRateLimit:    envInt("HOMELYTICS_REGISTER_RATE_LIMIT", 5),
			RegisterRateWindow:   envDuration("HOMELYTICS_REGISTER_RATE_WINDOW", time.Hour),
			SessionTTL:           envDuration("HOMELYTICS_SESSION_TTL", 24*time.Hour),
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
		Notifications: NotificationsConfig{
			EventChannel:         env("HOMELYTICS_NOTIFICATIONS_EVENT_CHANNEL", "events"),
			SendTimeout:          envDuration("HOMELYTICS_NOTIFICATIONS_SEND_TIMEOUT", 5*time.Second),
			LinkTTL:              envDuration("HOMELYTICS_NOTIFICATIONS_LINK_TTL", 10*time.Minute),
			TelegramBotUsername:  strings.TrimPrefix(strings.TrimSpace(os.Getenv("HOMELYTICS_NOTIFICATIONS_TELEGRAM_BOT_USERNAME")), "@"),
			TelegramBotToken:     env("HOMELYTICS_NOTIFICATIONS_TELEGRAM_BOT_TOKEN", os.Getenv("HOMELYTICS_TELEGRAM_BOT_TOKEN")),
			TelegramChatID:       env("HOMELYTICS_NOTIFICATIONS_TELEGRAM_CHAT_ID", os.Getenv("HOMELYTICS_TELEGRAM_CHAT_ID")),
			TelegramPollInterval: envDuration("HOMELYTICS_NOTIFICATIONS_TELEGRAM_POLL_INTERVAL", time.Second),
			TelegramPollTimeout:  envDuration("HOMELYTICS_NOTIFICATIONS_TELEGRAM_POLL_TIMEOUT", 25*time.Second),
		},
		AI: AIConfig{
			APIKey:  os.Getenv("AI_API_KEY"),
			BaseURL: os.Getenv("AI_BASE_URL"),
			Model:   os.Getenv("AI_MODEL"),
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
	if cfg.Environment == "production" {
		if !cfg.TLS.Enabled {
			return Config{}, fmt.Errorf("production requires TLS to be enabled (HOMELYTICS_TLS_ENABLED=true)")
		}
		for _, origin := range cfg.HTTP.AllowedOrigins {
			if origin == "*" {
				return Config{}, fmt.Errorf("production does not allow CORS origin '*' (HOMELYTICS_CORS_ALLOWED_ORIGINS)")
			}
		}
		if !cfg.Auth.RegistrationDisabled {
			// We allow it but it's dangerous. For now, let's just require it to be explicit.
		}

		hasMTLS := cfg.TLS.RequireClientCert && cfg.TLS.ClientCAFile != ""
		hasIngestToken := len(cfg.Auth.IngestTokens) > 0
		if !hasMTLS && !hasIngestToken {
			return Config{}, fmt.Errorf("production requires either mTLS (HOMELYTICS_TLS_REQUIRE_CLIENT_CERT=true and HOMELYTICS_TLS_CLIENT_CA_FILE) or ingest tokens (HOMELYTICS_INGEST_TOKENS)")
		}
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
	if cfg.Notifications.EventChannel == "" {
		return Config{}, fmt.Errorf("HOMELYTICS_NOTIFICATIONS_EVENT_CHANNEL must not be empty")
	}
	if cfg.Notifications.SendTimeout <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_NOTIFICATIONS_SEND_TIMEOUT must be positive")
	}
	if cfg.Notifications.LinkTTL <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_NOTIFICATIONS_LINK_TTL must be positive")
	}
	if cfg.Notifications.TelegramPollInterval <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_NOTIFICATIONS_TELEGRAM_POLL_INTERVAL must be positive")
	}
	if cfg.Notifications.TelegramPollTimeout <= 0 {
		return Config{}, fmt.Errorf("HOMELYTICS_NOTIFICATIONS_TELEGRAM_POLL_TIMEOUT must be positive")
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
	tokens := make(map[string]struct{})
	for _, value := range parseCSV(raw) {
		token := strings.TrimSpace(value)
		if token != "" {
			tokens[token] = struct{}{}
		}
	}
	return tokens
}

func parseCSV(raw string) []string {
	values := strings.Split(raw, ",")
	out := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

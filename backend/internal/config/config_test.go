package config

import (
	"os"
	"testing"
)

func TestLoadProductionRequiresAuth(t *testing.T) {
	os.Setenv("HOMELYTICS_ENV", "production")
	defer os.Unsetenv("HOMELYTICS_ENV")
	os.Unsetenv("HOMELYTICS_TLS_REQUIRE_CLIENT_CERT")
	os.Unsetenv("HOMELYTICS_TLS_CLIENT_CA_FILE")
	os.Unsetenv("HOMELYTICS_INGEST_TOKENS")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error in production without mTLS or ingest token")
	}
}

func TestLoadProductionAllowsIngestToken(t *testing.T) {
	os.Setenv("HOMELYTICS_ENV", "production")
	defer os.Unsetenv("HOMELYTICS_ENV")
	os.Setenv("HOMELYTICS_TLS_ENABLED", "true")
	defer os.Unsetenv("HOMELYTICS_TLS_ENABLED")
	os.Setenv("HOMELYTICS_TLS_CERT_FILE", "cert.pem")
	defer os.Unsetenv("HOMELYTICS_TLS_CERT_FILE")
	os.Setenv("HOMELYTICS_TLS_KEY_FILE", "key.pem")
	defer os.Unsetenv("HOMELYTICS_TLS_KEY_FILE")
	os.Setenv("HOMELYTICS_INGEST_TOKENS", "tok")
	defer os.Unsetenv("HOMELYTICS_INGEST_TOKENS")

	_, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadProductionAllowsMTLS(t *testing.T) {
	os.Setenv("HOMELYTICS_ENV", "production")
	defer os.Unsetenv("HOMELYTICS_ENV")
	os.Setenv("HOMELYTICS_TLS_ENABLED", "true")
	defer os.Unsetenv("HOMELYTICS_TLS_ENABLED")
	os.Setenv("HOMELYTICS_TLS_CERT_FILE", "cert.pem")
	defer os.Unsetenv("HOMELYTICS_TLS_CERT_FILE")
	os.Setenv("HOMELYTICS_TLS_KEY_FILE", "key.pem")
	defer os.Unsetenv("HOMELYTICS_TLS_KEY_FILE")
	os.Setenv("HOMELYTICS_TLS_REQUIRE_CLIENT_CERT", "true")
	defer os.Unsetenv("HOMELYTICS_TLS_REQUIRE_CLIENT_CERT")
	os.Setenv("HOMELYTICS_TLS_CLIENT_CA_FILE", "/etc/ca.pem")
	defer os.Unsetenv("HOMELYTICS_TLS_CLIENT_CA_FILE")

	_, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

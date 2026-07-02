package security

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"backend/internal/config"
)

func TestPairingClaimIssuesCertificateAndConsumesToken(t *testing.T) {
	cfg := config.Config{Pairing: config.PairingConfig{Tokens: map[string]struct{}{"once": {}}, CertTTL: time.Hour}}
	ca, err := NewCertificateAuthority(cfg)
	if err != nil {
		t.Fatal(err)
	}
	service := NewPairingService(cfg, ca)
	resp, err := service.Claim(context.Background(), "once", PairingRequest{AgentName: "devbox", Hostname: "arch"})
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if resp.AgentID == "" || !strings.Contains(resp.Certificate, "BEGIN CERTIFICATE") || !strings.Contains(resp.PrivateKey, "BEGIN RSA PRIVATE KEY") {
		t.Fatalf("response = %#v", resp)
	}
	block, _ := pem.Decode([]byte(resp.Certificate))
	if block == nil {
		t.Fatal("certificate PEM did not decode")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != resp.AgentID {
		t.Fatalf("CommonName = %q, want %q", cert.Subject.CommonName, resp.AgentID)
	}
	if _, err := service.Claim(context.Background(), "once", PairingRequest{AgentName: "devbox"}); err == nil {
		t.Fatal("Claim() expected consumed token error")
	}
}

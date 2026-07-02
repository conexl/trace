package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"backend/internal/config"

	"go.uber.org/fx"
)

var Module = fx.Module("security", fx.Provide(NewCertificateAuthority), fx.Provide(NewPairingService))

type CertificateAuthority struct {
	certPEM []byte
	cert    *x509.Certificate
	key     *rsa.PrivateKey
	ttl     time.Duration
}

type IssuedCertificate struct {
	AgentID     string    `json:"agent_id"`
	Certificate string    `json:"certificate_pem"`
	PrivateKey  string    `json:"private_key_pem"`
	CACert      string    `json:"ca_cert_pem"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func NewCertificateAuthority(cfg config.Config) (*CertificateAuthority, error) {
	if cfg.Pairing.CACertFile != "" || cfg.Pairing.CAKeyFile != "" {
		return loadCertificateAuthority(cfg.Pairing.CACertFile, cfg.Pairing.CAKeyFile, cfg.Pairing.CertTTL)
	}
	certPEM, cert, key, err := generateEphemeralCA()
	if err != nil {
		return nil, err
	}
	return &CertificateAuthority{certPEM: certPEM, cert: cert, key: key, ttl: cfg.Pairing.CertTTL}, nil
}

func (ca *CertificateAuthority) IssueClientCertificate(agentID string, agentName string) (IssuedCertificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return IssuedCertificate{}, err
	}
	serial, err := randomSerial()
	if err != nil {
		return IssuedCertificate{}, err
	}
	now := time.Now().UTC()
	expires := now.Add(ca.ttl)
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   agentID,
			Organization: []string{"Homelytics Agents"},
		},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              expires,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{agentID},
	}
	if agentName != "" && agentName != agentID {
		template.DNSNames = append(template.DNSNames, agentName)
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		return IssuedCertificate{}, err
	}
	return IssuedCertificate{
		AgentID:     agentID,
		Certificate: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})),
		PrivateKey:  string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
		CACert:      string(ca.certPEM),
		ExpiresAt:   expires,
	}, nil
}

func loadCertificateAuthority(certFile string, keyFile string, ttl time.Duration) (*CertificateAuthority, error) {
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("pairing ca cert and key must be configured together")
	}
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("read pairing ca cert: %w", err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read pairing ca key: %w", err)
	}
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("parse pairing ca cert: invalid PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("parse pairing ca key: invalid PEM")
	}
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return &CertificateAuthority{certPEM: certPEM, cert: cert, key: key, ttl: ttl}, nil
}

func generateEphemeralCA() ([]byte, *x509.Certificate, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}
	serial, err := randomSerial()
	if err != nil {
		return nil, nil, nil, err
	}
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "Homelytics Development Pairing CA",
			Organization: []string{"Homelytics"},
		},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, err
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return certPEM, cert, key, nil
}

func randomSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, limit)
}

func randomAgentID(agentName string) (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	prefix := "agt"
	if agentName != "" {
		sum := sha256.Sum256([]byte(agentName))
		prefix = hex.EncodeToString(sum[:])[:8]
	}
	return prefix + "-" + hex.EncodeToString(raw), nil
}

func TouchForFx(context.Context, *CertificateAuthority) error { return nil }

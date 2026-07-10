package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend/internal/config"
)

type PairingService struct {
	mu     sync.Mutex
	tokens map[string]time.Time
	ca     *CertificateAuthority
}

type PairingRequest struct {
	AgentName string `json:"agent_name"`
	Hostname  string `json:"hostname"`
}

type PairingResponse struct {
	AgentID     string    `json:"agent_id"`
	Certificate string    `json:"certificate_pem"`
	PrivateKey  string    `json:"private_key_pem"`
	CACert      string    `json:"ca_cert_pem"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func NewPairingService(cfg config.Config, ca *CertificateAuthority) *PairingService {
	tokens := make(map[string]time.Time, len(cfg.Pairing.Tokens))
	for token := range cfg.Pairing.Tokens {
		tokens[token] = time.Time{}
	}
	return &PairingService{tokens: tokens, ca: ca}
}

func (s *PairingService) CreateToken(ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", time.Time{}, err
	}
	code := strings.ToUpper(hex.EncodeToString(b[:]))
	expiresAt := time.Now().UTC().Add(ttl)
	s.mu.Lock()
	s.tokens[code] = expiresAt
	s.mu.Unlock()
	return code, expiresAt, nil
}

func (s *PairingService) Claim(ctx context.Context, token string, req PairingRequest) (PairingResponse, error) {
	select {
	case <-ctx.Done():
		return PairingResponse{}, ctx.Err()
	default:
	}
	if token == "" {
		return PairingResponse{}, fmt.Errorf("pairing token is required")
	}
	if req.AgentName == "" && req.Hostname == "" {
		return PairingResponse{}, fmt.Errorf("agent_name or hostname is required")
	}
	if !s.consumeToken(token) {
		return PairingResponse{}, fmt.Errorf("invalid pairing token")
	}
	agentID, err := randomAgentID(req.AgentName + req.Hostname)
	if err != nil {
		return PairingResponse{}, err
	}
	issued, err := s.ca.IssueClientCertificate(agentID, req.AgentName)
	if err != nil {
		return PairingResponse{}, err
	}
	return PairingResponse(issued), nil
}

func (s *PairingService) consumeToken(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for candidate := range s.tokens {
		expiresAt := s.tokens[candidate]
		if !expiresAt.IsZero() && now.After(expiresAt) {
			delete(s.tokens, candidate)
			continue
		}
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(token)) == 1 {
			delete(s.tokens, candidate)
			return true
		}
	}
	return false
}

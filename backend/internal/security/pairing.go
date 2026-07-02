package security

import (
	"context"
	"crypto/subtle"
	"fmt"
	"sync"
	"time"

	"backend/internal/config"
)

type PairingService struct {
	mu     sync.Mutex
	tokens map[string]struct{}
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
	tokens := make(map[string]struct{}, len(cfg.Pairing.Tokens))
	for token := range cfg.Pairing.Tokens {
		tokens[token] = struct{}{}
	}
	return &PairingService{tokens: tokens, ca: ca}
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
	for candidate := range s.tokens {
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(token)) == 1 {
			delete(s.tokens, candidate)
			return true
		}
	}
	return false
}

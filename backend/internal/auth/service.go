package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/users"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrRegistrationClosed = errors.New("registration is disabled")
	ErrForbidden          = errors.New("forbidden")
)

type Session struct {
	Email     string
	Role      string
	Plan      string
	ExpiresAt time.Time
}

type Service struct {
	cfg           config.Config
	store         users.Store
	sessionStore  SessionStore
	mu            sync.RWMutex
	sessionsCache map[string]Session // token -> session
}

func NewService(cfg config.Config, store users.Store, sessionStore SessionStore) *Service {
	return &Service{cfg: cfg, store: store, sessionStore: sessionStore, sessionsCache: make(map[string]Session)}
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	if len(password) < 8 {
		return "", ErrWeakPassword
	}

	count, err := s.store.Count(ctx)
	if err != nil {
		return "", fmt.Errorf("count users: %w", err)
	}

	firstUser := count == 0
	if s.cfg.Auth.RegistrationDisabled && !firstUser {
		return "", ErrRegistrationClosed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleMember,
		Plan:         domain.PlanFree,
		Verified:     false,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.store.Create(ctx, user); err != nil {
		if errors.Is(err, users.ErrExists) {
			return "", ErrUserExists
		}
		return "", err
	}
	return s.createSession(user), nil
}

func (s *Service) RegisterAdmin(ctx context.Context, adminToken, email, password string) (string, error) {
	if !s.IsAdminToken(adminToken) {
		return "", ErrForbidden
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	if len(password) < 8 {
		return "", ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleMember,
		Plan:         domain.PlanFree,
		Verified:     false,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.store.Create(ctx, user); err != nil {
		if errors.Is(err, users.ErrExists) {
			return "", ErrUserExists
		}
		return "", err
	}
	return s.createSession(user), nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return s.createSession(user), nil
}

func (s *Service) ValidateToken(token string) (Session, bool) {
	s.mu.RLock()
	session, ok := s.sessionsCache[token]
	s.mu.RUnlock()
	if !ok {
		return Session{}, false
	}
	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessionsCache, token)
		s.mu.Unlock()
		return Session{}, false
	}
	return session, ok
}

func (s *Service) LoadSession(ctx context.Context, token string) (Session, bool, error) {
	session, ok, err := s.sessionStore.Get(ctx, token)
	if err != nil {
		return Session{}, false, err
	}
	if ok {
		if time.Now().After(session.ExpiresAt) {
			_ = s.sessionStore.Delete(ctx, token)
			return Session{}, false, nil
		}
		s.mu.Lock()
		s.sessionsCache[token] = session
		s.mu.Unlock()
	}
	return session, ok, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	s.mu.Lock()
	delete(s.sessionsCache, token)
	s.mu.Unlock()
	return s.sessionStore.Delete(ctx, token)
}

func (s *Service) User(ctx context.Context, email string) (domain.User, error) {
	user, err := s.store.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return domain.User{}, err
	}
	user.Plan = domain.NormalizePlan(user.Plan)
	return user, nil
}

func (s *Service) UpdatePlan(ctx context.Context, email string, plan string) (domain.User, error) {
	if domain.NormalizePlan(plan) != plan {
		return domain.User{}, fmt.Errorf("unsupported plan %q", plan)
	}
	user, err := s.store.UpdatePlan(ctx, strings.ToLower(strings.TrimSpace(email)), plan)
	if err != nil {
		return domain.User{}, err
	}
	user.Plan = domain.NormalizePlan(user.Plan)
	s.updateCachedPlan(user.Email, user.Plan)
	return user, nil
}

func (s *Service) IsAdminToken(token string) bool {
	return s.cfg.Auth.AdminToken != "" && token == s.cfg.Auth.AdminToken
}

func (s *Service) IsAdmin(session Session) bool {
	return s.IsProductMember(session)
}

func (s *Service) IsOwner(session Session) bool {
	return session.Role == domain.RoleOwner
}

func (s *Service) IsProductMember(session Session) bool {
	switch session.Role {
	case domain.RoleMember, domain.RoleOwner, domain.RoleAdmin, domain.RoleViewer:
		return session.Email != ""
	default:
		return false
	}
}

func (s *Service) createSession(user domain.User) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	token := hex.EncodeToString(b)
	session := Session{
		Email:     user.Email,
		Role:      user.Role,
		Plan:      domain.NormalizePlan(user.Plan),
		ExpiresAt: time.Now().Add(s.cfg.Auth.SessionTTL),
	}
	s.mu.Lock()
	s.sessionsCache[token] = session
	s.mu.Unlock()
	// Best-effort persistence; failures are logged by caller if needed.
	_ = s.sessionStore.Set(context.Background(), token, session)
	return token
}

func (s *Service) updateCachedPlan(email string, plan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, session := range s.sessionsCache {
		if session.Email == email {
			session.Plan = plan
			s.sessionsCache[token] = session
			_ = s.sessionStore.Set(context.Background(), token, session)
		}
	}
}

package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	appmongo "backend/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionStore interface {
	Get(ctx context.Context, token string) (Session, bool, error)
	Set(ctx context.Context, token string, session Session) error
	Delete(ctx context.Context, token string) error
}

type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]Session)}
}

func (s *MemorySessionStore) Get(ctx context.Context, token string) (Session, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[token]
	return session, ok, nil
}

func (s *MemorySessionStore) Set(ctx context.Context, token string, session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = session
	return nil
}

func (s *MemorySessionStore) Delete(ctx context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
	return nil
}

type MongoSessionStore struct {
	client *appmongo.Client
}

func NewMongoSessionStore(client *appmongo.Client) *MongoSessionStore {
	return &MongoSessionStore{client: client}
}

func (s *MongoSessionStore) collection() *drivermongo.Collection {
	return s.client.Collection("sessions")
}

func (s *MongoSessionStore) Get(ctx context.Context, token string) (Session, bool, error) {
	var doc sessionDoc
	err := s.collection().FindOne(ctx, bson.M{"_id": token}).Decode(&doc)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return Session{}, false, nil
		}
		return Session{}, false, fmt.Errorf("find session: %w", err)
	}
	return Session{Email: doc.Email, Role: doc.Role, Plan: doc.Plan, ExpiresAt: doc.ExpiresAt}, true, nil
}

func (s *MongoSessionStore) Set(ctx context.Context, token string, session Session) error {
	_, err := s.collection().ReplaceOne(ctx, bson.M{"_id": token}, sessionDoc{
		Token:     token,
		Email:     session.Email,
		Role:      session.Role,
		Plan:      session.Plan,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: session.ExpiresAt,
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (s *MongoSessionStore) Delete(ctx context.Context, token string) error {
	_, err := s.collection().DeleteOne(ctx, bson.M{"_id": token})
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *MongoSessionStore) EnsureIndexes(ctx context.Context) error {
	_, err := s.collection().Indexes().CreateOne(ctx, drivermongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return err
}

type sessionDoc struct {
	Token     string    `bson:"_id"`
	Email     string    `bson:"email"`
	Role      string    `bson:"role"`
	Plan      string    `bson:"plan"`
	CreatedAt time.Time `bson:"created_at"`
	ExpiresAt time.Time `bson:"expires_at"`
}

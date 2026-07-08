package telegram

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"backend/internal/config"
	appmongo "backend/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var Module = fx.Module("telegram", fx.Provide(NewStore))

var (
	ErrNotFound = errors.New("telegram link not found")
	ErrExpired  = errors.New("telegram link expired")
	ErrUsed     = errors.New("telegram link already used")
)

type Chat struct {
	ID        int64  `json:"id" bson:"id"`
	Type      string `json:"type,omitempty" bson:"type,omitempty"`
	Username  string `json:"username,omitempty" bson:"username,omitempty"`
	Title     string `json:"title,omitempty" bson:"title,omitempty"`
	FirstName string `json:"first_name,omitempty" bson:"first_name,omitempty"`
}

type Link struct {
	Token     string     `json:"token" bson:"_id"`
	UserEmail string     `json:"user_email" bson:"user_email"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time  `json:"expires_at" bson:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
	Chat      *Chat      `json:"chat,omitempty" bson:"chat,omitempty"`
}

type Recipient struct {
	UserEmail string    `json:"user_email" bson:"_id"`
	Chat      Chat      `json:"chat" bson:"chat"`
	LinkedAt  time.Time `json:"linked_at" bson:"linked_at"`
}

type Store interface {
	CreateLink(ctx context.Context, userEmail string, ttl time.Duration) (Link, error)
	ClaimLink(ctx context.Context, token string, chat Chat) (Link, error)
	GetRecipient(ctx context.Context, userEmail string) (Recipient, error)
	ListRecipients(ctx context.Context) ([]Recipient, error)
	DeleteRecipient(ctx context.Context, userEmail string) error
}

func NewStore(cfg config.Config, client *appmongo.Client) Store {
	if cfg.Mongo.URI != "" && client != nil {
		return NewMongoStore(client)
	}
	return NewMemoryStore()
}

func NewToken() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

type MemoryStore struct {
	mu         sync.RWMutex
	links      map[string]Link
	recipients map[string]Recipient
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		links:      make(map[string]Link),
		recipients: make(map[string]Recipient),
	}
}

func (s *MemoryStore) CreateLink(ctx context.Context, userEmail string, ttl time.Duration) (Link, error) {
	token, err := NewToken()
	if err != nil {
		return Link{}, err
	}
	now := time.Now().UTC()
	link := Link{Token: token, UserEmail: userEmail, CreatedAt: now, ExpiresAt: now.Add(ttl)}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[token] = link
	return link, nil
}

func (s *MemoryStore) ClaimLink(ctx context.Context, token string, chat Chat) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, ok := s.links[token]
	if !ok {
		return Link{}, ErrNotFound
	}
	if link.UsedAt != nil {
		return Link{}, ErrUsed
	}
	now := time.Now().UTC()
	if now.After(link.ExpiresAt) {
		return Link{}, ErrExpired
	}
	link.UsedAt = &now
	link.Chat = &chat
	s.links[token] = link
	s.recipients[link.UserEmail] = Recipient{UserEmail: link.UserEmail, Chat: chat, LinkedAt: now}
	return link, nil
}

func (s *MemoryStore) GetRecipient(ctx context.Context, userEmail string) (Recipient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recipient, ok := s.recipients[userEmail]
	if !ok {
		return Recipient{}, ErrNotFound
	}
	return recipient, nil
}

func (s *MemoryStore) ListRecipients(ctx context.Context) ([]Recipient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recipients := make([]Recipient, 0, len(s.recipients))
	for _, recipient := range s.recipients {
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}

func (s *MemoryStore) DeleteRecipient(ctx context.Context, userEmail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.recipients, userEmail)
	return nil
}

type MongoStore struct {
	client *appmongo.Client
}

func NewMongoStore(client *appmongo.Client) *MongoStore {
	return &MongoStore{client: client}
}

func (s *MongoStore) links() *drivermongo.Collection {
	return s.client.Collection("telegram_links")
}

func (s *MongoStore) recipients() *drivermongo.Collection {
	return s.client.Collection("telegram_recipients")
}

func (s *MongoStore) CreateLink(ctx context.Context, userEmail string, ttl time.Duration) (Link, error) {
	token, err := NewToken()
	if err != nil {
		return Link{}, err
	}
	now := time.Now().UTC()
	link := Link{Token: token, UserEmail: userEmail, CreatedAt: now, ExpiresAt: now.Add(ttl)}
	if _, err := s.links().InsertOne(ctx, link); err != nil {
		return Link{}, fmt.Errorf("insert telegram link: %w", err)
	}
	return link, nil
}

func (s *MongoStore) ClaimLink(ctx context.Context, token string, chat Chat) (Link, error) {
	var link Link
	err := s.links().FindOne(ctx, bson.M{"_id": token}).Decode(&link)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return Link{}, ErrNotFound
		}
		return Link{}, fmt.Errorf("find telegram link: %w", err)
	}
	if link.UsedAt != nil {
		return Link{}, ErrUsed
	}
	now := time.Now().UTC()
	if now.After(link.ExpiresAt) {
		return Link{}, ErrExpired
	}

	update := bson.M{"$set": bson.M{"used_at": now, "chat": chat}}
	result, err := s.links().UpdateOne(ctx, bson.M{"_id": token, "used_at": bson.M{"$exists": false}}, update)
	if err != nil {
		return Link{}, fmt.Errorf("claim telegram link: %w", err)
	}
	if result.MatchedCount == 0 {
		return Link{}, ErrUsed
	}

	link.UsedAt = &now
	link.Chat = &chat
	recipient := Recipient{UserEmail: link.UserEmail, Chat: chat, LinkedAt: now}
	if _, err := s.recipients().ReplaceOne(ctx, bson.M{"_id": link.UserEmail}, recipient, options.Replace().SetUpsert(true)); err != nil {
		return Link{}, fmt.Errorf("save telegram recipient: %w", err)
	}
	return link, nil
}

func (s *MongoStore) GetRecipient(ctx context.Context, userEmail string) (Recipient, error) {
	var recipient Recipient
	err := s.recipients().FindOne(ctx, bson.M{"_id": userEmail}).Decode(&recipient)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return Recipient{}, ErrNotFound
		}
		return Recipient{}, fmt.Errorf("find telegram recipient: %w", err)
	}
	return recipient, nil
}

func (s *MongoStore) ListRecipients(ctx context.Context) ([]Recipient, error) {
	cursor, err := s.recipients().Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list telegram recipients: %w", err)
	}
	defer cursor.Close(ctx)
	var recipients []Recipient
	if err := cursor.All(ctx, &recipients); err != nil {
		return nil, fmt.Errorf("decode telegram recipients: %w", err)
	}
	return recipients, nil
}

func (s *MongoStore) DeleteRecipient(ctx context.Context, userEmail string) error {
	if _, err := s.recipients().DeleteOne(ctx, bson.M{"_id": userEmail}); err != nil {
		return fmt.Errorf("delete telegram recipient: %w", err)
	}
	return nil
}

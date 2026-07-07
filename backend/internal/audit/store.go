package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/domain"
	appmongo "backend/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store interface {
	Log(ctx context.Context, entry domain.AuditLog) error
	Recent(ctx context.Context, limit int) ([]domain.AuditLog, error)
}

type MongoStore struct {
	client *appmongo.Client
}

type MemoryStore struct {
	mu   sync.RWMutex
	logs []domain.AuditLog
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{logs: make([]domain.AuditLog, 0)}
}

func (s *MemoryStore) Log(ctx context.Context, entry domain.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	s.logs = append(s.logs, entry)
	return nil
}

func (s *MemoryStore) Recent(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AuditLog, len(s.logs))
	copy(out, s.logs)
	// Reverse
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func NewMongoStore(client *appmongo.Client) *MongoStore {
	return &MongoStore{client: client}
}

func (s *MongoStore) collection() *mongo.Collection {
	return s.client.Collection("audit_logs")
}

func (s *MongoStore) Log(ctx context.Context, entry domain.AuditLog) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	_, err := s.collection().InsertOne(ctx, entry)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (s *MongoStore) Recent(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(int64(limit))
	cursor, err := s.collection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("find audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []domain.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("decode audit logs: %w", err)
	}
	return logs, nil
}

func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	_, err := s.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "timestamp", Value: -1}},
	})
	return err
}

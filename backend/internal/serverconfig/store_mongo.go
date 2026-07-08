package serverconfig

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/domain"
	appmongo "backend/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client *appmongo.Client
}

func NewMongoStore(client *appmongo.Client) *MongoStore {
	return &MongoStore{client: client}
}

func (s *MongoStore) collection() *drivermongo.Collection {
	return s.client.Collection("server_configs")
}

func (s *MongoStore) historyCollection() *drivermongo.Collection {
	return s.client.Collection("server_config_history")
}

type historyDocument struct {
	ID         string                    `bson:"_id,omitempty"`
	ServerID   string                    `bson:"server_id"`
	Revision   int64                     `bson:"revision"`
	Config     domain.AgentDesiredConfig `bson:"config"`
	ArchivedAt time.Time                 `bson:"archived_at"`
}

func (s *MongoStore) Get(ctx context.Context, serverID string) (domain.AgentDesiredConfig, error) {
	var cfg domain.AgentDesiredConfig
	err := s.collection().FindOne(ctx, bson.M{"_id": serverID}).Decode(&cfg)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.AgentDesiredConfig{}, ErrNotFound
		}
		return domain.AgentDesiredConfig{}, fmt.Errorf("find server config: %w", err)
	}
	return cfg, nil
}

func (s *MongoStore) Set(ctx context.Context, serverID string, cfg domain.AgentDesiredConfig) error {
	current, err := s.Get(ctx, serverID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if err == nil {
		doc := historyDocument{
			ID:         fmt.Sprintf("%s:%d", serverID, current.Revision),
			ServerID:   serverID,
			Revision:   current.Revision,
			Config:     current,
			ArchivedAt: time.Now().UTC(),
		}
		if _, err := s.historyCollection().ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true)); err != nil {
			return fmt.Errorf("save server config history: %w", err)
		}
	}
	cfg.Revision = current.Revision + 1
	_, err = s.collection().ReplaceOne(ctx, bson.M{"_id": serverID}, cfg, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save server config: %w", err)
	}
	return nil
}

func (s *MongoStore) Previous(ctx context.Context, serverID string) (domain.AgentDesiredConfig, error) {
	var doc historyDocument
	err := s.historyCollection().FindOne(
		ctx,
		bson.M{"server_id": serverID},
		options.FindOne().SetSort(bson.D{{Key: "revision", Value: -1}}),
	).Decode(&doc)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.AgentDesiredConfig{}, ErrNotFound
		}
		return domain.AgentDesiredConfig{}, fmt.Errorf("find previous server config: %w", err)
	}
	return doc.Config, nil
}

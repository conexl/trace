package serverconfig

import (
	"context"
	"errors"
	"fmt"

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
	cfg.Revision = current.Revision + 1
	_, err = s.collection().ReplaceOne(ctx, bson.M{"_id": serverID}, cfg, options.Replace().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("save server config: %w", err)
	}
	return nil
}

package alerts

import (
	"context"

	"backend/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

type MongoStore struct {
	client *mongo.Client
	alerts *mongo.Collection
}

func NewStore(lc fx.Lifecycle, cfg config.Config) (Store, error) {
	if cfg.Mongo.URI == "" {
		return NewMemoryStore(cfg), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI).SetServerSelectionTimeout(cfg.Mongo.ConnectTimeout))
	if err != nil {
		return nil, err
	}
	store := &MongoStore{client: client, alerts: client.Database(cfg.Mongo.Database).Collection("alerts")}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
			defer cancel()
			if err := client.Ping(pingCtx, nil); err != nil {
				return err
			}
			_, err := store.alerts.Indexes().CreateMany(pingCtx, []mongo.IndexModel{
				{Keys: bson.D{{Key: "created_at", Value: -1}}},
				{Keys: bson.D{{Key: "server_id", Value: 1}, {Key: "created_at", Value: -1}}},
			})
			return err
		},
		OnStop: func(ctx context.Context) error {
			return client.Disconnect(ctx)
		},
	})
	return store, nil
}

func (s *MongoStore) Save(ctx context.Context, alert Alert) error {
	_, err := s.alerts.InsertOne(ctx, alert)
	return err
}

func (s *MongoStore) Recent(ctx context.Context, limit int) ([]Alert, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	cursor, err := s.alerts.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var alerts []Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

package store

import (
	"context"
	"errors"
	"time"

	"backend/internal/config"
	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var Module = fx.Module("store", fx.Provide(NewStore))

type MongoStore struct {
	cfg     config.Config
	client  *mongo.Client
	servers *mongo.Collection
	metrics *mongo.Collection
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
	store := &MongoStore{cfg: cfg, client: client, servers: client.Database(cfg.Mongo.Database).Collection("servers"), metrics: client.Database(cfg.Mongo.Database).Collection("metrics")}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
			defer cancel()
			if err := client.Ping(pingCtx, nil); err != nil {
				return err
			}
			_, err := store.servers.Indexes().CreateOne(pingCtx, mongo.IndexModel{
				Keys:    bson.D{{Key: "summary.id", Value: 1}},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				return err
			}
			_, err = store.metrics.Indexes().CreateMany(pingCtx, []mongo.IndexModel{
				{Keys: bson.D{{Key: "server_id", Value: 1}, {Key: "timestamp", Value: 1}}},
				{Keys: bson.D{{Key: "timestamp", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(7 * 24 * 3600)},
			})
			return err
		},
		OnStop: func(ctx context.Context) error {
			return client.Disconnect(ctx)
		},
	})
	return store, nil
}

func (s *MongoStore) UpsertSnapshot(ctx context.Context, snapshot domain.AgentSnapshot) (domain.ServerState, error) {
	state := stateFromSnapshot(snapshot, s.cfg, time.Now())
	if existing, err := s.GetServer(ctx, state.Summary.ID, time.Now()); err == nil {
		state.Events = append(existing.Events, snapshot.Events...)
		if len(state.Events) > s.cfg.State.MaxEvents {
			state.Events = state.Events[len(state.Events)-s.cfg.State.MaxEvents:]
		}
	}
	_, err := s.servers.ReplaceOne(ctx, bson.M{"summary.id": state.Summary.ID}, state, options.Replace().SetUpsert(true))
	if err != nil {
		return domain.ServerState{}, err
	}
	return state, nil
}

func (s *MongoStore) ListServers(ctx context.Context, now time.Time) ([]domain.ServerSummary, error) {
	cursor, err := s.servers.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "summary.name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var states []domain.ServerState
	if err := cursor.All(ctx, &states); err != nil {
		return nil, err
	}
	summaries := make([]domain.ServerSummary, 0, len(states))
	for _, state := range states {
		summary := state.Summary
		summary.Status = statusFor(summary.LastSeen, now, s.cfg.State.OfflineAfter)
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (s *MongoStore) GetServer(ctx context.Context, id string, now time.Time) (domain.ServerState, error) {
	var state domain.ServerState
	err := s.servers.FindOne(ctx, bson.M{"summary.id": id}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ServerState{}, ErrNotFound{ID: id}
	}
	if err != nil {
		return domain.ServerState{}, err
	}
	state.Summary.Status = statusFor(state.Summary.LastSeen, now, s.cfg.State.OfflineAfter)
	return state, nil
}

func (s *MongoStore) SaveMetric(ctx context.Context, metric domain.Metric) error {
	_, err := s.metrics.InsertOne(ctx, metric)
	return err
}

func (s *MongoStore) GetMetrics(ctx context.Context, serverID string, from, to time.Time) ([]domain.Metric, error) {
	filter := bson.M{
		"server_id": serverID,
		"timestamp": bson.M{"$gte": from, "$lte": to},
	}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := s.metrics.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var metrics []domain.Metric
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *MongoStore) UpdateDesiredRevision(ctx context.Context, serverID string, revision int64) error {
	_, err := s.servers.UpdateOne(ctx, bson.M{"summary.id": serverID}, bson.M{"$set": bson.M{"summary.desired_config_revision": revision}})
	return err
}

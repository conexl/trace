package tasks

import (
	"context"
	"errors"
	"time"

	"backend/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var Module = fx.Module("tasks", fx.Provide(NewStore))

type MongoStore struct {
	client *mongo.Client
	tasks  *mongo.Collection
}

func NewStore(lc fx.Lifecycle, cfg config.Config) (Store, error) {
	if cfg.Mongo.URI == "" {
		return NewMemoryStore(), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI).SetServerSelectionTimeout(cfg.Mongo.ConnectTimeout))
	if err != nil {
		return nil, err
	}
	store := &MongoStore{client: client, tasks: client.Database(cfg.Mongo.Database).Collection("tasks")}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
			defer cancel()
			if err := client.Ping(pingCtx, nil); err != nil {
				return err
			}
			_, err := store.tasks.Indexes().CreateMany(pingCtx, []mongo.IndexModel{
				{Keys: bson.D{{Key: "server_id", Value: 1}, {Key: "status", Value: 1}, {Key: "created_at", Value: 1}}},
				{Keys: bson.D{{Key: "created_at", Value: -1}}},
			})
			return err
		},
		OnStop: func(ctx context.Context) error { return client.Disconnect(ctx) },
	})
	return store, nil
}

func (s *MongoStore) Enqueue(ctx context.Context, serverID string, taskName string, createdBy string) (Task, error) {
	return s.EnqueueWithPayload(ctx, serverID, taskName, TaskPayload{}, createdBy)
}

func (s *MongoStore) EnqueueWithPayload(ctx context.Context, serverID string, taskName string, payload TaskPayload, createdBy string) (Task, error) {
	id, err := newTaskID()
	if err != nil {
		return Task{}, err
	}
	task := Task{
		ID: id, ServerID: serverID, Name: taskName, Payload: payload, Status: StatusPending, CreatedAt: time.Now().UTC(),
		CreatedBy: createdBy,
		MaxRetries: 3,
		Timeout:    300,
	}
	_, err = s.tasks.InsertOne(ctx, task)
	return task, err
}

func (s *MongoStore) ClaimPending(ctx context.Context, serverID string, limit int) ([]Task, error) {
	if limit <= 0 {
		limit = 1
	}
	claimed := make([]Task, 0, limit)
	for len(claimed) < limit {
		now := time.Now().UTC()
		res := s.tasks.FindOneAndUpdate(
			ctx,
			bson.M{"server_id": serverID, "status": StatusPending},
			bson.M{"$set": bson.M{"status": StatusRunning, "claimed_at": now}},
			options.FindOneAndUpdate().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetReturnDocument(options.After),
		)
		var task Task
		if err := res.Decode(&task); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				break
			}
			return nil, err
		}
		claimed = append(claimed, task)
	}
	return claimed, nil
}

func (s *MongoStore) Complete(ctx context.Context, taskID string, result TaskResult) (Task, error) {
	status := StatusCompleted
	if result.ExitCode != 0 || result.Error != "" {
		status = StatusFailed
	}
	now := time.Now().UTC()
	res := s.tasks.FindOneAndUpdate(
		ctx,
		bson.M{"_id": taskID, "status": StatusRunning},
		bson.M{"$set": bson.M{"status": status, "completed_at": now, "result": result}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var task Task
	if err := res.Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			if _, getErr := s.Get(ctx, taskID); getErr == nil {
				return Task{}, ErrInvalidState{ID: taskID, Status: "not-running"}
			}
			return Task{}, ErrNotFound{ID: taskID}
		}
		return Task{}, err
	}
	return task, nil
}

func (s *MongoStore) Cancel(ctx context.Context, taskID string, reason string) (Task, error) {
	now := time.Now().UTC()
	res := s.tasks.FindOneAndUpdate(
		ctx,
		bson.M{"_id": taskID, "status": bson.M{"$in": []Status{StatusPending, StatusRunning}}},
		bson.M{"$set": bson.M{"status": StatusCanceled, "completed_at": now, "result.error": reason}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var task Task
	if err := res.Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Task{}, ErrNotFound{ID: taskID}
		}
		return Task{}, err
	}
	return task, nil
}

func (s *MongoStore) Get(ctx context.Context, taskID string) (Task, error) {
	var task Task
	err := s.tasks.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Task{}, ErrNotFound{ID: taskID}
	}
	return task, err
}

func (s *MongoStore) List(ctx context.Context, limit int) ([]Task, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := s.tasks.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var tasks []Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

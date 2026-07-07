package incidents

import (
	"context"
	"time"

	"backend/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

// Store interface for incidents
type Store interface {
	Save(ctx context.Context, incident Incident) error
	Get(ctx context.Context, id string) (*Incident, error)
	Recent(ctx context.Context, serverID string, limit int) ([]Incident, error)
	Range(ctx context.Context, serverID string, since time.Time) ([]Incident, error)
	GetOpen(ctx context.Context, serverID, serviceName string) (*Incident, error)
	AddTimelineEvent(ctx context.Context, incidentID string, event TimelineEvent) error
	UpdateStatus(ctx context.Context, incidentID, status string, resolvedAt *time.Time) error
}

// MemoryStore for development
type MemoryStore struct {
	incidents map[string]*Incident
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{incidents: make(map[string]*Incident)}
}

func (s *MemoryStore) Save(ctx context.Context, incident Incident) error {
	s.incidents[incident.ID] = &incident
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, id string) (*Incident, error) {
	incident, ok := s.incidents[id]
	if !ok {
		return nil, ErrNotFound
	}
	return incident, nil
}

func (s *MemoryStore) Recent(ctx context.Context, serverID string, limit int) ([]Incident, error) {
	var result []Incident
	for _, incident := range s.incidents {
		if serverID == "" || incident.ServerID == serverID {
			result = append(result, *incident)
		}
	}
	// Sort by created_at desc
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *MemoryStore) Range(ctx context.Context, serverID string, since time.Time) ([]Incident, error) {
	var result []Incident
	for _, incident := range s.incidents {
		if serverID != "" && incident.ServerID != serverID {
			continue
		}
		if incident.CreatedAt.Before(since) {
			continue
		}
		result = append(result, *incident)
	}
	// Sort by created_at desc to match Recent.
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result, nil
}

func (s *MemoryStore) GetOpen(ctx context.Context, serverID, serviceName string) (*Incident, error) {
	for _, incident := range s.incidents {
		if incident.ServerID == serverID && incident.ServiceName == serviceName && incident.Status == "open" {
			return incident, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) AddTimelineEvent(ctx context.Context, incidentID string, event TimelineEvent) error {
	incident, ok := s.incidents[incidentID]
	if !ok {
		return ErrNotFound
	}
	incident.Timeline = append(incident.Timeline, event)
	incident.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *MemoryStore) UpdateStatus(ctx context.Context, incidentID, status string, resolvedAt *time.Time) error {
	incident, ok := s.incidents[incidentID]
	if !ok {
		return ErrNotFound
	}
	incident.Status = status
	incident.UpdatedAt = time.Now().UTC()
	incident.ResolvedAt = resolvedAt
	return nil
}

// MongoStore for production
type MongoStore struct {
	client    *mongo.Client
	incidents *mongo.Collection
}

func NewMongoStore(lc fx.Lifecycle, cfg config.Config) (Store, error) {
	if cfg.Mongo.URI == "" {
		return NewMemoryStore(), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI).SetServerSelectionTimeout(cfg.Mongo.ConnectTimeout))
	if err != nil {
		return nil, err
	}
	store := &MongoStore{client: client, incidents: client.Database(cfg.Mongo.Database).Collection("incidents")}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
			defer cancel()
			if err := client.Ping(pingCtx, nil); err != nil {
				return err
			}
			_, err := store.incidents.Indexes().CreateMany(pingCtx, []mongo.IndexModel{
				{Keys: bson.D{{Key: "created_at", Value: -1}}},
				{Keys: bson.D{{Key: "server_id", Value: 1}, {Key: "created_at", Value: -1}}},
				{Keys: bson.D{{Key: "status", Value: 1}}},
			})
			return err
		},
		OnStop: func(ctx context.Context) error {
			return client.Disconnect(ctx)
		},
	})
	return store, nil
}

func (s *MongoStore) Save(ctx context.Context, incident Incident) error {
	_, err := s.incidents.InsertOne(ctx, incident)
	return err
}

func (s *MongoStore) Get(ctx context.Context, id string) (*Incident, error) {
	var incident Incident
	err := s.incidents.FindOne(ctx, bson.M{"_id": id}).Decode(&incident)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &incident, nil
}

func (s *MongoStore) Recent(ctx context.Context, serverID string, limit int) ([]Incident, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	filter := bson.M{}
	if serverID != "" {
		filter["server_id"] = serverID
	}
	cursor, err := s.incidents.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var incidents []Incident
	if err := cursor.All(ctx, &incidents); err != nil {
		return nil, err
	}
	return incidents, nil
}

func (s *MongoStore) Range(ctx context.Context, serverID string, since time.Time) ([]Incident, error) {
	filter := bson.M{"created_at": bson.M{"$gte": since}}
	if serverID != "" {
		filter["server_id"] = serverID
	}
	cursor, err := s.incidents.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var incidents []Incident
	if err := cursor.All(ctx, &incidents); err != nil {
		return nil, err
	}
	return incidents, nil
}

func (s *MongoStore) GetOpen(ctx context.Context, serverID, serviceName string) (*Incident, error) {
	var incident Incident
	err := s.incidents.FindOne(ctx, bson.M{
		"server_id":    serverID,
		"service_name": serviceName,
		"status":       "open",
	}).Decode(&incident)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &incident, nil
}

func (s *MongoStore) AddTimelineEvent(ctx context.Context, incidentID string, event TimelineEvent) error {
	_, err := s.incidents.UpdateByID(ctx, incidentID, bson.M{
		"$push": bson.M{"timeline": event},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	})
	return err
}

func (s *MongoStore) UpdateStatus(ctx context.Context, incidentID, status string, resolvedAt *time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		},
	}
	if resolvedAt != nil {
		update["$set"].(bson.M)["resolved_at"] = resolvedAt
	}
	_, err := s.incidents.UpdateByID(ctx, incidentID, update)
	return err
}

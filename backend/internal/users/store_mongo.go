package users

import (
	"context"
	"fmt"

	"backend/internal/domain"
	appmongo "backend/internal/mongo"

	"go.mongodb.org/mongo-driver/bson"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
)

type MongoStore struct {
	client *appmongo.Client
}

func NewMongoStore(client *appmongo.Client) *MongoStore {
	return &MongoStore{client: client}
}

func (s *MongoStore) collection() *drivermongo.Collection {
	return s.client.Collection("users")
}

func (s *MongoStore) Create(ctx context.Context, user domain.User) error {
	_, err := s.collection().InsertOne(ctx, user)
	if err != nil {
		if drivermongo.IsDuplicateKeyError(err) {
			return ErrExists
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (s *MongoStore) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := s.collection().FindOne(ctx, bson.M{"_id": email}).Decode(&user)
	if err != nil {
		if err == drivermongo.ErrNoDocuments {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

func (s *MongoStore) Count(ctx context.Context) (int, error) {
	count, err := s.collection().CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return int(count), nil
}

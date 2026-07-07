package mongo

import (
	"context"
	"fmt"
	"time"

	"backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

func New(cfg config.Config) (*Client, error) {
	if cfg.Mongo.URI == "" {
		return nil, fmt.Errorf("mongo URI is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	return &Client{client: client, database: client.Database(cfg.Mongo.Database)}, nil
}

func (c *Client) Database() *mongo.Database {
	return c.database
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

func Now() time.Time {
	return time.Now().UTC()
}

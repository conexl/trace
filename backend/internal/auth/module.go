package auth

import (
	"backend/internal/config"
	appmongo "backend/internal/mongo"
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("auth",
	fx.Provide(NewSessionStore),
	fx.Provide(NewService),
	fx.Invoke(InitSessionStore),
)

func NewSessionStore(cfg config.Config, client *appmongo.Client) SessionStore {
	if cfg.Mongo.URI != "" && client != nil {
		return NewMongoSessionStore(client)
	}
	return NewMemorySessionStore()
}

func InitSessionStore(lc fx.Lifecycle, store SessionStore, logger *zap.Logger) {
	if mongoStore, ok := store.(*MongoSessionStore); ok {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info("initializing mongo session store indexes")
				return mongoStore.EnsureIndexes(ctx)
			},
		})
	}
}

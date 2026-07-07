package audit

import (
	"backend/internal/config"
	appmongo "backend/internal/mongo"
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("audit",
	fx.Provide(NewStore),
	fx.Invoke(InitStore),
)

func NewStore(cfg config.Config, client *appmongo.Client) Store {
	if cfg.Mongo.URI != "" && client != nil {
		return NewMongoStore(client)
	}
	return NewMemoryStore()
}

func InitStore(lc fx.Lifecycle, store Store, logger *zap.Logger) {
	if mongoStore, ok := store.(*MongoStore); ok {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info("initializing audit store indexes")
				return mongoStore.EnsureIndexes(ctx)
			},
		})
	}
}

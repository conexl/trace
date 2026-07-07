package redis

import (
	"context"
	"backend/internal/config"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("redis", fx.Provide(NewClientOptional))

func NewClientOptional(lc fx.Lifecycle, cfg config.Config) (*redis.Client, error) {
	if cfg.Redis.Addr == "" {
		return nil, nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.ConnectTimeout,
		ReadTimeout:  cfg.Redis.ConnectTimeout,
		WriteTimeout: cfg.Redis.ConnectTimeout,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return client.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}

package presence

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/internal/config"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

type RedisStore struct {
	cfg    config.Config
	client *redis.Client
}

func NewStore(lc fx.Lifecycle, cfg config.Config) (Store, error) {
	if cfg.Redis.Addr == "" {
		return NewMemoryStore(), nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.ConnectTimeout,
		ReadTimeout:  cfg.Redis.ConnectTimeout,
		WriteTimeout: cfg.Redis.ConnectTimeout,
	})
	store := &RedisStore{cfg: cfg, client: client}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.Redis.ConnectTimeout)
			defer cancel()
			return client.Ping(pingCtx).Err()
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})
	return store, nil
}

func (s *RedisStore) Touch(ctx context.Context, serverID string, seenAt time.Time) error {
	return s.client.Set(ctx, s.key(serverID), seenAt.UTC().UnixNano(), s.cfg.State.OfflineAfter).Err()
}

func (s *RedisStore) LastSeen(ctx context.Context, serverID string) (time.Time, bool, error) {
	raw, err := s.client.Get(ctx, s.key(serverID)).Result()
	if errors.Is(err, redis.Nil) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	nanos, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("decode presence timestamp: %w", err)
	}
	return time.Unix(0, nanos).UTC(), true, nil
}

func (s *RedisStore) key(serverID string) string {
	return s.cfg.Redis.KeyPrefix + ":presence:" + serverID
}

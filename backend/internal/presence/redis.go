package presence

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/internal/config"

	redis "github.com/redis/go-redis/v9"
)

type RedisStore struct {
	cfg    config.Config
	client *redis.Client
}

func NewStore(cfg config.Config, client *redis.Client) Store {
	if client != nil {
		return &RedisStore{cfg: cfg, client: client}
	}
	return NewMemoryStore()
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

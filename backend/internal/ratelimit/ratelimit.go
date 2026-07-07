package ratelimit

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Limiter interface {
	Allow(ctx context.Context, key string) bool
}

type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int
	window  time.Duration
}

type bucket struct {
	tokens  int
	resetAt time.Time
}

func NewMemory(rate int, window time.Duration) *MemoryLimiter {
	return &MemoryLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

func (l *MemoryLimiter) Allow(ctx context.Context, key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{tokens: l.rate - 1, resetAt: now.Add(l.window)}
		return true
	}
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func (l *MemoryLimiter) Cleanup(every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		now := time.Now().UTC()
		for key, b := range l.buckets {
			if now.After(b.resetAt) {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}

type RedisLimiter struct {
	client *redis.Client
	rate   int
	window time.Duration
	prefix string
}

func NewRedis(client *redis.Client, rate int, window time.Duration, prefix string) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		rate:   rate,
		window: window,
		prefix: prefix,
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) bool {
	fullKey := fmt.Sprintf("ratelimit:%s:%s", l.prefix, key)
	count, err := l.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return true // Fail open
	}
	if count == 1 {
		l.client.Expire(ctx, fullKey, l.window)
	}
	return count <= int64(l.rate)
}

func ClientIP(r *http.Request, trustForwarded bool) string {
	if trustForwarded {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
		if xri := r.Header.Get("X-Real-Ip"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func Middleware(l Limiter, trustForwarded bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(r.Context(), ClientIP(r, trustForwarded)) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}
		next(w, r)
	}
}

func Allow(ctx context.Context, l Limiter, key string) bool {
	return l.Allow(ctx, key)
}

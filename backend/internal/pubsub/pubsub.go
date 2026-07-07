package pubsub

import (
	"context"
	"sync"

	redis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

var Module = fx.Module("pubsub", fx.Provide(New))

type Service struct {
	client *redis.Client
	mu     sync.RWMutex
	subs   map[chan []byte]struct{}
}

func New(client *redis.Client) *Service {
	s := &Service{
		client: client,
		subs:   make(map[chan []byte]struct{}),
	}
	if client != nil {
		go s.listenRedis()
	}
	return s
}

func (s *Service) Publish(ctx context.Context, channel string, data []byte) error {
	if s.client != nil {
		return s.client.Publish(ctx, channel, data).Err()
	}
	s.broadcast(data)
	return nil
}

func (s *Service) Subscribe() chan []byte {
	ch := make(chan []byte, 100)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *Service) Unsubscribe(ch chan []byte) {
	s.mu.Lock()
	delete(s.subs, ch)
	s.mu.Unlock()
	close(ch)
}

func (s *Service) broadcast(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subs {
		select {
		case ch <- data:
		default:
		}
	}
}

func (s *Service) listenRedis() {
	ctx := context.Background()
	pubsub := s.client.Subscribe(ctx, "events")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		s.broadcast([]byte(msg.Payload))
	}
}

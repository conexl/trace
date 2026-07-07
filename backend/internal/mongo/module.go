package mongo

import (
	"backend/internal/config"

	"go.uber.org/fx"
)

var Module = fx.Module("mongo", fx.Provide(NewClientOptional))

func NewClientOptional(cfg config.Config) (*Client, error) {
	if cfg.Mongo.URI == "" {
		return nil, nil
	}
	return New(cfg)
}

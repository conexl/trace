package serverconfig

import (
	"backend/internal/config"
	appmongo "backend/internal/mongo"

	"go.uber.org/fx"
)

var Module = fx.Module("serverconfig", fx.Provide(NewStore))

func NewStore(cfg config.Config, client *appmongo.Client) Store {
	if cfg.Mongo.URI != "" && client != nil {
		return NewMongoStore(client)
	}
	return NewMemoryStore()
}

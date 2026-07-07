package notifications

import (
	"backend/internal/config"
	"backend/internal/logging"
	"backend/internal/redis"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		config.Module,
		logging.Module,
		redis.Module,
		Module,
	)
}

package notifications

import (
	"backend/internal/config"
	"backend/internal/logging"
	"backend/internal/mongo"
	"backend/internal/redis"
	"backend/internal/telegram"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		config.Module,
		logging.Module,
		mongo.Module,
		redis.Module,
		telegram.Module,
		Module,
	)
}

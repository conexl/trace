package logging

import (
	"backend/internal/config"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var Module = fx.Module("logging", fx.Provide(NewLogger), fx.WithLogger(NewFxLogger))

func NewLogger(cfg config.Config) (*zap.Logger, error) {
	if cfg.Environment == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func NewFxLogger(logger *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: logger.Named("fx")}
}

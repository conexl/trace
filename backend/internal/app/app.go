package app

import (
	"backend/internal/ai"
	"backend/internal/alerts"
	"backend/internal/audit"
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/httpapi"
	"backend/internal/incidents"
	"backend/internal/ingest"
	"backend/internal/logging"
	"backend/internal/mongo"
	"backend/internal/presence"
	"backend/internal/pubsub"
	"backend/internal/redis"
	"backend/internal/security"
	"backend/internal/serverconfig"
	"backend/internal/store"
	"backend/internal/tasks"
	"backend/internal/users"

	"go.uber.org/fx"
)

func New() *fx.App {
	return fx.New(
		config.Module,
		logging.Module,
		mongo.Module,
		users.Module,
		auth.Module,
		security.Module,
		store.Module,
		serverconfig.Module,
		tasks.Module,
		alerts.Module,
		incidents.Module,
		ai.Module,
		fx.Provide(ai.NewAnalyzer),
		audit.Module,
		presence.Module,
		pubsub.Module,
		redis.Module,
		ingest.Module,
		httpapi.Module,
	)
}

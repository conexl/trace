package app

import (
	"backend/internal/alerts"
	"backend/internal/config"
	"backend/internal/httpapi"
	"backend/internal/ingest"
	"backend/internal/logging"
	"backend/internal/security"
	"backend/internal/store"
	"backend/internal/tasks"

	"go.uber.org/fx"
)

func New() *fx.App {
	return fx.New(
		config.Module,
		logging.Module,
		security.Module,
		store.Module,
		tasks.Module,
		alerts.Module,
		ingest.Module,
		httpapi.Module,
	)
}

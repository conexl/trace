package app

import (
	"backend/internal/config"
	"backend/internal/httpapi"
	"backend/internal/ingest"
	"backend/internal/logging"
	"backend/internal/security"
	"backend/internal/store"

	"go.uber.org/fx"
)

func New() *fx.App {
	return fx.New(
		config.Module,
		logging.Module,
		security.Module,
		store.Module,
		ingest.Module,
		httpapi.Module,
	)
}

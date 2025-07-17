package api

import (
	"context"
	"effective-mobile/internal/config"
	"effective-mobile/internal/http/middleware"
	"log/slog"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type server struct {
	e      *echo.Echo
	config *config.CRUDConfig
	log    *slog.Logger
}

func NewHTTPServer(log *slog.Logger, deps HandlersDependencies, config *config.CRUDConfig) *server {
	e := echo.New()
	e.Use(echomiddleware.RequestID())
	e.Use(middleware.NewEnrichRequestContextMiddleware())
	e.Use(middleware.NewRequestLoggerMiddleware(log))
	RegisterHandlers(e, NewStrictHandler(
		deps,
		[]StrictMiddlewareFunc{},
	))

	e.File("/", config.Swagger.UIPath)
	e.File("/swagger.yml", config.Swagger.SpecPath)

	return &server{
		e:      e,
		config: config,
		log:    log,
	}
}

func (s *server) Start() error {
	cfg := s.config.HTTPServerConfig

	s.log.Info("starting server", slog.String("address", cfg.Address))
	return s.e.Start(cfg.Address)
}

func (s *server) Stop(ctx context.Context) error {
	s.log.Info("stopping server")
	return s.e.Shutdown(ctx)
}

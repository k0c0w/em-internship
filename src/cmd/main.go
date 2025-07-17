package main

import (
	"context"
	"effective-mobile/internal/config"
	"effective-mobile/internal/http/api"
	"effective-mobile/internal/service"
	storage "effective-mobile/internal/storage/postgresql"
	"effective-mobile/internal/storage/postgresql/migrations"
	"effective-mobile/pkg/logger/sl"
	"effective-mobile/pkg/storage/postgresql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoadCRUDConfig()

	log := setupLogger()
	log.Info("starting app")

	log.Info("Initializing storage")
	pgClient, pCfg := mustInitStorage(log, cfg)
	defer func() {
		log.Info("Closing connection")
		pgClient.Close()
	}()

	if cfg.ShouldMigrate {
		log.Info("Perfoming migrations")
		err := migrations.RunMigrations(context.Background(), pCfg.ConnectionString(), log)

		if err != nil {
			log.Error("error during migrations", sl.Err(err))
			return
		}
	}

	subStorage := storage.NewSubscriptionStorage(pgClient, log)

	log.Info("Initializing service")
	service := service.NewSubscriptionService(subStorage, log)

	log.Info("Setting up http server")
	deps := api.HandlersDependencies{
		Log:                 log,
		SubscriptionService: service,
	}
	srv := api.NewHTTPServer(log, deps, cfg)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("failed to start server", sl.Err(err))
			done <- syscall.SIGINT
		}
	}()

	<-done
	log.Info("Stopping http server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

	} else {
		log.Info("server stopped")
	}
}

func setupLogger() *slog.Logger {
	var log *slog.Logger

	switch true {
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func mustInitStorage(log *slog.Logger, cfg *config.CRUDConfig) (postgresql.Client, postgresql.PostgresConfig) {
	sCfg := cfg.StorageConfig
	pCfg := postgresql.PostgresConfig{
		Host:            sCfg.Host,
		Port:            sCfg.Port,
		User:            sCfg.User,
		Password:        sCfg.Password,
		DB:              sCfg.DB,
		ConnectAttempts: 5,
		ConnectTimeout:  10,
	}

	client, err := postgresql.NewClient(pCfg)
	if err != nil {
		log.Error("failed to connect to postgresql", sl.Err(err))
		os.Exit(1)
		return nil, pCfg
	}

	return client, pCfg
}

package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(ctx context.Context, connectionString string, log *slog.Logger) error {
	const op = "storage.postgresql.migrations.RunMigrations"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Error("failed to connect to database", slog.String("op", op), slog.Any("error", err))
		return fmt.Errorf("%s: unable to connect to database: %w", op, err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("failed to initialize database driver", slog.String("op", op), slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}

	// Use file-based migrations from the migrations directory
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/storage/postgresql/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Error("failed to initialize migration instance", slog.String("op", op), slog.Any("error", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer m.Close()

	// Apply migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("failed to apply migrations", slog.String("op", op), slog.Any("error", err))
		return fmt.Errorf("%s: migration up failed: %w", op, err)
	}

	log.Info("migrations applied successfully", slog.String("op", op))
	return nil
}

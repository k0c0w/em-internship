package postgresql

import (
	"context"
	"effective-mobile/internal/models"
	"effective-mobile/internal/storage"
	"effective-mobile/pkg/logger/sl"
	pgsql "effective-mobile/pkg/storage/postgresql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func NewSubscriptionStorage(c pgsql.Client, log *slog.Logger) storage.SubscriptionsStorage {
	log = log.With(slog.String("component", "SubscriptionsStorage"))
	return &subscriptionsStorage{
		client: c,
		log:    log,
	}
}

type subscriptionsStorage struct {
	client pgsql.Client
	log    *slog.Logger
}

func (s *subscriptionsStorage) logSqlQuery(sql string) {
	pretty := strings.ReplaceAll(sql, "\t", "")
	s.log.Info("performing query", slog.String("sql", pretty))
}

func (s *subscriptionsStorage) Add(ctx context.Context, sub models.Subscription) error {
	const op = "storage.postgresql.subscriptions.Add"
	const sql = `
		INSERT INTO subscriptions (id, owner_id, service_name, price, is_deleted, start_time, end_time)
			 VALUES ($1, $2, $3, $4, 0::BIT, $5, $6);`

	tx, err := s.client.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err), slog.String("op", op))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				s.log.Error("failed to rollback transaction", sl.Err(rollbackErr), slog.String("op", op))
			}
		}
	}()

	args := []any{sub.ID, sub.Owner, sub.ServiceName, sub.PriceRUB, sub.StartedAt.UTC()}
	if sub.IsCompleted() {
		args = append(args, sub.CompletedAt.UTC())
	} else {
		args = append(args, nil)
	}

	s.logSqlQuery(sql)
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.log.Error("database error during insert", sl.Err(pgErr), slog.String("op", op), slog.Any("subscription_id", sub.ID))
			return fmt.Errorf("%s: database error: %w", op, pgErr)
		}
		s.log.Error("failed to execute insert", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", sub.ID))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", sub.ID))
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("successfully added subscription", slog.Any("subscription_id", sub.ID), slog.String("op", op))
	return nil
}

func (s *subscriptionsStorage) RemoveByID(ctx context.Context, id models.SubscriptionID) error {
	const op = "storage.postgresql.subscriptions.RemoveByID"
	const sql = `
		UPDATE subscriptions
		   SET is_deleted = 1::BIT
		 WHERE id = $1 AND is_deleted = 0::BIT;`

	tx, err := s.client.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", id))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				s.log.Error("failed to rollback transaction", sl.Err(rollbackErr), slog.String("op", op))
			}
		}
	}()

	s.logSqlQuery(sql)
	_, err = tx.Exec(ctx, sql, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.log.Error("database error during update", sl.Err(pgErr), slog.String("op", op), slog.Any("subscription_id", id))
			return fmt.Errorf("%s: database error: %w", op, pgErr)
		}
		s.log.Error("failed to execute update", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", id))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", id))
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("successfully removed subscription", slog.Any("subscription_id", id), slog.String("op", op))
	return nil
}

func (s *subscriptionsStorage) Update(ctx context.Context, sub models.Subscription) error {
	const op = "storage.postgresql.subscriptions.Update"
	const sql = `
		UPDATE subscriptions
		   SET 	 
		   		  owner_id = $1
		   		, service_name = $2
				, price = $3
				, start_time = $4
				, end_time = $5
		 WHERE id = $6 AND is_deleted = 0::BIT;`

	tx, err := s.client.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", sub.ID))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				s.log.Error("failed to rollback transaction", sl.Err(rollbackErr), slog.String("op", op))
			}
		}
	}()

	args := []any{sub.Owner, sub.ServiceName, sub.PriceRUB, sub.StartedAt.UTC()}
	if sub.IsCompleted() {
		args = append(args, sub.CompletedAt.UTC())
	} else {
		args = append(args, nil)
	}
	args = append(args, sub.ID)

	s.logSqlQuery(sql)
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.log.Error("database error during update", sl.Err(pgErr), slog.String("op", op), slog.Any("subscription_id", sub.ID))
			return fmt.Errorf("%s: database error: %w", op, pgErr)
		}
		s.log.Error("failed to execute update", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", sub.ID))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		s.log.Error("Failed to commit transaction", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", sub.ID))
		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("successfully updated subscription", slog.Any("subscription_id", sub.ID), slog.String("op", op))
	return nil
}

func (s *subscriptionsStorage) FindByID(ctx context.Context, id models.SubscriptionID) (*models.Subscription, error) {
	const op = "storage.postgresql.subscriptions.FindByID"
	const sql = `
		SELECT 
				  id
				, owner_id
				, service_name
				, price
				, start_time
				, end_time
		  FROM subscriptions
		 WHERE id = $1 AND is_deleted = 0::BIT;`

	var sub models.Subscription

	s.logSqlQuery(sql)
	err := s.client.QueryRow(ctx, sql, id).Scan(&sub.ID, &sub.Owner, &sub.ServiceName, &sub.PriceRUB, &sub.StartedAt, &sub.CompletedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.log.Error("database error during fetch", sl.Err(pgErr), slog.String("op", op), slog.Any("subscription_id", id))
			return nil, fmt.Errorf("%s: database error: %w", op, pgErr)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn("subscription not found", slog.String("op", op), slog.Any("subscription_id", id))
			return nil, storage.ErrSubscriptionNotFound
		}
		s.log.Error("failed to fetch subscription", sl.Err(err), slog.String("op", op), slog.Any("subscription_id", id))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("successfully fetched subscription", slog.Any("subscription_id", id), slog.String("op", op))
	return &sub, nil
}

func (s *subscriptionsStorage) Find(ctx context.Context, f storage.SubscriptionsFilter) ([]*models.Subscription, error) {
	const op = "storage.postgresql.subscriptions.Find"
	const sqlBase = `
		SELECT 
				  id
				, owner_id
				, service_name
				, price
				, start_time
				, end_time
		FROM subscriptions
		WHERE is_deleted = 0::BIT`

	sqlB := strings.Builder{}
	sqlB.WriteString(sqlBase)
	args := make([]interface{}, 0, 4)

	if f.OwnerID != uuid.Nil {
		sqlB.WriteString(fmt.Sprintf(" AND owner_id = $%d", len(args)+1))
		args = append(args, f.OwnerID)
	}

	if f.ServiceName != "" {
		sqlB.WriteString(fmt.Sprintf(" AND service_name = $%d", len(args)+1))
		args = append(args, f.ServiceName)
	}

	if f.StartTime != nil {
		sqlB.WriteString(fmt.Sprintf(" AND start_time >= $%d", len(args)+1))
		args = append(args, f.StartTime.UTC())
	}

	if f.EndTime != nil {
		sqlB.WriteString(fmt.Sprintf(" AND start_time <= $%d", len(args)+1))
		args = append(args, f.EndTime.UTC())
	}

	sqlB.WriteString((" ORDER BY id;"))
	sql := sqlB.String()
	s.logSqlQuery(sql)
	rows, err := s.client.Query(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.log.Error("database error during query", sl.Err(pgErr), slog.String("op", op), slog.Any("owner_id", f.OwnerID))
			return nil, fmt.Errorf("%s: database error: %w", op, pgErr)
		}
		s.log.Error("failed to execute query", sl.Err(err), slog.String("op", op), slog.Any("owner_id", f.OwnerID))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err = rows.Scan(&sub.ID, &sub.Owner, &sub.ServiceName, &sub.PriceRUB, &sub.StartedAt, &sub.CompletedAt)
		if err != nil {
			s.log.Warn("failed to scan row, continuing", sl.Err(err), slog.String("op", op))
			continue
		}
		subs = append(subs, &sub)
	}

	if err = rows.Err(); err != nil {
		s.log.Error("error iterating rows", sl.Err(err), slog.String("op", op))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("successfully fetched subscriptions", slog.String("op", op), slog.Any("owner_id", f.OwnerID), slog.Int("count", len(subs)))
	return subs, nil
}

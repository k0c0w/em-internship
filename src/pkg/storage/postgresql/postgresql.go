package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DB              string
	Secure          bool
	ConnectAttempts int
	ConnectTimeout  time.Duration
}

func (s PostgresConfig) ConnectionString() string {
	ssl := ""
	if !s.Secure {
		ssl = "?sslmode=disable"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s%s",
		s.User,
		s.Password,
		s.Host,
		s.Port,
		s.DB,
		ssl,
	)
}

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Close()
}

func NewClient(cfg PostgresConfig) (pool *pgxpool.Pool, err error) {
	connectProtocol := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	err = doWithTries(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err = pgxpool.Connect(ctx, connectProtocol)
		if err != nil {
			return err
		}

		return nil
	}, cfg.ConnectAttempts, 5*time.Second)

	return pool, err
}

func doWithTries(fn func() error, attemtps int, delayFragment time.Duration) (err error) {
	delay := delayFragment
	var delayIncreaseFactor int64 = 1
	for attemtps > 0 {

		if err = fn(); err != nil {
			time.Sleep(delay)
			attemtps--
			delayIncreaseFactor++
			delay = time.Duration(delay.Nanoseconds() * delayIncreaseFactor)
			continue
		}

		return nil
	}

	return
}

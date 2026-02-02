package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBAdapter satisfies the sqlc.DBTX interface.
type DBAdapter struct {
	pool *pgxpool.Pool
}

// Constructor
func NewDBAdapter(pool *pgxpool.Pool) *DBAdapter {
	return &DBAdapter{pool: pool}
}

// QueryRow wrapper
func (d *DBAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.pool.QueryRow(ctx, sql, args...)
}

// Exec wrapper
func (d *DBAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return d.pool.Exec(ctx, sql, args...)
}

// Query wrapper
func (d *DBAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return d.pool.Query(ctx, sql, args...)
}

// NewPool creates a connection pool and ensures the DB is reachable.
func NewPool(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("connecting to postgres...")

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("database connection established successfully")
	return pool, nil
}

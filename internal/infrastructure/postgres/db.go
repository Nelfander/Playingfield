package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

// Helper to create a new pool
func NewPool(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

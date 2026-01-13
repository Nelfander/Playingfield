package postgres

import (
	"context"
	"database/sql"

	"github.com/nelfander/Playingfield/internal/domain/user"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type UserRepository struct {
	db      *DBAdapter
	queries *sqlc.Queries
}

func NewUserRepository(db *DBAdapter, q *sqlc.Queries) *UserRepository {
	return &UserRepository{db: db, queries: q}
}

// GetByEmail returns a domain User pointer
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

// Create inserts a new user and returns a pointer to domain User
func (r *UserRepository) Create(ctx context.Context, u user.User) (*user.User, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role, status)
	 VALUES ($1, $2, $3, $4)
	 RETURNING id, email, password_hash, role, status, created_at`,
		u.Email, u.PasswordHash, u.Role, u.Status,
	)

	var created user.User
	var createdAt sql.NullTime
	if err := row.Scan(
		&created.ID,
		&created.Email,
		&created.PasswordHash,
		&created.Role,
		&created.Status,
		&createdAt,
	); err != nil {
		return nil, err
	}
	created.CreatedAt = createdAt.Time

	return &created, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]sqlc.ListUsersRow, error) {
	return r.queries.ListUsers(ctx)
}

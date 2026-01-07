package postgres

import (
	"context"

	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(q *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: q}
}

// GetByEmail returns a domain User
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,   // SQLC field
		Role:         row.Role,           // optional, adjust migration
		CreatedAt:    row.CreatedAt.Time, // pgtype → time.Time
	}, nil
}

// Create inserts a new user and returns a domain User
func (r *UserRepository) Create(ctx context.Context, u user.User) (*user.User, error) {
	row, err := r.queries.CreateUser(ctx, u.Email, u.PasswordHash)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

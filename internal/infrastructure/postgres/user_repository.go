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
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

// Create inserts a new user and returns a pointer to domain User
func (r *UserRepository) Create(ctx context.Context, u user.User) (*user.User, error) {
	params := sqlc.CreateUserParams{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	}

	row, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

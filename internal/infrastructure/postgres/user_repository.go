package postgres

import (
	"context"

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

// create inserts a new user and returns a pointer to domain User
func (r *UserRepository) Create(ctx context.Context, u user.User) (*user.User, error) {
	res, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Status:       u.Status,
	})
	if err != nil {
		return nil, err
	}

	// map the database result back to your Domain User
	return &user.User{
		ID:           res.ID,
		Email:        res.Email,
		PasswordHash: res.PasswordHash,
		Role:         res.Role,
		Status:       res.Status,
		CreatedAt:    res.CreatedAt.Time,
	}, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]user.UserListRow, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var result []user.UserListRow
	for _, row := range rows {
		result = append(result, user.UserListRow{
			ID:    row.ID,
			Email: row.Email,
		})
	}
	return result, nil
}

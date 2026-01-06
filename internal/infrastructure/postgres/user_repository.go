package postgres

import (
	"context"

	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{
		queries: queries,
	}
}

func (r *UserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	created, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          u.Email,
		HashedPassword: u.Password,
	})
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		ID:        int64(created.ID),
		Email:     created.Email,
		Password:  created.HashedPassword,
		CreatedAt: created.CreatedAt.Time,
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	found, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		ID:        int64(found.ID),
		Email:     found.Email,
		Password:  found.HashedPassword,
		CreatedAt: found.CreatedAt.Time,
	}, nil
}

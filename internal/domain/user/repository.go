package user

import (
	"context"
	"os/user"
)

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

func (r *UserRepository) GetByEmail(ctx, email) (*user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		CreatedAt:    row.CreatedAt.Time, // convert pgtype.Timestamptz → time.Time
	}, nil
}

package user

import (
	"context"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type Service interface {
	RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	ListAllUsers(ctx context.Context) ([]sqlc.ListUsersRow, error)
}

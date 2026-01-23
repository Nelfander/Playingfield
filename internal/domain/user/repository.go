package user

import (
	"context"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, user User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context) ([]UserListRow, error)
}

package user

import (
	"context"
	"time"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

// FakeRepository implements Repository for testing without a real DB
type FakeRepository struct {
	Users []User
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Users: []User{},
	}
}

func (f *FakeRepository) Create(ctx context.Context, u User) (*User, error) {
	for _, user := range f.Users {
		if user.Email == u.Email {
			return nil, ErrUserAlreadyExists
		}
	}

	u.ID = int64(len(f.Users) + 1)

	if u.Role == "" {
		u.Role = "user"
	}

	if u.Status == "" {
		u.Status = "active"
	}

	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	f.Users = append(f.Users, u)
	return &u, nil
}

func (f *FakeRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range f.Users {
		if u.Email == email {
			c := u
			return &c, nil
		}
	}
	return nil, ErrInvalidCredentials
}

func (f *FakeRepository) ListUsers(ctx context.Context) ([]sqlc.ListUsersRow, error) {
	var list []sqlc.ListUsersRow

	for _, u := range f.Users {
		list = append(list, sqlc.ListUsersRow{
			ID:    u.ID,
			Email: u.Email,
		})
	}

	return list, nil
}

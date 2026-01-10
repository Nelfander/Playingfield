package user

import (
	"context"
)

// FakeRepository implements Repository for testing without a real DB
type FakeRepository struct {
	users []User
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		users: []User{},
	}
}

func (f *FakeRepository) Create(ctx context.Context, u User) (*User, error) {
	for _, user := range f.users {
		if user.Email == u.Email {
			return nil, ErrUserAlreadyExists
		}
	}

	u.ID = int64(len(f.users) + 1)
	if u.Role == "" {
		u.Role = "user"
	}

	f.users = append(f.users, u)
	return &u, nil
}

func (f *FakeRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range f.users {
		if u.Email == email {
			c := u
			return &c, nil
		}
	}
	return nil, ErrInvalidCredentials
}

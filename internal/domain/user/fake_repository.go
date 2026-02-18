package user

import (
	"context"
	"fmt"
	"time"
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
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("fake repo: %w", err)
	}
	for _, user := range f.Users {
		if user.Email == u.Email {
			return nil, ErrUserAlreadyExists
		}
	}
	// simulate how a db handles primary keys
	u.ID = int64(len(f.Users) + 1)

	// fill in defaults if not set
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
	// return the version in the slice to be safe
	return &f.Users[len(f.Users)-1], nil
}

func (f *FakeRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("fake repo: %w", err)
	}
	for i := range f.Users { // look at the INDEX (position)
		if f.Users[i].Email == email {
			return &f.Users[i], nil // return the address of the actual slot in the slice
		}
	}
	return nil, ErrInvalidCredentials
}

func (f *FakeRepository) ListUsers(ctx context.Context) ([]UserListRow, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("fake repo: %w", err)
	}

	var result []UserListRow
	for _, u := range f.Users {
		result = append(result, UserListRow{
			ID:     u.ID,
			Email:  u.Email,
			Status: u.Status,
		})
	}
	return result, nil
}

func (f *FakeRepository) ScrubUser(ctx context.Context, userID int64) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("fake repo: %w", err)
	}

	found := false
	for i := range f.Users {
		if f.Users[i].ID == userID {
			// Simulate the SQL Scrub logic
			f.Users[i].Email = fmt.Sprintf("deleted_%d@playingfield.internal", userID)
			f.Users[i].PasswordHash = "SCRUBBED"
			f.Users[i].Status = "deleted"
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("fake repo: user not found")
	}

	return nil
}

func (f *FakeRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("fake repo: %w", err)
	}
	for i := range f.Users {
		if f.Users[i].ID == id {
			return &f.Users[i], nil
		}
	}
	return nil, fmt.Errorf("fake repo: user not found")
}

func (f *FakeRepository) UpdateUserStatus(ctx context.Context, userID int64, status string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("fake repo: %w", err)
	}
	for i := range f.Users {
		if f.Users[i].ID == userID {
			f.Users[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("fake repo: user not found")
}

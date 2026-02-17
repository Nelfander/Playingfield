package user

import (
	"context"
	"errors"
	"testing"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser_Service(t *testing.T) {
	repo := NewFakeRepository()
	svc := NewService(repo)
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		email := "new@example.com"
		u, err := svc.RegisterUser(ctx, email, "hashed_pass")

		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, email, u.Email)
		assert.Equal(t, "active", u.Status)
		assert.Equal(t, "user", u.Role)
	})

	t.Run("duplicate email returns sentinel error", func(t *testing.T) {
		email := "dup@example.com"
		_, _ = svc.RegisterUser(ctx, email, "pass1")

		u, err := svc.RegisterUser(ctx, email, "pass2")

		assert.Nil(t, u)
		assert.True(t, errors.Is(err, ErrUserAlreadyExists))
	})
}

func TestLogin_Service(t *testing.T) {
	repo := NewFakeRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Seed a user
	password := "correct_password"
	hashed, _ := auth.HashPassword(password)
	email := "login_test@example.com"
	_, _ = svc.RegisterUser(ctx, email, hashed)

	t.Run("successful login", func(t *testing.T) {
		u, err := svc.Login(ctx, email, password)
		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, email, u.Email)
	})

	t.Run("wrong password returns invalid credentials", func(t *testing.T) {
		u, err := svc.Login(ctx, email, "wrong_password")
		assert.Nil(t, u)
		assert.True(t, errors.Is(err, ErrInvalidCredentials))
	})

	t.Run("non-existent user returns invalid credentials", func(t *testing.T) {
		u, err := svc.Login(ctx, "fake@example.com", password)
		assert.Nil(t, u)
		assert.True(t, errors.Is(err, ErrInvalidCredentials))
	})

	t.Run("inactive user cannot login", func(t *testing.T) {
		// Manually set a user to inactive in the fake repo
		for i := range repo.Users {
			if repo.Users[i].Email == email {
				repo.Users[i].Status = "inactive"
			}
		}

		u, err := svc.Login(ctx, email, password)
		assert.Nil(t, u)
		assert.True(t, errors.Is(err, ErrInactiveAccount))
	})
}

func TestListAllUsers_Service(t *testing.T) {
	repo := NewFakeRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// seed multiple users directly into repo
	repo.Users = []User{
		{ID: 1, Email: "alice@example.com", Status: "active"},
		{ID: 2, Email: "bob@example.com", Status: "active"},
	}

	t.Run("list all users successfully", func(t *testing.T) {
		users, err := svc.ListAllUsers(ctx)

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "alice@example.com", users[0].Email)
		assert.Equal(t, "bob@example.com", users[1].Email)
	})

	t.Run("returns empty list if no users exist", func(t *testing.T) {
		repo.Users = []User{} // Clear repo
		users, err := svc.ListAllUsers(ctx)

		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestAdminScrubUser_Service(t *testing.T) {
	repo := NewFakeRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Seed a user to scrub
	targetID := int64(1)
	repo.Users = []User{
		{ID: targetID, Email: "fire_me@example.com", PasswordHash: "secret", Status: "active"},
	}

	t.Run("successful scrub", func(t *testing.T) {
		err := svc.AdminScrubUser(ctx, targetID)
		assert.NoError(t, err)

		// Verify the user is still in the slice but "scrubbed"
		assert.Len(t, repo.Users, 1)
		u := repo.Users[0]
		assert.Contains(t, u.Email, "deleted_1")
		assert.Equal(t, "SCRUBBED", u.PasswordHash)
		assert.Equal(t, "deleted", u.Status)
	})

	t.Run("scrub non-existent user returns error", func(t *testing.T) {
		err := svc.AdminScrubUser(ctx, 999) // ID doesn't exist
		assert.Error(t, err)
	})
}

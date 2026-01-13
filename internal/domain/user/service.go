package user

import (
	"context"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// RegisterUser creates a new user
func (s *service) RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing.ID != 0 {
		return nil, ErrUserAlreadyExists
	}

	u := User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
		Status:       "active",
	}

	createdUser, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	// Ensure role is set correctly (in case the repository/fake repo didn't preserve it)
	if createdUser.Role == "" {
		createdUser.Role = "user"
	}

	return createdUser, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	if u.Status != "active" {
		return nil, ErrInactiveAccount
	}

	return u, nil
}

func (s *service) ListAllUsers(ctx context.Context) ([]sqlc.ListUsersRow, error) {
	// This calls the generated code you just verified in Step 18
	return s.repo.ListUsers(ctx)
}

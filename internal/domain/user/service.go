package user

import (
	"context"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// RegisterUser creates a new user
func (s *Service) RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing.ID != 0 {
		return nil, ErrUserAlreadyExists
	}

	u := User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	return s.repo.Create(ctx, u)
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return u, nil
}

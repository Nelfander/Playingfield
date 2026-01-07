package user

import (
	"context"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) RegisterUser(
	ctx context.Context,
	email string,
	hashedPassword string,
) (User, error) {

	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing.ID != 0 {
		return User{}, ErrUserAlreadyExists
	}

	u := User{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	return s.repo.Create(ctx, u)
}

// Login verifies credentials and returns a domain User
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

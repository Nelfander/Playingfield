package user

import (
	"context"
	"errors"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveAccount    = errors.New("account is inactive or banned")
)

type service struct {
	repo Repository
}

type UserListRow struct {
	ID    int64
	Email string
}

type Service interface {
	RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	ListAllUsers(ctx context.Context) ([]UserListRow, error)
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrUserAlreadyExists
	}
	u := User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
		Status:       "active",
	}
	return s.repo.Create(ctx, u)
}

func (s *service) Login(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil || u == nil {
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

func (s *service) ListAllUsers(ctx context.Context) ([]UserListRow, error) {
	return s.repo.ListUsers(ctx)
}

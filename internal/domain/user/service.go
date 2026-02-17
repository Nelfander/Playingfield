package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	AdminScrubUser(ctx context.Context, userID int64) error
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		slog.Warn("registration attempt with existing email", "email", email)
		return nil, ErrUserAlreadyExists
	}
	u := User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
		Status:       "active",
	}
	created, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("new user registered", "user_id", created.ID, "email", created.Email)
	return created, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil || u == nil {
		// could be brute force attempt, so log the attempt
		slog.Warn("login failed: user not found", "email", email)
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, u.PasswordHash) {
		slog.Warn("login failed: wrong password", "user_id", u.ID, "email", email)
		return nil, ErrInvalidCredentials
	}

	if u.Status != "active" {
		slog.Warn("login attempt for inactive account", "user_id", u.ID, "status", u.Status)
		return nil, ErrInactiveAccount
	}

	slog.Info("user logged in", "user_id", u.ID)

	return u, nil
}

func (s *service) ListAllUsers(ctx context.Context) ([]UserListRow, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *service) AdminScrubUser(ctx context.Context, userID int64) error {
	slog.Info("admin initiating user scrub", "target_user_id", userID)

	err := s.repo.ScrubUser(ctx, userID)
	if err != nil {
		slog.Error("failed to scrub user", "user_id", userID, "error", err)
		return fmt.Errorf("service: scrub user failed: %w", err)
	}

	slog.Info("user scrubbed successfully", "user_id", userID)
	return nil
}

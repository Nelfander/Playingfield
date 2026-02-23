package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveAccount    = errors.New("account is inactive or banned")
)

type service struct {
	repo Repository
	hub  *ws.Hub
}

type UserListRow struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type Service interface {
	RegisterUser(ctx context.Context, email, hashedPassword string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	ListAllUsers(ctx context.Context) ([]UserListRow, error)
	AdminScrubUser(ctx context.Context, actor *auth.Claims, targetUserID int64) error
	ToggleUserStatus(ctx context.Context, actor *auth.Claims, targetUserID int64) error
}

func NewService(repo Repository, hub *ws.Hub) Service {
	return &service{
		repo: repo,
		hub:  hub,
	}
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

	if s.hub != nil {
		notification := fmt.Sprintf("USER_CREATED:%d", created.ID)
		s.hub.Broadcast <- []byte(notification)
	}

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
	// get everything from the repo (including the scrubbed ones)
	allUsers, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// create a slice for the "clean" data
	// pre-allocate with a capacity of len(allUsers) for efficiency
	filteredUsers := make([]UserListRow, 0, len(allUsers))

	// if the status is NOT 'deleted', keep them
	for _, u := range allUsers {
		// Clean the string before checking
		cleanStatus := strings.ToLower(strings.TrimSpace(u.Status))

		if cleanStatus != "deleted" {
			filteredUsers = append(filteredUsers, u)
		}
	}

	return filteredUsers, nil
}

// Deletes user but keeps history
func (s *service) AdminScrubUser(ctx context.Context, actor *auth.Claims, targetUserID int64) error {
	if actor == nil {
		return fmt.Errorf("missing authentication claims")
	}

	if !auth.IsAdmin(actor.Role) {
		slog.Warn("non-admin attempted to scrub user",
			"user_id", actor.UserID,
			"role", actor.Role,
			"target_id", targetUserID)
		return fmt.Errorf("only administrators can scrub users")
	}

	if actor.UserID == targetUserID {
		slog.Warn("self-scrub attempt blocked",
			"user_id", actor.UserID)
		return fmt.Errorf("cannot scrub your own account")
	}

	slog.Info("admin scrubbing user",
		"admin_id", actor.UserID,
		"role", actor.Role,
		"target_id", targetUserID)

	// proceed to the repo only if the check passes
	if err := s.repo.ScrubUser(ctx, targetUserID); err != nil {
		slog.Error("failed to scrub user",
			"target_id", targetUserID,
			"error", err)
		return fmt.Errorf("failed to scrub user: %w", err)
	}

	slog.Info("user scrubbed successfully", "target_id", targetUserID)

	if s.hub != nil {
		notification := fmt.Sprintf("USER_SCRUBBED:%d", targetUserID)
		s.hub.Broadcast <- []byte(notification)
	}
	return nil
}

func (s *service) ToggleUserStatus(ctx context.Context, actor *auth.Claims, targetUserID int64) error {
	if actor == nil {
		return fmt.Errorf("missing authentication claims")
	}

	if !auth.IsAdmin(actor.Role) {
		slog.Warn("non-admin attempted to toggle user status",
			"user_id", actor.UserID,
			"role", actor.Role,
			"target_id", targetUserID)
		return fmt.Errorf("only administrators can toggle user status")
	}

	if actor.UserID == targetUserID {
		return fmt.Errorf("cannot toggle your own account status")
	}

	u, err := s.repo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	newStatus := "active"
	if u.Status == "active" {
		newStatus = "inactive"
	}

	slog.Info("user status toggled",
		"admin_id", actor.UserID,
		"target_id", targetUserID,
		"from", u.Status,
		"to", newStatus)

	if err := s.repo.UpdateUserStatus(ctx, targetUserID, newStatus); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	if s.hub != nil {
		notification := fmt.Sprintf("USER_UPDATED:%d:%s", targetUserID, newStatus)
		s.hub.Broadcast <- []byte(notification)
	}

	return nil
}

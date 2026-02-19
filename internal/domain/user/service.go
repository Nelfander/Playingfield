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
	AdminScrubUser(ctx context.Context, adminID int64, targetUserID int64) error
	ToggleUserStatus(ctx context.Context, userID int64) error
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

func (s *service) AdminScrubUser(ctx context.Context, requesterID int64, targetUserID int64) error {
	// compare IDs before doing anything else
	if requesterID == targetUserID {
		slog.Warn("Self-scrub attempt blocked", "admin_id", requesterID)
		return fmt.Errorf("self-preservation active: you cannot delete the account you are currently logged into")
	}

	slog.Info("admin initiating user scrub", "admin_id", requesterID, "target_user_id", targetUserID)

	// proceed to the repo only if the check passes
	err := s.repo.ScrubUser(ctx, targetUserID)
	if err != nil {
		slog.Error("failed to scrub user", "target_id", targetUserID, "error", err)
		return fmt.Errorf("service: scrub user failed: %w", err)
	}

	slog.Info("user scrubbed successfully", "target_id", targetUserID)

	if s.hub != nil {
		notification := fmt.Sprintf("USER_SCRUBBED:%d", targetUserID)
		s.hub.Broadcast <- []byte(notification)
	}
	return nil
}

func (s *service) ToggleUserStatus(ctx context.Context, userID int64) error {
	// fetch the user to see current status
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("service: toggle status failed to fetch user: %w", err)
	}

	// switch
	newStatus := "active"
	if u.Status == "active" {
		newStatus = "inactive"
	}

	if s.hub != nil {
		notification := fmt.Sprintf("USER_UPDATED:%d:%s", userID, newStatus)
		s.hub.Broadcast <- []byte(notification)
	}

	return s.repo.UpdateUserStatus(ctx, userID, newStatus)
}

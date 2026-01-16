package projects

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
)

type Service struct {
	repo  Repository
	store *sqlc.Queries
	hub   *ws.Hub
}

func (s *Service) ListUsersInProject(projectID int64) ([]sqlc.ListUsersInProjectRow, error) {
	return s.store.ListUsersInProject(context.Background(), projectID)
}

func NewService(repo Repository, store *sqlc.Queries, hub *ws.Hub) *Service {
	return &Service{
		repo:  repo,
		store: store,
		hub:   hub,
	}
}

func (s *Service) CreateProject(ctx context.Context, name, description string, ownerID int64) (*Project, error) {
	p := Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}

	project, err := s.repo.Create(ctx, p)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("you already have a project with the name '%s'", name)
		}
		return nil, err
	}

	arg := sqlc.AddUserToProjectParams{
		ProjectID: project.ID,
		UserID:    ownerID,
		Role:      pgtype.Text{String: "owner", Valid: true},
	}

	_, err = s.store.AddUserToProject(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("project created but failed to assign ownership: %w", err)
	}

	// BROADCAST: Only signal success if the database work is fully done.
	s.hub.Broadcast <- []byte("PROJECT_CREATED")

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, ownerID int64) ([]Project, error) {
	return s.repo.GetAllByOwner(ctx, ownerID)
}

func (s *Service) DeleteProject(ctx context.Context, projectID, ownerID int64) error {
	project, err := s.store.GetProjectByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	if project.OwnerID != ownerID {
		return fmt.Errorf("only the project owner can delete this project")
	}

	err = s.store.DeleteProject(ctx, sqlc.DeleteProjectParams{
		ID:      projectID,
		OwnerID: ownerID,
	})

	if err != nil {
		return err
	}

	notification := fmt.Sprintf("PROJECT_DELETED:%d", projectID)
	s.hub.Broadcast <- []byte(notification)

	return nil
}

func (s *Service) AddUserToProject(projectID, userID int64, role string) (sqlc.ProjectUser, error) {
	arg := sqlc.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      pgtype.Text{String: role, Valid: true},
	}

	// 1. Perform the actual database insertion
	projectUser, err := s.store.AddUserToProject(context.Background(), arg)
	if err != nil {
		return projectUser, err
	}

	//  Broadcast the change to the Hub
	notification := fmt.Sprintf("USER_ADDED:%d:%d:%s", projectID, userID, role)
	s.hub.Broadcast <- []byte(notification)

	return projectUser, nil
}

func (s *Service) RemoveUserFromProject(requesterID, projectID, userID int64) error {
	project, err := s.store.GetProjectByID(context.Background(), projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	if project.OwnerID != requesterID {
		return fmt.Errorf("only the project owner can remove users")
	}

	err = s.store.RemoveUserFromProject(context.Background(), sqlc.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}

	notification := fmt.Sprintf("USER_REMOVED:%d:%d", projectID, userID)
	s.hub.Broadcast <- []byte(notification)

	return nil
}

func (s *Service) GetProject(ctx context.Context, id int64) (*Project, error) {
	return s.repo.GetByID(ctx, id)
}

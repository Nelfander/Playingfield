package projects

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type Service struct {
	repo  Repository
	store *sqlc.Queries
}

func (s *Service) ListUsersInProject(projectID int64) ([]sqlc.ListUsersInProjectRow, error) {
	return s.store.ListUsersInProject(context.Background(), projectID)
}

func NewService(repo Repository, store *sqlc.Queries) *Service {
	return &Service{
		repo:  repo,
		store: store,
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
		// Check if it's a PostgreSQL unique violation (per-user project name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, fmt.Errorf("you already have a project with the name '%s'", name)
		}
		return nil, err
	}

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, ownerID int64) ([]Project, error) {
	return s.repo.GetAllByOwner(ctx, ownerID)
}

func (s *Service) AddUserToProject(projectID, userID int64, role string) (sqlc.ProjectUser, error) {
	arg := sqlc.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      pgtype.Text{String: role, Valid: true},
	}
	return s.store.AddUserToProject(context.Background(), arg)
}

func (s *Service) RemoveUserFromProject(requesterID, projectID, userID int64) error {
	project, err := s.store.GetProjectByID(context.Background(), projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	if project.OwnerID != requesterID {
		return fmt.Errorf("only the project owner can remove users")
	}

	return s.store.RemoveUserFromProject(context.Background(), sqlc.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})

}

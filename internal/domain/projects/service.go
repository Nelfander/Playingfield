package projects

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
)

type Service struct {
	repo Repository
	hub  *ws.Hub
}

func NewService(repo Repository, hub *ws.Hub) *Service {
	return &Service{
		repo: repo,
		hub:  hub,
	}
}

func (s *Service) ListUsersInProject(ctx context.Context, projectID int64) ([]ProjectMember, error) {
	return s.repo.ListUsersInProject(ctx, projectID)
}

func (s *Service) CreateProject(ctx context.Context, name, description string, ownerID int64) (*Project, error) {
	p := Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}

	project, err := s.repo.CreateProject(ctx, p)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("you already have a project with the name '%s'", name)
		}
		return nil, err
	}

	// This will call the Fake in tests and the Real DB in production
	err = s.repo.AddUserToProject(ctx, project.ID, ownerID, "owner")
	if err != nil {
		return nil, fmt.Errorf("project created but failed to assign ownership: %w", err)
	}

	// (a nil-check for the hub to prevent panics in tests)
	if s.hub != nil {
		s.hub.Broadcast <- []byte("PROJECT_CREATED")
	}
	return project, nil
}

func (s *Service) UpdateProject(ctx context.Context, requesterID, projectID int64, name, description string) (*Project, error) {
	// get current project to check ownership
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	//  only owner can update
	if project.OwnerID != requesterID {
		return nil, fmt.Errorf("unauthorized: user %d is not the owner", requesterID)
	}

	// update the fields
	project.Name = name
	project.Description = description

	// update in the database
	updatedProject, err := s.repo.Update(ctx, *project)
	if err != nil {
		return nil, err
	}

	// broadcast the change to the Hub
	if s.hub != nil {
		notification := fmt.Sprintf("PROJECT_UPDATED:%d", projectID)
		s.hub.Broadcast <- []byte(notification)
	}

	return updatedProject, nil
}

func (s *Service) ListProjects(ctx context.Context, ownerID int64) ([]Project, error) {
	return s.repo.GetAllByOwner(ctx, ownerID)
}

func (s *Service) DeleteProject(ctx context.Context, projectID, ownerID int64) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	if project.OwnerID != ownerID {
		return fmt.Errorf("only the project owner can delete this project")
	}

	//  repo.DeleteProject (Safe for tests)
	err = s.repo.DeleteProject(ctx, projectID, ownerID)
	if err != nil {
		return err
	}

	if s.hub != nil {
		notification := fmt.Sprintf("PROJECT_DELETED:%d", projectID)
		s.hub.Broadcast <- []byte(notification)
	}

	return nil
}

func (s *Service) AddUserToProject(ctx context.Context, requesterID int64, projectID int64, userID int64, role string) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// only project owner can add members
	if project.OwnerID != requesterID {
		return fmt.Errorf("unauthorized: user %d is not the owner of project %d", requesterID, projectID)
	}

	// duplicate check
	members, err := s.repo.ListUsersInProject(ctx, projectID)
	if err == nil {
		for _, m := range members {
			if m.ID == userID {
				return fmt.Errorf("user is already a member of this project")
			}
		}

	}

	// add the user
	err = s.repo.AddUserToProject(ctx, projectID, userID, role)
	if err != nil {
		return err
	}

	// broadcast the change
	if s.hub != nil {
		notification := fmt.Sprintf("USER_ADDED:%d:%d:%s", projectID, userID, role)
		s.hub.Broadcast <- []byte(notification)
	}

	return nil
}

func (s *Service) RemoveUserFromProject(requesterID, projectID, userID int64) error {
	project, err := s.repo.GetByID(context.Background(), projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	if project.OwnerID != requesterID {
		return fmt.Errorf("only the project owner can remove users")
	}

	err = s.repo.RemoveUserFromProject(context.Background(), projectID, userID)
	if err != nil {
		return err
	}

	if s.hub != nil {
		notification := fmt.Sprintf("USER_REMOVED:%d:%d", projectID, userID)
		s.hub.Broadcast <- []byte(notification)
	}

	return nil
}

func (s *Service) GetProject(ctx context.Context, id int64) (*Project, error) {
	return s.repo.GetByID(ctx, id)
}

package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type ProjectRepository struct {
	db      *DBAdapter
	queries *sqlc.Queries
}

func NewProjectRepository(db *DBAdapter) *ProjectRepository {
	return &ProjectRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, p projects.Project) (*projects.Project, error) {
	// call generated SQLC method
	res, err := r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		Name:        p.Name,
		Description: pgtype.Text{String: p.Description, Valid: p.Description != ""},
		OwnerID:     p.OwnerID,
	})
	if err != nil {
		return nil, fmt.Errorf("db: create project: %w", err)
	}

	// info because projects are significant events!
	slog.Info("new project created", "project_id", res.ID, "name", res.Name, "owner_id", res.OwnerID)

	// map it back to /domain/project
	return &projects.Project{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description.String,
		OwnerID:     res.OwnerID,
		CreatedAt:   res.CreatedAt.Time,
		OwnerName:   res.OwnerName,
	}, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p projects.Project) (*projects.Project, error) {
	err := r.queries.UpdateProject(ctx, sqlc.UpdateProjectParams{
		ID:   p.ID,
		Name: p.Name,
		Description: pgtype.Text{
			String: p.Description,
			Valid:  p.Description != "",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("db: update project: %w", err)
	}
	return &p, nil
}

// GetAllByOwner fetches all projects the user owns OR is a member of
func (r *ProjectRepository) GetAllByOwner(ctx context.Context, ownerID int64) ([]projects.Project, error) {
	rows, err := r.queries.ListProjectsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("db: list projects by owner: %w", err)
	}

	// map the slice of SQLC rows to a slice of domain projects
	var list []projects.Project
	for _, row := range rows {
		list = append(list, projects.Project{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description.String,
			OwnerID:     row.OwnerID,
			CreatedAt:   row.CreatedAt.Time,
			OwnerName:   row.OwnerName.String,
		})
	}

	return list, nil
}
func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*projects.Project, error) {
	res, err := r.queries.GetProjectByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("db: get project by id: %w", err)
	}

	// MAP: Convert SQLC types to Domain types
	return &projects.Project{
		ID:   res.ID,
		Name: res.Name,
		// pgtype.Text -> string
		Description: res.Description.String,
		OwnerID:     res.OwnerID,
		// pgtype.Timestamp/Timestamptz -> time.Time
		CreatedAt: res.CreatedAt.Time,
	}, nil
}

func (r *ProjectRepository) ListUsersInProject(ctx context.Context, projectID int64) ([]projects.ProjectMember, error) {
	rows, err := r.queries.ListUsersInProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("db: list users in project: %w", err)
	}

	var members []projects.ProjectMember
	for _, row := range rows {
		// Use type assertion .(string) to convert interface{} to string
		roleStr, _ := row.Role.(string)
		if roleStr == "" {
			roleStr = "member"
		}

		members = append(members, projects.ProjectMember{
			ID:    row.ID,
			Email: row.Email,
			Role:  roleStr,
		})
	}

	return members, nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, id int64, ownerID int64) error {
	err := r.queries.DeleteProject(ctx, sqlc.DeleteProjectParams{
		ID:      id,
		OwnerID: ownerID,
	})
	if err != nil {
		return fmt.Errorf("db: delete project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) AddUserToProject(ctx context.Context, projectID int64, userID int64, role string) error {
	_, err := r.queries.AddUserToProject(ctx, sqlc.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      pgtype.Text{String: role, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("db: add user to project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error {
	err := r.queries.RemoveUserFromProject(ctx, sqlc.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("db: remove user from project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) UsersShareProject(ctx context.Context, userA, userB int64) (bool, error) {
	shared, err := r.queries.CheckSharedProject(ctx, sqlc.CheckSharedProjectParams{
		SenderID:   userA,
		ReceiverID: userB,
	})
	if err != nil {
		return false, fmt.Errorf("db: check shared project: %w", err)
	}
	return shared, nil
}

package postgres

import (
	"context"

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
		return nil, err
	}

	// map it back to /domain/project
	return &projects.Project{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description.String,
		OwnerID:     res.OwnerID,
		CreatedAt:   res.CreatedAt.Time,
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
		return nil, err
	}
	return &p, nil
}

// GetAllByOwner fetches all projects the user owns OR is a member of
func (r *ProjectRepository) GetAllByOwner(ctx context.Context, ownerID int64) ([]projects.Project, error) {
	rows, err := r.queries.ListProjectsByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// MAP: Convert SQLC types to Domain types
	return &projects.Project{
		ID:   res.ID,
		Name: res.Name,

		// pgtype.Text -> string
		Description: res.Description.String,

		OwnerID: res.OwnerID,

		// pgtype.Timestamp/Timestamptz -> time.Time
		CreatedAt: res.CreatedAt.Time,
	}, nil
}

func (r *ProjectRepository) ListUsersInProject(ctx context.Context, projectID int64) ([]projects.ProjectMember, error) {
	rows, err := r.queries.ListUsersInProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var members []projects.ProjectMember
	for _, row := range rows {
		// Use type assertion .(string) to convert interface{} to string
		roleStr, ok := row.Role.(string)
		if !ok {
			roleStr = "member" // fallback safety
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
	return r.queries.DeleteProject(ctx, sqlc.DeleteProjectParams{
		ID:      id,
		OwnerID: ownerID,
	})
}

func (r *ProjectRepository) AddUserToProject(ctx context.Context, projectID int64, userID int64, role string) error {
	// Uses the :one query you defined
	_, err := r.queries.AddUserToProject(ctx, sqlc.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      pgtype.Text{String: role, Valid: true},
	})
	return err
}

func (r *ProjectRepository) RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error {
	return r.queries.RemoveUserFromProject(ctx, sqlc.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
}

func (r *ProjectRepository) UsersShareProject(ctx context.Context, userA, userB int64) (bool, error) {
	shared, err := r.queries.CheckSharedProject(ctx, sqlc.CheckSharedProjectParams{
		SenderID:   userA,
		ReceiverID: userB,
	})
	if err != nil {
		return false, err
	}
	return shared, nil
}

package postgres

import (
	"context"
	"time"

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

func (r *ProjectRepository) Create(ctx context.Context, p projects.Project) (*projects.Project, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO projects (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, created_at`,
		p.Name, p.Description, p.OwnerID,
	)

	var created projects.Project
	var createdAt time.Time
	if err := row.Scan(
		&created.ID,
		&created.Name,
		&created.Description,
		&created.OwnerID,
		&createdAt,
	); err != nil {
		return nil, err
	}
	created.CreatedAt = createdAt
	return &created, nil
}

// GetAllByOwner fetches all projects the user owns OR is a member of
func (r *ProjectRepository) GetAllByOwner(ctx context.Context, ownerID int64) ([]projects.Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT 
            p.id, 
            p.name, 
            p.description, 
            p.owner_id, 
            p.created_at,
            u.email AS owner_name
         FROM projects p
         LEFT JOIN users u ON p.owner_id = u.id
         WHERE p.owner_id = $1 
            OR p.id IN (SELECT project_id FROM project_users WHERE user_id = $1)
         ORDER BY p.created_at ASC`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []projects.Project
	for rows.Next() {
		var p projects.Project
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.OwnerID,
			&p.CreatedAt,
			&p.OwnerName,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
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

func (r *ProjectRepository) DeleteProject(ctx context.Context, id int64, ownerID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM projects WHERE id = $1 AND owner_id = $2`,
		id, ownerID,
	)
	return err
}

func (r *ProjectRepository) AddUserToProject(ctx context.Context, projectID int64, userID int64, role string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, $3)`,
		projectID, userID, role,
	)
	return err
}

func (r *ProjectRepository) RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error {
	return r.queries.RemoveUserFromProject(ctx, sqlc.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	})
}

func (r *ProjectRepository) ListUsers(ctx context.Context, projectID int64) ([]sqlc.ListUsersInProjectRow, error) {
	return r.queries.ListUsersInProject(ctx, projectID)
}

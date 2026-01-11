package postgres

import (
	"context"
	"time"

	"github.com/nelfander/Playingfield/internal/domain/projects"
)

type ProjectRepository struct {
	db *DBAdapter
}

func NewProjectRepository(db *DBAdapter) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create inserts a new project
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

// GetAllByOwner fetches all projects for a specific owner
func (r *ProjectRepository) GetAllByOwner(ctx context.Context, ownerID int64) ([]projects.Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, owner_id, created_at
		 FROM projects
		 WHERE owner_id = $1`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []projects.Project
	for rows.Next() {
		var p projects.Project
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt
		list = append(list, p)
	}

	return list, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*projects.Project, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, description, owner_id, created_at
         FROM projects
         WHERE id = $1`,
		id,
	)

	var p projects.Project
	var createdAt time.Time
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &createdAt); err != nil {
		return nil, err
	}
	p.CreatedAt = createdAt
	return &p, nil
}

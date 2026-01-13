package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type Repository interface {
	Create(ctx context.Context, user User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context) ([]sqlc.ListUsersRow, error)
}

// repo implementation
type repo struct {
	db *sql.DB
}

// NewRepository constructor
func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}

}

// ListUsers fetches all users from the database so that they can be added to a project
func (r *repo) ListUsers(ctx context.Context) ([]sqlc.ListUsersRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, email FROM users ORDER BY email ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []sqlc.ListUsersRow
	for rows.Next() {
		var u sqlc.ListUsersRow
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Create inserts a new user into the database
func (r *repo) Create(ctx context.Context, u User) (*User, error) {
	fmt.Printf("Running INSERT: email=%s, role=%s, status=%s\n", u.Email, u.Role, u.Status)
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, role, status)
         VALUES ($1, $2, $3, $4)
         RETURNING id, email, password_hash, role, status, created_at`,
		u.Email, u.PasswordHash, u.Role, u.Status,
	)

	var created User
	if err := row.Scan(
		&created.ID,
		&created.Email,
		&created.PasswordHash,
		&created.Role,
		&created.Status,
		&created.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &created, nil
}

// GetByEmail fetches a user by email
func (r *repo) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, role, status, created_at
         FROM users
         WHERE email = $1`, email)

	var u User
	if err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

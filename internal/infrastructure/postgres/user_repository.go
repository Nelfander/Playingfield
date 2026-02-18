package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/domain/user"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type UserRepository struct {
	db      *DBAdapter
	queries *sqlc.Queries
}

func NewUserRepository(db *DBAdapter, q *sqlc.Queries) *UserRepository {
	return &UserRepository{db: db, queries: q}
}

// GetByEmail returns a domain User pointer
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Error("database: get user by email failed", "email", email, "error", err)
		return nil, fmt.Errorf("db: get user by email: %w", err)
	}

	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

// create inserts a new user and returns a pointer to domain User
func (r *UserRepository) Create(ctx context.Context, u user.User) (*user.User, error) {
	res, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Status:       u.Status,
	})
	if err != nil {
		slog.Error("database: user creation failed", "email", u.Email, "error", err)
		return nil, fmt.Errorf("db: create user: %w", err)
	}

	// map the database result back to your Domain User
	return &user.User{
		ID:           res.ID,
		Email:        res.Email,
		PasswordHash: res.PasswordHash,
		Role:         res.Role,
		Status:       res.Status,
		CreatedAt:    res.CreatedAt.Time,
	}, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]user.UserListRow, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		slog.Error("database: list users failed", "error", err)
		return nil, fmt.Errorf("db: list users: %w", err)
	}

	var result []user.UserListRow
	for _, row := range rows {
		result = append(result, user.UserListRow{
			ID:     row.ID,
			Email:  row.Email,
			Status: row.Status,
		})
	}
	return result, nil
}

func (r *UserRepository) ScrubUser(ctx context.Context, userID int64) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		slog.Error("database: failed to begin transaction", "error", err)
		return fmt.Errorf("db: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if err = qtx.DeleteProjectsByOwner(ctx, userID); err != nil {
		return fmt.Errorf("db tx: delete projects: %w", err)
	}

	if err = qtx.RemoveUserFromAllProjectMemberships(ctx, userID); err != nil {
		return fmt.Errorf("db tx: remove memberships: %w", err)
	}

	// this one specifically requires pgtype.Int8 because 'assigned_to' is nullable in schema
	dbID := pgtype.Int8{Int64: userID, Valid: true}
	if err = qtx.UnassignUserFromAllTasks(ctx, dbID); err != nil {
		return fmt.Errorf("db tx: unassign tasks: %w", err)
	}

	if err = qtx.ScrubUserAccount(ctx, userID); err != nil {
		return fmt.Errorf("db tx: scrub account: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *UserRepository) UpdateUserStatus(ctx context.Context, userID int64, status string) error {
	return r.queries.UpdateUserStatus(ctx, sqlc.UpdateUserStatusParams{
		ID:     userID,
		Status: status,
	})
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("db: get user by id: %w", err)
	}

	return &user.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         row.Role,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

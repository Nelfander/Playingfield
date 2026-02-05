package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/domain/tasks"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type TaskRepository struct {
	db      *DBAdapter
	queries *sqlc.Queries
}

func NewTaskRepository(db *DBAdapter) *TaskRepository {
	return &TaskRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *TaskRepository) CreateTask(ctx context.Context, t *tasks.Task) (*tasks.Task, error) {
	// assigned to can be empty when the task is beeing created,
	// the project owner might make tasks which he still doesnt know which members to assign them to!
	res, err := r.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ProjectID:   t.ProjectID,
		Title:       t.Title,
		Description: pgtype.Text{String: t.Description, Valid: t.Description != ""},
		Status:      t.Status,
		AssignedTo:  toPgInt8(t.AssignedTo),
	})
	if err != nil {
		slog.Error("database: create task failed", "project_id", t.ProjectID, "title", t.Title, "error", err)
		return nil, fmt.Errorf("db: create task: %w", err)
	}

	return mapSQLCTaskToDomain(res), nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, t *tasks.Task) (*tasks.Task, error) {
	res, err := r.queries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:          t.ID,
		Title:       t.Title,
		Description: pgtype.Text{String: t.Description, Valid: t.Description != ""},
		Status:      t.Status,
		AssignedTo:  toPgInt8(t.AssignedTo),
	})
	if err != nil {
		slog.Error("database: update task failed", "task_id", t.ID, "error", err)
		return nil, fmt.Errorf("db: update task: %w", err)
	}

	return mapSQLCTaskToDomain(res), nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, id int64) error {
	if err := r.queries.DeleteTask(ctx, id); err != nil {
		slog.Error("database: delete task failed", "task_id", id, "error", err)
		return fmt.Errorf("db: delete task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, id int64) (*tasks.Task, error) {
	res, err := r.queries.GetTaskByID(ctx, id)
	if err != nil {
		slog.Error("database: get task by id failed", "task_id", id, "error", err)
		return nil, fmt.Errorf("db: get task by id: %w", err)
	}
	return mapSQLCTaskToDomain(res), nil
}

func (r *TaskRepository) ListTaskByProject(ctx context.Context, projectID int64) ([]*tasks.Task, error) {
	rows, err := r.queries.ListTasksForProject(ctx, projectID)
	if err != nil {
		slog.Error("database: list tasks by project failed", "project_id", projectID, "error", err)
		return nil, fmt.Errorf("db: list tasks by project: %w", err)
	}

	var list []*tasks.Task
	for _, row := range rows {
		// safe pointer check for the nullable AssignedTo field
		var assignedID *int64
		if row.AssignedTo.Valid {
			id := row.AssignedTo.Int64
			assignedID = &id
		}

		list = append(list, &tasks.Task{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			Title:       row.Title,
			Description: row.Description.String,
			Status:      row.Status,
			AssignedTo:  assignedID, // Safe nil or pointer
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return list, nil
}

func (r *TaskRepository) RecordTaskActivity(ctx context.Context, a *tasks.TaskActivity) error {
	err := r.queries.RecordTaskActivity(ctx, sqlc.RecordTaskActivityParams{
		TaskID:  a.TaskID,
		UserID:  a.UserID,
		Action:  a.Action,
		Details: pgtype.Text{String: a.Details, Valid: a.Details != ""},
	})
	if err != nil {
		slog.Error("database: record task activity failed", "task_id", a.TaskID, "action", a.Action, "error", err)
		return fmt.Errorf("db: record task activity: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetTaskHistory(ctx context.Context, taskID int64) ([]*tasks.TaskActivity, error) {
	rows, err := r.queries.GetTaskHistory(ctx, taskID)
	if err != nil {
		slog.Error("database: get task history failed", "task_id", taskID, "error", err)
		return nil, fmt.Errorf("db: get task history: %w", err)
	}

	var history []*tasks.TaskActivity
	for _, row := range rows {
		history = append(history, &tasks.TaskActivity{
			ID:        row.ID,
			TaskID:    row.TaskID,
			UserID:    row.UserID,
			UserEmail: row.UserEmail,
			Action:    row.Action,
			Details:   row.Details.String,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return history, nil
}

// Helper: Mapper logic to keep things clean
func mapSQLCTaskToDomain(row sqlc.Task) *tasks.Task {
	var assignedID *int64
	if row.AssignedTo.Valid {
		id := row.AssignedTo.Int64
		assignedID = &id
	}

	return &tasks.Task{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		Title:       row.Title,
		Description: row.Description.String,
		Status:      row.Status,
		AssignedTo:  assignedID,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

// helper: toPgInt8 converts a nullable *int64 to a pgtype.Int8
func toPgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

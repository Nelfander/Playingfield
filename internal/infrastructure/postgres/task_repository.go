package postgres

import (
	"context"

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
	// assigned to can be empty when the task is beeing created. The project owner might make tasks which he still doesnt know which members to assign them to!
	var assignedTo pgtype.Int8
	if t.AssignedTo != nil {
		assignedTo = pgtype.Int8{Int64: *t.AssignedTo, Valid: true}
	} else {
		assignedTo = pgtype.Int8{Valid: false}
	}
	// Map Domain -> SQLC Params
	res, err := r.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ProjectID:   t.ProjectID,
		Title:       t.Title,
		Description: pgtype.Text{String: t.Description, Valid: t.Description != ""},
		Status:      t.Status,
		AssignedTo:  assignedTo,
	})
	if err != nil {
		return nil, err
	}

	return mapSQLCTaskToDomain(res), nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, t *tasks.Task) (*tasks.Task, error) {
	var assignedTo pgtype.Int8
	if t.AssignedTo != nil {
		assignedTo = pgtype.Int8{Int64: *t.AssignedTo, Valid: true}
	} else {
		assignedTo = pgtype.Int8{Valid: false}
	}

	res, err := r.queries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ID:          t.ID,
		Title:       t.Title,
		Description: pgtype.Text{String: t.Description, Valid: t.Description != ""},
		Status:      t.Status,
		AssignedTo:  assignedTo,
	})
	if err != nil {
		return nil, err
	}

	return mapSQLCTaskToDomain(res), nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, id int64) error {
	return r.queries.DeleteTask(ctx, id)
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, id int64) (*tasks.Task, error) {
	res, err := r.queries.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapSQLCTaskToDomain(res), nil
}

func (r *TaskRepository) ListTaskByProject(ctx context.Context, projectID int64) ([]*tasks.Task, error) {
	rows, err := r.queries.ListTasksForProject(ctx, projectID)
	if err != nil {
		return nil, err
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
	return r.queries.RecordTaskActivity(ctx, sqlc.RecordTaskActivityParams{
		TaskID:  a.TaskID,
		UserID:  a.UserID,
		Action:  a.Action,
		Details: pgtype.Text{String: a.Details, Valid: a.Details != ""},
	})
}

func (r *TaskRepository) GetTaskHistory(ctx context.Context, taskID int64) ([]*tasks.TaskActivity, error) {
	rows, err := r.queries.GetTaskHistory(ctx, taskID)
	if err != nil {
		return nil, err
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

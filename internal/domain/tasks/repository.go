package tasks

import (
	"context"
	"time"
)

// Task represents the domain model.
type Task struct {
	ID          int64
	ProjectID   int64
	Title       string
	Description string
	Status      string
	AssignedTo  *int64 // Pointer because it could be unassigned
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaskActivity represents a single history log entry.
type TaskActivity struct {
	ID        int64
	TaskID    int64
	UserID    int64
	UserEmail string
	Action    string
	Details   string
	CreatedAt time.Time
}

type Repository interface {
	CreateTask(ctx context.Context, task *Task) (*Task, error)
	UpdateTask(ctx context.Context, task *Task) (*Task, error)
	GetTaskByID(ctx context.Context, id int64) (*Task, error)
	ListByProject(ctx context.Context, projectID int64) ([]*Task, error)

	// History methods
	RecordActivity(ctx context.Context, activity *TaskActivity) error
	GetHistory(ctx context.Context, taskID int64) ([]*TaskActivity, error)
}

package tasks

import (
	"context"
	"time"
)

// Task represents the domain model.
type Task struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	AssignedTo  *int64    `json:"assigned_to"` // Pointer because it could be unassigned
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskActivity represents a single history log entry.
type TaskActivity struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	CreateTask(ctx context.Context, task *Task) (*Task, error)
	UpdateTask(ctx context.Context, task *Task) (*Task, error)
	DeleteTask(ctx context.Context, id int64) error
	GetTaskByID(ctx context.Context, id int64) (*Task, error)
	ListTaskByProject(ctx context.Context, projectID int64) ([]*Task, error)

	// History methods
	RecordTaskActivity(ctx context.Context, activity *TaskActivity) error
	GetTaskHistory(ctx context.Context, taskID int64) ([]*TaskActivity, error)
}

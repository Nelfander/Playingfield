package tasks

import (
	"context"
	"time"
)

type FakeRepository struct {
	tasks      []*Task
	activities []*TaskActivity
	nextID     int64
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		tasks:      []*Task{},
		activities: []*TaskActivity{},
		nextID:     1,
	}
}

func (f *FakeRepository) CreateTask(ctx context.Context, task *Task) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	task.ID = f.nextID
	f.nextID++
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	f.tasks = append(f.tasks, task)
	return task, nil
}

func (f *FakeRepository) UpdateTask(ctx context.Context, task *Task) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for i, t := range f.tasks {
		if t.ID == task.ID {
			task.UpdatedAt = time.Now()
			f.tasks[i] = task
			return f.tasks[i], nil
		}
	}
	return nil, ErrTaskNotFound
}

func (f *FakeRepository) GetTaskByID(ctx context.Context, id int64) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, t := range f.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, ErrTaskNotFound
}

func (f *FakeRepository) DeleteTask(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for i, t := range f.tasks {
		if t.ID == id {
			f.tasks = append(f.tasks[:i], f.tasks[i+1:]...)
			return nil
		}
	}
	return ErrTaskNotFound
}

func (f *FakeRepository) ListTaskByProject(ctx context.Context, projectID int64) ([]*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var res []*Task
	for _, t := range f.tasks {
		if t.ProjectID == projectID {
			res = append(res, t)
		}
	}
	return res, nil
}

// History Methods
func (f *FakeRepository) RecordTaskActivity(ctx context.Context, activity *TaskActivity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	activity.ID = int64(len(f.activities) + 1)
	activity.CreatedAt = time.Now()
	f.activities = append(f.activities, activity)
	return nil
}

func (f *FakeRepository) GetTaskHistory(ctx context.Context, taskID int64) ([]*TaskActivity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var res []*TaskActivity
	for _, a := range f.activities {
		if a.TaskID == taskID {
			res = append(res, a)
		}
	}
	return res, nil
}

package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type FakeRepository struct {
	tasks       []*Task
	activities  []*TaskActivity
	attachments []*TaskAttachment
	nextID      int64
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		tasks:       []*Task{},
		activities:  []*TaskActivity{},
		attachments: []*TaskAttachment{},
		nextID:      1,
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

func (f *FakeRepository) CreateAttachment(ctx context.Context, att *TaskAttachment, fileKey string) (*TaskAttachment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	att.ID = int64(len(f.attachments) + 1)
	att.CreatedAt = time.Now()
	f.attachments = append(f.attachments, att)
	return att, nil
}

func (f *FakeRepository) GetTaskAttachments(ctx context.Context, taskID int64) ([]*TaskAttachment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var res []*TaskAttachment
	for _, a := range f.attachments {
		if a.TaskID == taskID {
			res = append(res, a)
		}
	}
	return res, nil
}

func (f *FakeRepository) GetAttachmentByID(ctx context.Context, id int64) (*TaskAttachment, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	for _, a := range f.attachments {
		if a.ID == id {
			// in a real DB we store the key, in the fake we'll just
			// pretend the filename or a hardcoded string is the key.
			fakeKey := fmt.Sprintf("keys/%s", a.FileName)
			return a, fakeKey, nil
		}
	}
	return nil, "", errors.New("attachment not found")
}

func (f *FakeRepository) DeleteAttachment(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for i, a := range f.attachments {
		if a.ID == id {
			f.attachments = append(f.attachments[:i], f.attachments[i+1:]...)
			return nil
		}
	}
	return errors.New("attachment not found")
}

package tasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/nelfander/Playingfield/internal/domain/projects"

	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
)

var (
	ErrUnauthorized = errors.New("unauthorized: you do not have permission for this action")
	ErrTaskNotFound = errors.New("task not found")
)

type Service struct {
	repo        Repository
	projectRepo projects.Repository
	hub         *ws.Hub
}

func NewService(repo Repository, projectRepo projects.Repository, hub *ws.Hub) *Service {
	return &Service{
		repo:        repo,
		projectRepo: projectRepo,
		hub:         hub,
	}
}

func (s *Service) CreateTask(ctx context.Context, requesterID int64, t Task) (*Task, error) {
	//  Fetch the project to verify ownership.
	project, err := s.projectRepo.GetByID(ctx, t.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}

	// Security Check.
	if project.OwnerID != requesterID {
		return nil, ErrUnauthorized
	}

	// Save the task.
	createdTask, err := s.repo.CreateTask(ctx, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	//  Record Activity (STRICT: fail if this fails).
	activity := &TaskActivity{
		TaskID:  createdTask.ID,
		UserID:  requesterID,
		Action:  "CREATED",
		Details: fmt.Sprintf("Task created and assigned to user %d", *createdTask.AssignedTo),
	}
	err = s.repo.RecordActivity(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("task created but history log failed: %w", err)
	}

	//  Broadcast.
	if s.hub != nil {
		notification := fmt.Sprintf("TASK_CREATED:%d", t.ProjectID)
		s.hub.Broadcast <- []byte(notification)
	}

	return createdTask, nil
}

func (s *Service) UpdateTask(ctx context.Context, requesterID int64, t Task) (*Task, error) {
	// Fetch the existing task to see who is assigned and which project it belongs to.
	existingTask, err := s.repo.GetTaskByID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// Fetch the project to see who the owner is.
	project, err := s.projectRepo.GetByID(ctx, existingTask.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %w", err)
	}

	// Authorization Check: Is requester the Owner OR the Assignee?
	isOwner := project.OwnerID == requesterID
	isAssignee := existingTask.AssignedTo != nil && *existingTask.AssignedTo == requesterID

	if !isOwner && !isAssignee {
		return nil, fmt.Errorf("unauthorized: you are not the owner or the assigned member")
	}

	// Perform the update.
	updatedTask, err := s.repo.UpdateTask(ctx, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Record Activity (Log what happened).
	activity := &TaskActivity{
		TaskID:  updatedTask.ID,
		UserID:  requesterID,
		Action:  "UPDATED",
		Details: fmt.Sprintf("Task updated by user %d. New Status: %s", requesterID, updatedTask.Status),
	}
	_ = s.repo.RecordActivity(ctx, activity)

	if s.hub != nil {
		notification := fmt.Sprintf("TASK_UPDATED:%d:%d", updatedTask.ProjectID, updatedTask.ID)
		s.hub.Broadcast <- []byte(notification)
	}

	return updatedTask, nil
}

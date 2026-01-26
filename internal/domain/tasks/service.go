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

	// task can have no "assignedTo" in the moment of their creation,
	// a project owner might make the project and create the tasks needed but decide
	// to assign them to project members later on
	details := "Initial task creation"
	if t.AssignedTo != nil {
		details = "Task created and assigned to team member"
	}
	//  Record Activity (STRICT: fail if this fails).
	activity := &TaskActivity{
		TaskID:  createdTask.ID,
		UserID:  requesterID,
		Action:  "CREATED",
		Details: details,
	}
	err = s.repo.RecordTaskActivity(ctx, activity)
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

func (s *Service) UpdateTask(ctx context.Context, requesterID int64, t Task, commitMsg string) (*Task, error) {
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
		Details: fmt.Sprintf("[%s] %s", updatedTask.Status, commitMsg),
	}
	err = s.repo.RecordTaskActivity(ctx, activity)
	if err != nil {
		return nil, fmt.Errorf("task updated but history log failed: %w", err)
	}

	if s.hub != nil {
		notification := fmt.Sprintf("TASK_UPDATED:%d:%d", updatedTask.ProjectID, updatedTask.ID)
		s.hub.Broadcast <- []byte(notification)
	}

	return updatedTask, nil
}

func (s *Service) DeleteTask(ctx context.Context, requesterID int64, taskID int64) error {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	//  Fetch the project to verify ownership
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to verify project: %w", err)
	}
	// Security Check: Only the project owner can delete tasks
	if project.OwnerID != requesterID {
		return fmt.Errorf("unauthorized: only the project owner can delete tasks")
	}
	// Delete the task
	err = s.repo.DeleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	// Broadcast deletion
	if s.hub != nil {
		notification := fmt.Sprintf("TASK_DELETED:%d:%d", task.ProjectID, taskID)
		s.hub.Broadcast <- []byte(notification)
	}

	return nil
}

func (s *Service) GetTaskHistory(ctx context.Context, requesterID int64, taskID int64) ([]*TaskActivity, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// 2. Authorization Check: Is the requester a member of this project?
	// We use the projectRepo's membership check logic here
	members, err := s.projectRepo.ListUsersInProject(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project membership: %w", err)
	}

	isMember := false
	for _, m := range members {
		if m.ID == requesterID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, fmt.Errorf("unauthorized: you must be a project member to view history")
	}

	// Fetch and return history
	return s.repo.GetTaskHistory(ctx, taskID)
}

// ListTasks returns all tasks for a project, but only if the requester is a member.
func (s *Service) ListTasks(ctx context.Context, requesterID int64, projectID int64) ([]*Task, error) {
	// Authorization: Is the user in this project?
	members, err := s.projectRepo.ListUsersInProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("could not verify project membership: %w", err)
	}

	isMember := false
	for _, m := range members {
		if m.ID == requesterID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, fmt.Errorf("unauthorized: you are not a member of this project")
	}

	//  Fetch the tasks
	return s.repo.ListTaskByProject(ctx, projectID)
}

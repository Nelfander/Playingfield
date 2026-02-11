package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/nelfander/Playingfield/internal/domain"
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
	storage     domain.StorageProvider
	hub         *ws.Hub
}

func NewService(repo Repository, projectRepo projects.Repository, storage domain.StorageProvider, hub *ws.Hub) *Service {
	return &Service{
		repo:        repo,
		projectRepo: projectRepo,
		storage:     storage,
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
		slog.Warn("unauthorized task creation attempt", "project_id", t.ProjectID, "requester_id", requesterID)
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
	if err := s.repo.RecordTaskActivity(ctx, activity); err != nil {
		// no slog.Error here because the Repo already has it!
		return nil, fmt.Errorf("task created but history log failed: %w", err)
	}

	slog.Info("task created", "task_id", createdTask.ID, "project_id", t.ProjectID)

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

	// Authorization check: is requester the owner OR the assignee?
	isOwner := project.OwnerID == requesterID
	isAssignee := existingTask.AssignedTo != nil && *existingTask.AssignedTo == requesterID

	if !isOwner && !isAssignee {
		slog.Warn("unauthorized task update attempt", "task_id", t.ID, "requester_id", requesterID)
		return nil, ErrUnauthorized
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
	if err := s.repo.RecordTaskActivity(ctx, activity); err != nil {
		// slogged again on repo side
		return nil, fmt.Errorf("task updated but history log failed: %w", err)
	}

	slog.Info("task updated", "task_id", updatedTask.ID, "status", updatedTask.Status)

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
		slog.Warn("unauthorized task deletion attempt", "task_id", taskID, "requester_id", requesterID)
		return fmt.Errorf("unauthorized: only the project owner can delete tasks")
	}
	// Delete the task
	err = s.repo.DeleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	slog.Info("task deleted", "task_id", taskID, "project_id", task.ProjectID)

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
		return nil, ErrUnauthorized
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
		return nil, ErrUnauthorized
	}

	//  Fetch the tasks
	return s.repo.ListTaskByProject(ctx, projectID)
}

func (s *Service) UploadAttachment(ctx context.Context, requesterID int64, taskID int64, fileName string, fileSize int64, content io.Reader) (*TaskAttachment, error) {
	// fetch the task
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// fetch the project to verify ownership
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify project ownership: %w", err)
	}

	// ss requester the owner OR the assignee?
	isOwner := project.OwnerID == requesterID
	isAssignee := task.AssignedTo != nil && *task.AssignedTo == requesterID

	if !isOwner && !isAssignee {
		slog.Warn("unauthorized attachment upload attempt",
			"task_id", taskID,
			"requester_id", requesterID)
		return nil, ErrUnauthorized
	}

	// this streams content without loading the whole file into RAM
	uploadResult, err := s.storage.UploadFile(ctx, fileName, content)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// save metadata to DB
	attachment := &TaskAttachment{
		TaskID:   taskID,
		UserID:   requesterID,
		FileName: fileName,
		FileSize: fileSize,
		FileUrl:  uploadResult.URL,
	}

	// This repo method also records the "UPLOADED" activity log
	created, err := s.repo.CreateAttachment(ctx, attachment, uploadResult.Key)
	if err != nil {
		return nil, err
	}

	slog.Info("file attached to task",
		"task_id", taskID,
		"file_name", fileName,
		"size_mb", fmt.Sprintf("%.2f MB", float64(fileSize)/(1024*1024)),
	)

	return created, nil
}

func (s *Service) DeleteAttachment(ctx context.Context, requesterID int64, attachmentID int64) error {
	// get attachment info
	att, fileKey, err := s.repo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("attachment not found: %w", err)
	}

	// fetch the task to see who is assigned
	task, err := s.repo.GetTaskByID(ctx, att.TaskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// fetch the project to see who the owner is
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to verify project: %w", err)
	}

	//  Owner OR Assignee can delete
	isOwner := project.OwnerID == requesterID
	isAssignee := task.AssignedTo != nil && *task.AssignedTo == requesterID

	if !isOwner && !isAssignee {
		slog.Warn("unauthorized attachment deletion attempt",
			"attachment_id", attachmentID,
			"requester_id", requesterID,
			"task_id", task.ID)
		return ErrUnauthorized
	}

	// remove from MinIO
	if err := s.storage.DeleteFile(ctx, fileKey); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// remove from DB
	if err := s.repo.DeleteAttachment(ctx, attachmentID); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// record Activity
	_ = s.repo.RecordTaskActivity(ctx, &TaskActivity{
		TaskID:  att.TaskID,
		UserID:  requesterID,
		Action:  "DELETED_FILE",
		Details: fmt.Sprintf("Deleted file: %s", att.FileName),
	})

	slog.Info("attachment permanently deleted",
		"attachment_id", attachmentID,
		"task_id", att.TaskID,
		"file_name", att.FileName,
		"deleted_by", requesterID,
	)

	return nil
}

func (s *Service) GetTaskAttachments(ctx context.Context, requesterID int64, taskID int64) ([]*TaskAttachment, error) {
	// fetch task to find the Project ID
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// is user a member of this project?
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
		return nil, ErrUnauthorized
	}

	return s.repo.GetTaskAttachments(ctx, taskID)
}

func (s *Service) GetAttachmentStream(ctx context.Context, requesterID int64, attachmentID int64) (io.ReadCloser, string, error) {
	// get attachment metadata and the secret fileKey (S3 Key) from the repo
	att, fileKey, err := s.repo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return nil, "", fmt.Errorf("attachment not found: %w", err)
	}

	task, err := s.repo.GetTaskByID(ctx, att.TaskID)
	if err != nil {
		return nil, "", fmt.Errorf("task associated with attachment not found: %w", err)
	}

	members, err := s.projectRepo.ListUsersInProject(ctx, task.ProjectID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to verify project membership: %w", err)
	}

	isMember := false
	for _, m := range members {
		if m.ID == requesterID {
			isMember = true
			break
		}
	}

	if !isMember {
		slog.Warn("unauthorized attachment download attempt",
			"attachment_id", attachmentID,
			"requester_id", requesterID,
			"project_id", task.ProjectID)
		return nil, "", ErrUnauthorized
	}

	// 4. Use the StorageProvider to get the stream
	// Note: You may need to add 'DownloadFile' to your domain.StorageProvider interface
	// if it only has Upload/Delete currently.
	content, contentType, err := s.storage.DownloadFile(ctx, fileKey)
	if err != nil {
		// slog is already on s.storage.DownloadFile
		return nil, "", fmt.Errorf("failed to retrieve file from storage: %w", err)
	}

	return content, contentType, nil
}

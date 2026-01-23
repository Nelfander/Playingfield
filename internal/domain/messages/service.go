package messages

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
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

type ChatService interface {
	SendProjectMessage(ctx context.Context, senderID int64, projectID int64, content string) (*Message, error)
	GetProjectHistory(ctx context.Context, projectID int64) ([]Message, error)
	SendDirectMessage(ctx context.Context, senderID, receiverID int64, content string) (*Message, error)
	GetDMHistory(ctx context.Context, userA, userB int64) ([]Message, error)
}

func (s *Service) SendProjectMessage(ctx context.Context, senderID int64, projectID int64, content string) (*Message, error) {
	// get Project details via the Project Interface
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// get the member list via the Project Interface
	members, err := s.projectRepo.ListUsersInProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch project members: %w", err)
	}

	// check if sender is in the member list
	// ListUsersInProject includes the owner!
	isAuthorized := false
	for _, m := range members {
		if m.ID == senderID {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return nil, fmt.Errorf("unauthorized: user %d is not a member of project %s", senderID, project.Name)
	}

	// prepare the Message domain object
	msg := Message{
		SenderID:  senderID,
		ProjectID: &projectID,
		Content:   content,
	}

	saved, err := s.repo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}

	// broadcast via websocket hub
	broadcastData := map[string]interface{}{
		"type": "new_project_message",
		"data": saved,
	}
	payload, _ := json.Marshal(broadcastData)
	s.hub.BroadcastToProject(projectID, payload)

	return saved, nil
}

func (s *Service) GetProjectHistory(ctx context.Context, projectID int64) ([]Message, error) {
	return s.repo.GetByProject(ctx, projectID)
}

// SendDirectMessage checks for shared projects before saving and broadcasting
func (s *Service) SendDirectMessage(ctx context.Context, senderID, receiverID int64, content string) (*Message, error) {
	shared, err := s.projectRepo.UsersShareProject(ctx, senderID, receiverID)
	if err != nil {
		return nil, fmt.Errorf("could not verify connection: %w", err)
	}
	if !shared {
		return nil, fmt.Errorf("you can only message users who share a project with you")
	}
	msg := Message{
		SenderID:   senderID,
		ReceiverID: &receiverID,
		Content:    content,
	}
	saved, err := s.repo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}
	broadcastData := map[string]interface{}{
		"type": "new_direct_message",
		"data": saved,
	}
	payload, _ := json.Marshal(broadcastData)

	s.hub.SendToUser(receiverID, payload)
	s.hub.SendToUser(senderID, payload)

	return saved, nil
}

// GetDMHistory fetches private conversation
func (s *Service) GetDMHistory(ctx context.Context, userA, userB int64) ([]Message, error) {
	return s.repo.GetDirectMessages(ctx, userA, userB)
}

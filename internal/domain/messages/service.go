package messages

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
)

type Service struct {
	repo  Repository
	store *sqlc.Queries
	hub   *ws.Hub
}

func NewService(repo Repository, store *sqlc.Queries, hub *ws.Hub) *Service {
	return &Service{
		repo:  repo,
		store: store,
		hub:   hub,
	}
}

func (s *Service) SendProjectMessage(ctx context.Context, senderID int64, projectID int64, content string) (*Message, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or db error: %w", err)
	}

	members, err := s.store.ListUsersInProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, m := range members {
		if m.ID == senderID {
			isMember = true
		}
	}

	isOwner := project.OwnerID == senderID
	if !isMember && !isOwner {
		return nil, fmt.Errorf("unauthorized: you are not a member or owner of this project")
	}

	msg := Message{
		SenderID:  senderID,
		ProjectID: &projectID,
		Content:   content,
	}

	saved, err := s.repo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}

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
	shared, err := s.store.CheckSharedProject(ctx, sqlc.CheckSharedProjectParams{
		SenderID:   senderID,
		ReceiverID: receiverID,
	})

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

	// Send to both so all their open devices/tabs sync instantly
	s.hub.SendToUser(receiverID, payload)
	s.hub.SendToUser(senderID, payload)

	return saved, nil
}

// GetDMHistory fetches private conversation
func (s *Service) GetDMHistory(ctx context.Context, userA, userB int64) ([]Message, error) {
	return s.repo.GetDirectMessages(ctx, userA, userB)
}

package messages

import (
	"context"
	"time"
)

type FakeRepository struct {
	messages []Message
	nextID   int64
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		messages: []Message{},
		nextID:   1,
	}
}

func (f *FakeRepository) Create(ctx context.Context, m Message) (*Message, error) {
	m.ID = f.nextID
	f.nextID++
	m.CreatedAt = time.Now()
	// In a real DB, the email comes from a join. In fake, we hardcode it for the UI.
	m.SenderEmail = "test@example.com"

	f.messages = append(f.messages, m)
	return &f.messages[len(f.messages)-1], nil
}

func (f *FakeRepository) GetByProject(ctx context.Context, projectID int64) ([]Message, error) {
	var res []Message
	for _, m := range f.messages {
		if m.ProjectID != nil && *m.ProjectID == projectID {
			res = append(res, m)
		}
	}
	return res, nil
}

func (f *FakeRepository) GetDirectMessages(ctx context.Context, userA, userB int64) ([]Message, error) {
	var res []Message
	for _, m := range f.messages {
		if m.ReceiverID == nil {
			continue
		}
		// Conversation is bi-directional
		isAtoB := m.SenderID == userA && *m.ReceiverID == userB
		isBtoA := m.SenderID == userB && *m.ReceiverID == userA
		if isAtoB || isBtoA {
			res = append(res, m)
		}
	}
	return res, nil
}

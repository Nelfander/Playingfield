package projects

import (
	"context"
	"errors"
	"testing"

	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
	"github.com/stretchr/testify/assert"
)

func TestProjectService_Create(t *testing.T) {
	repo := NewFakeRepository()
	// We pass nil for the hub to simplify the test
	svc := NewService(repo, nil)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		p, err := svc.CreateProject(ctx, "New Project", "Desc", 1)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", p.Name)
		assert.Equal(t, int64(1), p.OwnerID)

		// Verify owner was automatically added as a member
		members, _ := repo.ListUsersInProject(ctx, p.ID)
		assert.Len(t, members, 1)
		assert.Equal(t, int64(1), members[0].ID)
		assert.Equal(t, "owner", members[0].Role)
	})

	t.Run("duplicate project name for same owner fails", func(t *testing.T) {
		_, _ = svc.CreateProject(ctx, "Unique Name", "Desc", 1)
		p, err := svc.CreateProject(ctx, "Unique Name", "Desc", 1)

		assert.Nil(t, p)
		assert.True(t, errors.Is(err, ErrDuplicateProject))
	})
}

func TestProjectService_Security(t *testing.T) {
	repo := NewFakeRepository()
	svc := NewService(repo, nil)
	ctx := context.Background()

	// Seed a project owned by User 1
	p, _ := svc.CreateProject(ctx, "Owner 1 Project", "Desc", 1)

	t.Run("unauthorized user cannot update project", func(t *testing.T) {
		updated, err := svc.UpdateProject(ctx, 999, p.ID, "Hacked Name", "Hacked Desc")
		assert.Nil(t, updated)
		assert.True(t, errors.Is(err, ErrUnauthorized))
	})

	t.Run("unauthorized user cannot delete project", func(t *testing.T) {
		err := svc.DeleteProject(ctx, p.ID, 999)
		assert.True(t, errors.Is(err, ErrUnauthorized))

		// Verify project still exists in repo
		exists, _ := repo.GetByID(ctx, p.ID)
		assert.NotNil(t, exists)
	})
}

func TestProjectService_Membership(t *testing.T) {
	repo := NewFakeRepository()
	svc := NewService(repo, nil)
	ctx := context.Background()

	ownerID := int64(1)
	otherUserID := int64(2)
	p, _ := svc.CreateProject(ctx, "Team Project", "Desc", ownerID)

	t.Run("owner can add member", func(t *testing.T) {
		err := svc.AddUserToProject(ctx, ownerID, p.ID, otherUserID, "editor")
		assert.NoError(t, err)

		members, _ := repo.ListUsersInProject(ctx, p.ID)
		assert.Len(t, members, 2) // Owner + New Member
	})

	t.Run("cannot add same member twice", func(t *testing.T) {
		err := svc.AddUserToProject(ctx, ownerID, p.ID, otherUserID, "editor")
		assert.True(t, errors.Is(err, ErrAlreadyMember))
	})

	t.Run("non-owner cannot add members", func(t *testing.T) {
		err := svc.AddUserToProject(ctx, 999, p.ID, 3, "viewer")
		assert.True(t, errors.Is(err, ErrUnauthorized))
	})

	t.Run("owner can remove member", func(t *testing.T) {
		err := svc.RemoveUserFromProject(ctx, ownerID, p.ID, otherUserID)
		assert.NoError(t, err)

		members, _ := repo.ListUsersInProject(ctx, p.ID)
		assert.Len(t, members, 1) // Only owner remains
	})
}

func TestProjectService_Notifications(t *testing.T) {
	repo := NewFakeRepository()

	// create a Hub but MANUALLY initialize the channel with a buffer of 1
	// This allows the service to send the message without blocking.
	hub := &ws.Hub{
		Broadcast: make(chan []byte, 1),
	}

	svc := NewService(repo, hub)
	ctx := context.Background()

	t.Run("creating project sends notification to hub", func(t *testing.T) {
		_, err := svc.CreateProject(ctx, "Notif Project", "Desc", 1)
		assert.NoError(t, err)

		// safely read it
		select {
		case msg := <-hub.Broadcast:
			assert.Equal(t, "PROJECT_CREATED", string(msg))
		default:
			t.Fatal("Expected broadcast message but none received")
		}
	})
}

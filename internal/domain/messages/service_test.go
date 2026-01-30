package messages

import (
	"context"
	"testing"

	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
	"github.com/stretchr/testify/assert"
)

func TestMessageService(t *testing.T) {
	ctx := context.Background()
	testHub := ws.NewHub()
	go testHub.Run()

	t.Run("Project Messaging - Authorization", func(t *testing.T) {
		msgRepo := NewFakeRepository()
		projRepo := projects.NewFakeRepository()
		svc := NewService(msgRepo, projRepo, testHub)

		p, _ := projRepo.CreateProject(ctx, projects.Project{Name: "Backend Team", OwnerID: 1})
		projRepo.AddUserToProject(ctx, p.ID, 1, "owner")

		// Success: Owner sends message
		res, err := svc.SendProjectMessage(ctx, 1, p.ID, "Hello")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Failure: Stranger sends message
		// We use "_" because we don't need to store the values, just assert them
		_, errFail := svc.SendProjectMessage(ctx, 99, p.ID, "Intruder")
		assert.Error(t, errFail)
		assert.Contains(t, errFail.Error(), "unauthorized")
	})

	t.Run("Direct Messaging - Shared Project Logic", func(t *testing.T) {
		msgRepo := NewFakeRepository()
		projRepo := projects.NewFakeRepository()
		svc := NewService(msgRepo, projRepo, testHub)

		userA := int64(10)
		userB := int64(20)
		stranger := int64(30)

		//  Setup: UserA and UserB share a project, Stranger is alone
		p, _ := projRepo.CreateProject(ctx, projects.Project{Name: "Shared", OwnerID: userA})
		projRepo.AddUserToProject(ctx, p.ID, userA, "owner")
		projRepo.AddUserToProject(ctx, p.ID, userB, "member")

		//  Success: UserA and UserB can talk
		res, err := svc.SendDirectMessage(ctx, userA, userB, "Hey partner")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		//  Failure: UserA tries to message Stranger (no shared projects)
		resFail, errFail := svc.SendDirectMessage(ctx, userA, stranger, "Hey you")
		assert.Error(t, errFail)
		assert.Nil(t, resFail)
		assert.Contains(t, errFail.Error(), "share a project")
	})
}

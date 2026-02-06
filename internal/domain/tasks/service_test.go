package tasks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
	"github.com/stretchr/testify/assert"
)

func TestTaskService_Create(t *testing.T) {
	taskRepo := NewFakeRepository()
	projRepo := projects.NewFakeRepository()
	svc := NewService(taskRepo, projRepo, nil)
	ctx := context.Background()

	ownerID := int64(1)

	// seed a project via project repo
	p, _ := projRepo.CreateProject(ctx, projects.Project{Name: "Task Project", OwnerID: ownerID})

	t.Run("owner can create task", func(t *testing.T) {
		newTask := Task{
			ProjectID: p.ID,
			Title:     "Write Tests",
			Status:    "todo",
		}

		created, err := svc.CreateTask(ctx, ownerID, newTask)

		assert.NoError(t, err)
		assert.Equal(t, "Write Tests", created.Title)

		// verify history was recorded
		history, _ := taskRepo.GetTaskHistory(ctx, created.ID)
		assert.Len(t, history, 1)
		assert.Equal(t, "CREATED", history[0].Action)
	})

	t.Run("non-owner cannot create task", func(t *testing.T) {
		hackerID := int64(666)
		newTask := Task{ProjectID: p.ID, Title: "Hacked Task"}

		created, err := svc.CreateTask(ctx, hackerID, newTask)

		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestTaskService_UpdatePermissions(t *testing.T) {
	taskRepo := NewFakeRepository()
	projRepo := projects.NewFakeRepository()
	svc := NewService(taskRepo, projRepo, nil)
	ctx := context.Background()

	ownerID := int64(1)
	assigneeID := int64(2)
	otherMemberID := int64(3)

	// setup: project and a task assigned to user 2
	p, _ := projRepo.CreateProject(ctx, projects.Project{Name: "Collab", OwnerID: ownerID})

	// use the service to create the task so it's tracked in the fake repo
	tInitial := Task{ProjectID: p.ID, Title: "Fix Bug", AssignedTo: &assigneeID, Status: "todo"}
	task, _ := svc.CreateTask(ctx, ownerID, tInitial)

	t.Run("assignee can update task status", func(t *testing.T) {
		task.Status = "in_progress"
		updated, err := svc.UpdateTask(ctx, assigneeID, *task, "Starting work")

		assert.NoError(t, err)
		assert.Equal(t, "in_progress", updated.Status)

		// verify history shows the assignee made the change
		history, _ := taskRepo.GetTaskHistory(ctx, task.ID)
		// 1 from create, 1 from this update
		assert.Len(t, history, 2)
		assert.Equal(t, assigneeID, history[1].UserID)
	})

	t.Run("non-assignee member cannot update task", func(t *testing.T) {
		task.Status = "done"
		updated, err := svc.UpdateTask(ctx, otherMemberID, *task, "I'm helping!")

		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("owner can update task even if not assigned", func(t *testing.T) {
		task.Title = "Updated by Boss"
		updated, err := svc.UpdateTask(ctx, ownerID, *task, "Overriding title")

		assert.NoError(t, err)
		assert.Equal(t, "Updated by Boss", updated.Title)
	})
}

func TestTaskService_HistoryPermissions(t *testing.T) {
	taskRepo := NewFakeRepository()
	projRepo := projects.NewFakeRepository()
	svc := NewService(taskRepo, projRepo, nil)
	ctx := context.Background()

	ownerID := int64(1)
	memberID := int64(2)
	strangerID := int64(999)

	// setup project & member
	p, _ := projRepo.CreateProject(ctx, projects.Project{Name: "History Test", OwnerID: ownerID})
	_ = projRepo.AddUserToProject(ctx, p.ID, memberID, "editor")

	// create task
	task, _ := svc.CreateTask(ctx, ownerID, Task{ProjectID: p.ID, Title: "Log Me"})

	t.Run("member can view history", func(t *testing.T) {
		history, err := svc.GetTaskHistory(ctx, memberID, task.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, history)
	})

	t.Run("stranger cannot view history", func(t *testing.T) {
		history, err := svc.GetTaskHistory(ctx, strangerID, task.ID)
		assert.Error(t, err)
		assert.Nil(t, history)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestTaskService_WebSocketBroadcast(t *testing.T) {
	// setup Hub
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// setup Service
	taskRepo := NewFakeRepository()
	projRepo := projects.NewFakeRepository()
	svc := NewService(taskRepo, projRepo, hub)

	ctx := context.Background()
	ownerID := int64(100)
	p, _ := projRepo.CreateProject(ctx, projects.Project{ID: 1, Name: "Live Proj", OwnerID: ownerID})

	//  Register a Fake Client
	client := &ws.Client{
		UserID: ownerID,
		Send:   make(chan []byte, 10),
	}
	hub.Register <- client

	// wait for registration to process
	time.Sleep(10 * time.Millisecond)

	t.Run("creating a task triggers a global broadcast", func(t *testing.T) {
		task := Task{ProjectID: p.ID, Title: "Realtime Task"}
		_, err := svc.CreateTask(ctx, ownerID, task)
		assert.NoError(t, err)

		select {
		case msg := <-client.Send:
			expected := fmt.Sprintf("TASK_CREATED:%d", p.ID)
			assert.Equal(t, expected, string(msg))
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for WebSocket broadcast")
		}
	})
}

func TestTaskService_WebSocketIntegration(t *testing.T) {
	// setup Hub and Service
	hub := ws.NewHub()
	go hub.Run() // Start the post office

	taskRepo := NewFakeRepository()
	projRepo := projects.NewFakeRepository()
	svc := NewService(taskRepo, projRepo, hub) // inject the REAL hub

	ctx := context.Background()
	ownerID := int64(100)
	p, _ := projRepo.CreateProject(ctx, projects.Project{ID: 1, Name: "Live Proj", OwnerID: ownerID})

	//  Simulate a "Connected Client"
	// We create a client with a buffered channel so we can read from it
	client := &ws.Client{
		UserID: ownerID, // so the Hub knows who it is
		Send:   make(chan []byte, 1),
	}
	hub.Register <- client

	time.Sleep(10 * time.Millisecond)

	t.Run("creating a task sends a real-time signal to connected clients", func(t *testing.T) {
		// Trigger the action
		_, err := svc.CreateTask(ctx, ownerID, Task{ProjectID: p.ID, Title: "Signal Task"})
		assert.NoError(t, err)

		// Verify the Client received the message
		select {
		case msg := <-client.Send:
			expected := fmt.Sprintf("TASK_CREATED:%d", p.ID)
			assert.Equal(t, expected, string(msg))
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Client never received the WebSocket notification!")
		}
	})
}

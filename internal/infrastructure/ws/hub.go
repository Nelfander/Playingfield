package ws

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single connected user
type Client struct {
	UserID    int64
	ProjectID int64
	Conn      *websocket.Conn
	Send      chan []byte // A channel to send messages to this specific user
}

// Hub maintains the set of active clients
type Hub struct {
	clients      map[int64]*Client
	ProjectRooms map[int64]map[*Client]bool
	Broadcast    chan []byte
	Register     chan *Client
	Unregister   chan *Client
	mu           sync.RWMutex // mutex to protect the clients map
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:    make(chan []byte),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		clients:      make(map[int64]*Client),
		ProjectRooms: make(map[int64]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client

			//  Add client to their specific project room
			if client.ProjectID != 0 {
				if h.ProjectRooms[client.ProjectID] == nil {
					h.ProjectRooms[client.ProjectID] = make(map[*Client]bool)
				}
				h.ProjectRooms[client.ProjectID][client] = true
				fmt.Printf("✅ Chat Room: User %d joined Project %d\n", client.UserID, client.ProjectID)
			} else {
				fmt.Printf("ℹ️ Global Hub: User %d connected\n", client.UserID)
			}

			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				//  Remove from project room
				if client.ProjectID != 0 && h.ProjectRooms[client.ProjectID] != nil {
					delete(h.ProjectRooms[client.ProjectID], client)
					// Clean up empty rooms
					if len(h.ProjectRooms[client.ProjectID]) == 0 {
						delete(h.ProjectRooms, client.ProjectID)
					}
				}
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				client.Send <- message
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client, ok := h.clients[userID]; ok {
		select {
		case client.Send <- message:
		default:
		}
	}
}

func (h *Hub) SendToProjectMembers(userIDs []int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, id := range userIDs {
		if client, ok := h.clients[id]; ok {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

func (h *Hub) BroadcastToProject(projectID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.ProjectRooms[projectID]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				continue
			}
		}
	}
}

package ws

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// client represents a single connected user
type Client struct {
	UserID    int64
	ProjectID int64
	Conn      *websocket.Conn
	Send      chan []byte   // a channel to send messages to this specific user
	done      chan struct{} // internal shutdown signal
}

// Hub maintains the set of active clients and project rooms
type Hub struct {
	// Map of UserID to a set of active connections (allows multi-tab)
	clients      map[int64]map[*Client]bool
	ProjectRooms map[int64]map[*Client]bool
	Broadcast    chan []byte
	Register     chan *Client
	Unregister   chan *Client
	mu           sync.RWMutex  // mutex to protect the clients map
	stop         chan struct{} // empty struct 0 bytes(thx anthony ^_^)
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:    make(chan []byte),
		Register:     make(chan *Client, 128),
		Unregister:   make(chan *Client, 128),
		clients:      make(map[int64]map[*Client]bool),
		ProjectRooms: make(map[int64]map[*Client]bool),
		stop:         make(chan struct{}),
	}
}

func NewClient(userID, projectID int64, conn *websocket.Conn) *Client {
	return &Client{
		UserID:    userID,
		ProjectID: projectID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		done:      make(chan struct{}), // initialized internally
	}
}

// getter for safe external access
func (c *Client) DoneChan() <-chan struct{} {
	return c.done
}

func (h *Hub) Stop() {
	close(h.stop) // this broadcasts to the Hub's loop to stop
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, connections := range h.clients {
		for client := range connections {
			// close the Send channel so the client's writePump stops
			if client.Send != nil {
				close(client.Send)
			}

			// close the actual WebSocket connection
			if client.Conn != nil {
				client.Conn.Close()
			}

			// safe close for internal done signal
			if client.done != nil {
				select {
				case <-client.done:
				default:
					close(client.done)
				}
			}
		}
		delete(h.clients, userID)
	}

	// clear the rooms map too
	h.ProjectRooms = make(map[int64]map[*Client]bool)
	slog.Info("hub cleanup complete", "status", "all connections closed")
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			// initialize the user's connection map if it doesn't exist
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true

			// add specific connection to project room
			if client.ProjectID != 0 {
				if h.ProjectRooms[client.ProjectID] == nil {
					h.ProjectRooms[client.ProjectID] = make(map[*Client]bool)
				}
				h.ProjectRooms[client.ProjectID][client] = true
			}

			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			// remove specific connection from user's map
			if connections, ok := h.clients[client.UserID]; ok {
				delete(connections, client)
				if len(connections) == 0 {
					delete(h.clients, client.UserID)
				}
			}

			// remove specific connection from project room
			if client.ProjectID != 0 {
				if room, ok := h.ProjectRooms[client.ProjectID]; ok {
					delete(room, client)
					if len(room) == 0 {
						delete(h.ProjectRooms, client.ProjectID)
					}
				}
			}

			if client.Conn != nil {
				client.Conn.SetWriteDeadline(time.Now()) // Immediate timeout
				client.Conn.SetReadDeadline(time.Now())  // Immediate timeout
				client.Conn.Close()
			}

			// signal this specific connection to stop
			// check if already closed to avoid panics
			select {
			case <-client.done:
			default:
				close(client.done)
				close(client.Send) // closing this wakes up the WritePump from the channel block
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, connections := range h.clients {
				for client := range connections {
					select {
					case client.Send <- message:
					default: // skip slow clients
					}
				}
			}
			h.mu.RUnlock()

		case <-h.stop:
			slog.Info("hub stopping", "action", "closing all client connections")
			h.cleanup() // A helper function to kick everyone out politely
			return      // Exit the loop and the goroutine

		}
	}
}

// SendToUser sends to ALL open tabs/devices for that user
func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if connections, ok := h.clients[userID]; ok {
		for client := range connections {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

// SendToProjectMembers ensures all connections for multiple users get the message
func (h *Hub) SendToProjectMembers(userIDs []int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, id := range userIDs {
		if connections, ok := h.clients[id]; ok {
			for client := range connections {
				select {
				case client.Send <- message:
				default:
				}
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

// BroadcastToProjectExcept sends a message to everyone in a project room EXCEPT a specific userID
// this is for the "user is typing..." broadcast
func (h *Hub) BroadcastToProjectExcept(projectID int64, excludeUserID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.ProjectRooms[projectID]; ok {
		for client := range clients {
			// Skip ALL connections belonging to the excluded user
			if client.UserID == excludeUserID {
				continue
			}
			select {
			case client.Send <- message:
			default:
				// if a client's buffer is full skip them to avoid blocking the hub
				continue
			}
		}
	}
}

package ws

import (
	"fmt"
	"log"
	"sync"

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

// Hub maintains the set of active clients
type Hub struct {
	clients      map[int64]*Client
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
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		clients:      make(map[int64]*Client),
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

// getter for safe external access,( actually not needed now since i implemented the newclient constructor )
func (c *Client) DoneChan() <-chan struct{} {
	return c.done
}

func (h *Hub) Stop() {
	close(h.stop) // this broadcasts to the Hub's loop to stop
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, client := range h.clients {
		// close the Send channel so the client's writePump stops
		close(client.Send)

		// close the actual WebSocket connection
		client.Conn.Close()

		// signal the client's internal handlers to stop
		// and check if it's already closed to avoid a panic
		select {
		case <-client.done:
		default:
			close(client.done)
		}

		delete(h.clients, client.UserID)
	}

	// clear the rooms map too
	h.ProjectRooms = make(map[int64]map[*Client]bool)
	log.Println("✅ Hub cleanup complete: all connections closed.")
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client

			//  add client to their specific project room
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
				// remove from project room
				if client.ProjectID != 0 {
					if room, ok := h.ProjectRooms[client.ProjectID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(h.ProjectRooms, client.ProjectID)
						}
					}
				}
				delete(h.clients, client.UserID)
				// signal client to shut down
				close(client.done)
			}

			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// drop message or mark client as slow
				}
			}
			h.mu.RUnlock()

		case <-h.stop:
			log.Println("Hub stopping: closing all client connections")
			h.cleanup() // A helper function to kick everyone out politely
			return      // Exit the loop and the goroutine

			/* old unsafe version i leave it here to compare
			case message := <-h.Broadcast:
				h.mu.RLock()
				for _, client := range h.clients {
					client.Send <- message
				}
				h.mu.RUnlock() */
		}
	}
}

func (h *Hub) SendToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client, ok := h.clients[userID]; ok {
		select {
		case client.Send <- message:
		default: // advoids blocking
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

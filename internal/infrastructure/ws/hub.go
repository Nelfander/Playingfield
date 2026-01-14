package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single connected user
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte // A channel to send messages to this specific user
}

// Hub maintains the set of active clients
type Hub struct {
	// Registered clients: Key is UserID
	clients map[int64]*Client

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	mu sync.RWMutex // To protect the clients map
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[int64]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			//println("Hub: Registered user", client.UserID)

			// Direct test: Send only to THIS client immediately
			// This confirms the "Mouth" (Writer Loop) in the handler is working
			client.Send <- []byte("System: You have successfully joined the Hub!")

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
				//println("Hub: Unregistered user", client.UserID)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			// For now, this sends to EVERYONE.
			// Later,  will make "SendToUser" and "SendToProject" methods.
			h.mu.RLock()
			for _, client := range h.clients {
				client.Send <- message
			}
			h.mu.RUnlock()
		}
	}
}

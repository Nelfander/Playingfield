package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
	"github.com/nelfander/Playingfield/internal/metrics"
)

const (
	writeWait  = 10 * time.Second // time allowed to write a message to the peer.
	pongWait   = 30 * time.Second // time allowed to read the next pong message from the peer.
	pingPeriod = 25 * time.Second // send pings to peer with this period (must be ~10% less than pongWait).
)

type WSHandler struct {
	jwtManager  *auth.JWTManager
	hub         *ws.Hub
	chatService *messages.Service
}

type WSIncomingMessage struct {
	Type       string `json:"type"` // "project_chat", "direct_message", "typing", "read_receipt"
	ProjectID  int64  `json:"project_id"`
	ReceiverID int64  `json:"receiver_id"`
	MessageID  int64  `json:"message_id"`
	Content    string `json:"content"`
	IsTyping   bool   `json:"is_typing"`
}

func NewWSHandler(jwtManager *auth.JWTManager, hub *ws.Hub, chatService *messages.Service) *WSHandler {
	return &WSHandler{
		jwtManager:  jwtManager,
		hub:         hub,
		chatService: chatService,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allowing all for dev
	},
}

func (h *WSHandler) HandleConnection(c echo.Context) error {
	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
	}

	// This identifies which project "room" the user is joining
	pIDStr := c.QueryParam("projectId")
	var projectID int64
	if pIDStr != "" {
		pID, err := strconv.ParseInt(pIDStr, 10, 64)
		if err == nil {
			projectID = pID
		}
	}

	//  Validate user
	claims, err := h.jwtManager.VerifyToken(tokenStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
		// enable TCP Keep-Alive
		if err := tcpConn.SetKeepAlive(true); err != nil {
			slog.Warn("failed to enable TCP keep-alive", "remote_addr", tcpConn.RemoteAddr(), "err", err)
		}
		// how long to wait before starting probes
		if err := tcpConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
			slog.Warn("failed to set TCP keep-alive period", "remote_addr", tcpConn.RemoteAddr(), "err", err)
		}
	}

	// include the ProjectID so the Hub knows where to route messages
	client := ws.NewClient(claims.UserID, projectID, conn)

	// ensure cleanup happens exactly once
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			// master kill switch: this forces ReadMessage to return err immediately
			if err := conn.SetReadDeadline(time.Now()); err != nil {
				slog.Warn("failed to force read deadline in cleanup", "user_id", claims.UserID, "err", err)
			}
			if err := conn.SetWriteDeadline(time.Now()); err != nil {
				slog.Warn("failed to force write deadline in cleanup", "user_id", claims.UserID, "err", err)
			}
			if err := conn.Close(); err != nil {
				slog.Warn("failed to close connection in cleanup", "user_id", claims.UserID, "err", err)
			}

			select {
			case h.hub.Unregister <- client:
			default:
			}
			slog.Info("ws client disconnected", "user_id", claims.UserID)
		})
	}

	// Register the client
	h.hub.Register <- client
	slog.Info("ws client registered", "user_id", claims.UserID)

	// WritePump handles the outgoing message queue and the AWS heartbeat (ping),
	// it ensures only one goroutine is writing to the connection
	go func() {
		// a ticker for the AWS Heartbeat
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			cleanup()
		}()

		for {
			select {
			case message, ok := <-client.Send:
				// deadline: if the write takes > 10s, kill the connection
				if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					slog.Warn("failed to set write deadline during pong handler",
						"remote_addr", conn.RemoteAddr(),
						"err", err)
				}

				if !ok {
					// the Hub closed the channel (example: server shutting down)
					if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
						slog.Warn("failed to send websocket close message",
							"remote_addr", conn.RemoteAddr(),
							"err", err)
					}
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			case <-ticker.C:
				// send a Ping every ~54 seconds
				if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					slog.Warn("failed to set write deadline in write pump",
						"remote_addr", conn.RemoteAddr(),
						"err", err)
				}
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}

			case <-client.DoneChan(): // safe read-only access
				//	close(client.Send)
				return
			}
		}
	}()

	// configure the connection to handle Pongs and set deadlines
	defer cleanup()               // Trigger cleanup when this loop breaks
	conn.SetReadLimit(512 * 1024) // max message size 512KB
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		slog.Warn("failed to reset read deadline after pong",
			"remote_addr", conn.RemoteAddr(),
			"err", err)
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			slog.Warn("failed to reset read deadline after pong",
				"remote_addr", conn.RemoteAddr(),
				"err", err)
		}
		return nil
	})

	// ReadPump listens for incoming messages from the client,
	// it sets a read deadline to detect 'zombie' connections if pongs aren't received
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			slog.Debug("read error", "user_id", claims.UserID, "err", err)
			// this will trigger if the user closes the tab OR if we don't get a Pong in time
			break // This exits the loop and hits the defer cleanup()
		}

		var msg WSIncomingMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			h.sendWSError(client, "Invalid JSON format")
			continue
		}

		// triggers if: not already active AND (it's the manual signal OR an actual message)
		if !client.IsActive && (msg.Type == "chat_open" || msg.Type == "direct_message" || msg.Type == "project_chat") {
			metrics.ActiveChatConnections.Inc()
			client.IsActive = true
			slog.Info("connection promoted to active chat", "user_id", client.UserID, "type", msg.Type)
		}

		ctx := context.Background()
		var chatErr error

		switch msg.Type {
		case "chat_open":
			// this is basically for grafana to see that the chat is open
			continue
		case "project_chat":
			_, chatErr = h.chatService.SendProjectMessage(ctx, claims.UserID, msg.ProjectID, msg.Content)
		case "direct_message":
			_, chatErr = h.chatService.SendDirectMessage(ctx, claims.UserID, msg.ReceiverID, msg.Content)
		case "typing":
			// create a small JSON payload to tell others who is typing
			typingSignal, _ := json.Marshal(map[string]interface{}{
				"type":       "user_typing",
				"user_id":    claims.UserID,
				"email":      claims.Email,
				"project_id": msg.ProjectID,
				"is_typing":  msg.IsTyping,
			})

			//  Route based on context
			if msg.ReceiverID != 0 {
				// It's a Direct Message: Send only to the specific person
				h.hub.SendToUser(msg.ReceiverID, typingSignal)
				continue
			} else if msg.ProjectID != 0 {
				// It's a Project Chat: Broadcast to the whole room
				h.hub.BroadcastToProjectExcept(msg.ProjectID, claims.UserID, typingSignal)
			}

		case "read_receipt":
			// persist the "Read" status to the database via service
			// pass the claims.UserID to ensure only the recipient can mark it read
			chatErr = h.chatService.MarkAsRead(ctx, msg.MessageID, claims.UserID)

			if chatErr == nil {
				// prepare the notification for the sender
				receiptSignal, _ := json.Marshal(map[string]interface{}{
					"type":       "message_read",
					"message_id": msg.MessageID,
					"reader_id":  claims.UserID,
					"project_id": msg.ProjectID,
				})

				if msg.ReceiverID != 0 {
					// it's a DM: tell the original sender
					h.hub.SendToUser(msg.ReceiverID, receiptSignal)
				} else if msg.ProjectID != 0 {
					// it's a project message: broadcast to the room so the sender's UI updates
					h.hub.BroadcastToProjectExcept(msg.ProjectID, claims.UserID, receiptSignal)
				}
			}
		default:
			slog.Warn("ws unknown message type", "type", msg.Type, "user_id", claims.UserID)
			continue
		}

		if chatErr != nil {
			slog.Error("ws chat processing failed", "user_id", claims.UserID, "error", chatErr)
			h.sendWSError(client, "Could not send message")
			continue
		}
	}

	return nil
}

// Helper method to send error messages over the socket
func (h *WSHandler) sendWSError(client *ws.Client, message string) {
	errPayload := map[string]string{
		"type":  "error",
		"error": message,
	}
	b, _ := json.Marshal(errPayload)
	select {
	case client.Send <- b:
		// message sent successfully to the client's write-pump
	default:
		// drop message if the client's 256-message buffer is totally full (if his connection is extremely slow for example)
		slog.Warn("ws error message dropped",
			"user_id", client.UserID,
			"reason", "buffer_full",
			"dropped_message", message)
	}
}

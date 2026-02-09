package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
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

	// include the ProjectID so the Hub knows where to route messages
	client := ws.NewClient(claims.UserID, projectID, conn)

	h.hub.Register <- client

	slog.Info("ws client connected", "user_id", claims.UserID, "project_id", projectID)

	// This goroutine listens to the Hub and pushes messages to the browser
	go func() {
		defer conn.Close()
		for {
			select {
			case message, ok := <-client.Send:
				if !ok {
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			case <-client.DoneChan(): // safe read-only access
				close(client.Send)
				return
			}
		}
	}()

	// Cleanup
	defer func() {
		h.hub.Unregister <- client
		slog.Info("ws client disconnected", "user_id", claims.UserID)
	}()

	// The Read Loop (The "Ear")
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WSIncomingMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			h.sendWSError(conn, "Invalid JSON format")
			continue
		}

		ctx := context.Background()
		var chatErr error

		switch msg.Type {
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
			h.sendWSError(conn, "Could not send message")
			continue
		}
	}

	return nil
}

// Helper method to send error messages over the socket
func (h *WSHandler) sendWSError(conn *websocket.Conn, message string) {
	errPayload := map[string]string{
		"type":  "error",
		"error": message,
	}
	b, _ := json.Marshal(errPayload)
	conn.WriteMessage(websocket.TextMessage, b)
}

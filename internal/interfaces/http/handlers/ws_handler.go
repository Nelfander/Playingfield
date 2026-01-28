package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	Type       string `json:"type"`
	ProjectID  int64  `json:"project_id"`
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
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
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "missing token"})
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
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// include the ProjectID so the Hub knows where to route messages
	client := &ws.Client{
		UserID:    claims.UserID,
		ProjectID: projectID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	h.hub.Register <- client

	// This goroutine listens to the Hub and pushes messages to the browser
	go func() {
		for {
			message, ok := <-client.Send
			if !ok {
				return
			}
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}()

	// Cleanup
	defer func() {
		h.hub.Unregister <- client
		conn.Close()
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
		default:
			fmt.Printf("Unknown message type: %s\n", msg.Type)
			continue
		}

		if chatErr != nil {
			fmt.Printf("Chat error: %v\n", chatErr)
			h.sendWSError(conn, chatErr.Error())
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

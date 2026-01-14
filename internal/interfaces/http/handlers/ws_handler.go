package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/ws"
)

type WSHandler struct {
	jwtManager *auth.JWTManager
	hub        *ws.Hub
}

func NewWSHandler(jwtManager *auth.JWTManager, hub *ws.Hub) *WSHandler {
	return &WSHandler{
		jwtManager: jwtManager,
		hub:        hub,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allowing all for dev
	},
}

func (h *WSHandler) HandleConnection(c echo.Context) error {
	//  Get token from query string
	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "missing token"})
	}

	// Validate user
	claims, err := h.jwtManager.VerifyToken(tokenStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	//  Create Client
	client := &ws.Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	//  Register with Hub
	h.hub.Register <- client

	// This goroutine listens to the Hub and pushes messages to the browser
	go func() {
		for {
			message, ok := <-client.Send
			if !ok {
				// Hub closed the channel, connection ending
				return
			}
			// Write the message to the browser
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
		//fmt.Printf("User %d (Email: %s) disconnected\n", claims.UserID, claims.Email)
	}()

	//fmt.Printf("User %d (Email: %s) connected and registered in Hub!\n", claims.UserID, claims.Email)

	// The Read Loop (The "Ear")
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Printf("Message received from %d: %s\n", claims.UserID, string(payload))
	}

	return nil
}

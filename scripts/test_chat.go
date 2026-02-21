package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	baseURL   = "http://localhost:880"
	wsURL     = "ws://localhost:880/ws"
	testEmail = "test1@example.com"
	testPass  = "123456"
)

func main() {
	// 1. Login to get a fresh token
	token, err := getAuthToken(testEmail, testPass)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Println("✅ Authenticated")

	// 2. Connect to WebSocket
	u, _ := url.Parse(wsURL)
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("WS Connection failed: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("failed to close websocket client connection in test_chat: %v", err) // no need to slog for this script
		}
	}()
	fmt.Println("✅ WebSocket Connected")

	// 3. Start Listener
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			fmt.Printf("\n[RECEIVE] %s\n", string(message))
		}
	}()

	// 4. Test Scenario: Send Project Message
	// This tests: JSON parsing -> DB Save -> Member Lookup -> Targeted Broadcast
	msg := map[string]interface{}{
		"type":       "project_chat",
		"project_id": 20, // Change this to a project the user IS in
		"content":    "Automated test message at " + time.Now().Format(time.Kitchen),
	}

	payload, _ := json.Marshal(msg)
	err = c.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		log.Printf("Send failed: %v", err)
	}
	fmt.Println("🚀 Message Sent. Waiting for echo/notification...")

	// Keep alive for 5 seconds to catch the broadcast
	time.Sleep(5 * time.Second)
}

func getAuthToken(email, password string) (string, error) {
	loginData, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close HTTP response body in test_chat: %v", err) // no need to slog for script
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("failed to decode JSON response in test_chat script: %v", err)
	}
	return result.Token, nil
}

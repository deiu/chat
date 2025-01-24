/*
Copyright 2025 Andrei Sambra

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewChatServer(t *testing.T) {
	server := NewChatServer()
	if server == nil {
		t.Fatal("NewChatServer returned nil")
	}
	if server.clients == nil {
		t.Error("clients map not initialized")
	}
}

func TestHandleWebSocketUsernameValidation(t *testing.T) {
	server := NewChatServer()

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"empty username", "", true},
		{"valid username", "testuser", false},
		{"duplicate username", "testuser", true},
		{"case insensitive duplicate", "TestUser", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.HandleWebSocket(w, r)
			}))
			defer s.Close()

			// Convert http to ws
			wsURL := "ws" + strings.TrimPrefix(s.URL, "http") + "?username=" + tt.username

			_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func setupWebSocketConnection(t *testing.T, server *ChatServer, username string) (*websocket.Conn, *httptest.Server) {
	t.Helper()

	var conn *websocket.Conn
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(w, r)
	}))

	// Convert http to ws
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http") + "?username=" + username

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open websocket connection: %v", err)
	}
	conn = ws

	return conn, s
}

func TestWebSocketCommunication(t *testing.T) {
	server := NewChatServer()

	// Connect first client
	conn1, s1 := setupWebSocketConnection(t, server, "user1")
	defer s1.Close()
	defer conn1.Close()

	// Connect second client
	conn2, s2 := setupWebSocketConnection(t, server, "user2")
	defer s2.Close()
	defer conn2.Close()

	// Set read deadline
	conn2.SetReadDeadline(time.Now().Add(time.Second))

	// Test sending message
	message := Message{
		To:      "user2",
		Content: "Hello",
	}

	if err := conn1.WriteJSON(message); err != nil {
		t.Fatalf("could not send message: %v", err)
	}

	// Read messages until we get our test message
	conn2.SetReadDeadline(time.Now().Add(time.Second))
	for {
		var raw json.RawMessage
		if err := conn2.ReadJSON(&raw); err != nil {
			t.Fatalf("could not receive message: %v", err)
		}

		// Try to decode as Message
		var received Message
		if err := json.Unmarshal(raw, &received); err == nil && received.Content == message.Content {
			// Found our message
			if received.Content != message.Content {
				t.Errorf("received message content = %v, want %v", received.Content, message.Content)
			}
			break
		}
	}
}

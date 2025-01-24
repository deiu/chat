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
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	clients    map[string]*Client // key is username
	logoutChan chan string
	upgrader   websocket.Upgrader
	mu         sync.Mutex
}

type OnlineUser struct {
	Username string `json:"username"`
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: make(map[string]*Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logoutChan: make(chan string),
	}
}

func (s *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Check for username uniqueness case-insensitively
	s.mu.Lock()
	lowercaseUsername := strings.ToLower(username)
	for existingUser := range s.clients {
		if strings.ToLower(existingUser) == lowercaseUsername {
			s.mu.Unlock()
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.mu.Unlock()
		log.Printf("Websocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:     conn,
		username: username, // Keep original case for display
		sendChan: make(chan Message, 256),
	}

	s.clients[username] = client
	s.mu.Unlock()

	go func() {
		username := <-s.logoutChan
		if username == client.username {
			s.removeClient(username)
			conn.Close()
		}
	}()

	go s.broadcastOnlineUsers()
	go client.writePump()
	go client.readPump(s)
}

// Add new method to handle logout
func (s *ChatServer) HandleLogout(username string) {
	s.mu.Lock()
	if client, exists := s.clients[username]; exists {
		client.conn.Close()
		delete(s.clients, username)
	}
	s.mu.Unlock()

	// Broadcast updated online users list
	s.broadcastOnlineUsers()
}

func (s *ChatServer) HandleGetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	users := make([]OnlineUser, 0, len(s.clients))
	for username := range s.clients {
		users = append(users, OnlineUser{Username: username})
	}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *ChatServer) broadcastOnlineUsers() {
	s.mu.Lock()
	users := make([]OnlineUser, 0, len(s.clients))
	for username := range s.clients {
		users = append(users, OnlineUser{Username: username})
	}
	s.mu.Unlock()

	usersJSON, _ := json.Marshal(users)

	s.mu.Lock()
	for _, client := range s.clients {
		client.conn.WriteMessage(websocket.TextMessage, usersJSON)
	}
	s.mu.Unlock()
}

func (s *ChatServer) removeClient(username string) {
	s.mu.Lock()
	if client, exists := s.clients[username]; exists {
		delete(s.clients, username)
		close(client.sendChan)
	}
	s.mu.Unlock()

	// Notify all clients about the logout
	s.mu.Lock()
	logoutMsg := Message{
		Type:     "logout",
		Username: username,
	}
	msgBytes, _ := json.Marshal(logoutMsg)
	for _, client := range s.clients {
		client.conn.WriteMessage(websocket.TextMessage, msgBytes)
	}
	s.mu.Unlock()

	// Broadcast updated user list
	s.broadcastOnlineUsers()
}

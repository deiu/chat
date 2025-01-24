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
	"sync"
)

type ChatRoom struct {
	clients map[*Client]bool
	mu      sync.Mutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients: make(map[*Client]bool),
	}
}

func (cr *ChatRoom) AddClient(client *Client) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.clients[client] = true
}

func (cr *ChatRoom) RemoveClient(client *Client) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	delete(cr.clients, client)
}

func (cr *ChatRoom) GetClients() map[*Client]bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	// Return a copy to prevent external modifications
	clientsCopy := make(map[*Client]bool)
	for client, value := range cr.clients {
		clientsCopy[client] = value
	}
	return clientsCopy
}

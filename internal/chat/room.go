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

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

import "github.com/gorilla/websocket"

type Client struct {
	conn     *websocket.Conn
	username string
	sendChan chan Message
}

type Message struct {
	Type     string `json:"type,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Content  string `json:"content,omitempty"`
	Username string `json:"username,omitempty"` // For logout notifications
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.sendChan {
		err := c.conn.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func (c *Client) readPump(s *ChatServer) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, c.username)
		s.mu.Unlock()
		c.conn.Close()
	}()

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			break
		}

		message.From = c.username
		s.mu.Lock()
		if targetClient, exists := s.clients[message.To]; exists {
			select {
			case targetClient.sendChan <- message:
			default:
				close(targetClient.sendChan)
				delete(s.clients, targetClient.username)
			}
		}
		s.mu.Unlock()
	}
}

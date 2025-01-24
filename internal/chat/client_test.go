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
	"testing"
	"time"
)

func TestMessageSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "regular message",
			message: Message{
				From:    "user1",
				To:      "user2",
				Content: "Hello",
			},
		},
		{
			name: "logout message",
			message: Message{
				Type:     "logout",
				Username: "user1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("failed to marshal message: %v", err)
			}

			// Unmarshal
			var decoded Message
			err = json.Unmarshal(data, &decoded)
			if err != nil {
				t.Fatalf("failed to unmarshal message: %v", err)
			}

			// Compare
			if decoded.Type != tt.message.Type ||
				decoded.From != tt.message.From ||
				decoded.To != tt.message.To ||
				decoded.Content != tt.message.Content ||
				decoded.Username != tt.message.Username {
				t.Errorf("decoded message = %+v, want %+v", decoded, tt.message)
			}
		})
	}
}

func TestClientMessageBuffering(t *testing.T) {
	server := NewChatServer()
	conn1, s1 := setupWebSocketConnection(t, server, "sender")
	defer conn1.Close()
	defer s1.Close()

	conn2, s2 := setupWebSocketConnection(t, server, "receiver")
	defer conn2.Close()
	defer s2.Close()

	// Send multiple messages
	messages := []string{
		"Message 1",
		"Message 2",
		"Message 3",
	}

	for _, msg := range messages {
		err := conn1.WriteJSON(Message{
			To:      "receiver",
			Content: msg,
		})
		if err != nil {
			t.Fatalf("failed to send message: %v", err)
		}
	}

	// Verify all messages are received in order
	for _, want := range messages {
		// Read messages until we get a chat message
		for {
			var raw json.RawMessage
			err := conn2.ReadJSON(&raw)
			if err != nil {
				t.Fatalf("failed to receive message: %v", err)
			}

			// Try to decode as Message
			var msg Message
			if err := json.Unmarshal(raw, &msg); err == nil && msg.Content != "" {
				if msg.Content != want {
					t.Errorf("received message = %v, want %v", msg.Content, want)
				}
				break
			}
			// If not a chat message, continue reading
		}
	}
}

func TestClientDisconnection(t *testing.T) {
	server := NewChatServer()

	// Connect a client
	conn, s := setupWebSocketConnection(t, server, "disconnectuser")
	defer s.Close()

	// Close the connection
	conn.Close()

	// Give some time for the server to process the disconnection
	time.Sleep(100 * time.Millisecond)

	// Verify the client was removed from the server
	server.mu.Lock()
	if _, exists := server.clients["disconnectuser"]; exists {
		t.Error("client was not removed after disconnection")
	}
	server.mu.Unlock()
}

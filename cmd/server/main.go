package main

import (
	"chat/internal/chat"
	"log"
	"net/http"
)

func main() {
	server := chat.NewChatServer()

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve the main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	// Handle WebSocket connections
	http.HandleFunc("/ws", server.HandleWebSocket)

	// Handle getting online users
	http.HandleFunc("/users", server.HandleGetOnlineUsers)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

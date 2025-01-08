package Functions

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (secure this for production)
	},
}

// Client management
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var mu sync.Mutex

// Message structure
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// Handle incoming WebSocket connections
func HandleConnections(c *gin.Context) {
	// Upgrade HTTP request to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Register client
	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	// Read messages from the client
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}

		// Send the message to the broadcast channel
		broadcast <- msg
	}
}

// Broadcast messages to all connected clients
func BroadcastMessages() {
	for {
		msg := <-broadcast

		// Send message to all connected clients
		mu.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error broadcasting to client: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

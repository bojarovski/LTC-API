package Functions

import (
	"backend/FunctionsHelper"
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

		// Check the message content with AI
		isAppropriate, err := FunctionsHelper.CallAIService(msg.Content, 1, "You are a bot that checks if the post is appropriate or not. By appropriate it is meant there are bad words. If it is appropriate return 1; else return 0.")
		if err != nil {
			log.Printf("AI check error: %v", err)
			continue
		}

		if isAppropriate == "1" {
			// Add to the broadcast channel if appropriate
			broadcast <- msg
		} else {
			// Log or notify that the message was blocked
			log.Printf("Message blocked by AI: %s", msg.Content)
			hiddenMessage := Message{
				Username: msg.Username,
				Content:  "This message was hidden by AI moderation.",
			}
			broadcast <- hiddenMessage
		}
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

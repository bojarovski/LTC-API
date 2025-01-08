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

// ChatRoom structure
type ChatRoom struct {
	Name      string
	Clients   map[*websocket.Conn]bool
	Broadcast chan Message
}

var rooms = make(map[string]*ChatRoom) // Stores all chatrooms
var roomsMu sync.Mutex                 // Mutex for thread-safe room operations

// Message structure
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// Create a new chatroom
func CreateRoom(c *gin.Context) {
	var req struct {
		RoomName string `json:"room_name"`
	}

	// Parse the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room name"})
		return
	}

	if req.RoomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room name cannot be empty"})
		return
	}

	// Check if the room already exists
	roomsMu.Lock()
	if _, exists := rooms[req.RoomName]; exists {
		roomsMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": "Room already exists"})
		return
	}

	// Create the new room
	room := &ChatRoom{
		Name:      req.RoomName,
		Clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
	}
	rooms[req.RoomName] = room
	roomsMu.Unlock()

	// Start broadcasting messages for this room
	go broadcastRoomMessages(room)

	c.JSON(http.StatusCreated, gin.H{"message": "Room created successfully", "room_name": req.RoomName})
}

// Get a list of all available chatrooms
func GetAllRooms(c *gin.Context) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	// Collect room names
	var roomNames []string
	for roomName := range rooms {
		roomNames = append(roomNames, roomName)
	}

	c.JSON(http.StatusOK, gin.H{"rooms": roomNames})
}

// Handle WebSocket connections for a specific room
func HandleConnections(c *gin.Context) {
	// Get the room name from the query parameter
	roomName := c.Query("room")
	if roomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room name is required"})
		return
	}

	// Check if the room exists
	roomsMu.Lock()
	room, exists := rooms[roomName]
	roomsMu.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Upgrade HTTP request to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Register client in the room
	roomsMu.Lock()
	room.Clients[conn] = true
	roomsMu.Unlock()

	// Read messages from the client
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			roomsMu.Lock()
			delete(room.Clients, conn)
			roomsMu.Unlock()
			break
		}

		// Check the message content with AI
		isAppropriate, err := FunctionsHelper.CallAIService(msg.Content, 1, "You are a bot that checks if the post is appropriate or not. By appropriate it is meant there are bad words. If it is appropriate return 1; else return 0.")
		if err != nil {
			log.Printf("AI check error: %v", err)
			continue
		}

		if isAppropriate == "1" {
			// Add to the room's broadcast channel if appropriate
			room.Broadcast <- msg
		} else {
			// Send a hidden message for moderation
			log.Printf("Message blocked by AI: %s", msg.Content)
			hiddenMessage := Message{
				Username: msg.Username,
				Content:  "This message was hidden by AI moderation.",
			}
			room.Broadcast <- hiddenMessage
		}
	}
}

// Broadcast messages to all clients in a room
func broadcastRoomMessages(room *ChatRoom) {
	for {
		msg := <-room.Broadcast

		// Send message to all connected clients in the room
		roomsMu.Lock()
		for client := range room.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error broadcasting to client: %v", err)
				client.Close()
				delete(room.Clients, client)
			}
		}
		roomsMu.Unlock()
	}
}

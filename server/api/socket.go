package api

import (
	"encoding/json"
	"log"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Global socket manager with mutex for thread safety
	socketManager = &m.SocketManager{
		Sockets: make(map[uint64]*websocket.Conn),
		Mu:      sync.RWMutex{},
	}

	// Channel for broadcasting messages
	broadcast = make(chan m.BroadcastMessage, 100)
)

func init() {
	// Start the broadcast handler
	go handleBroadcasts()
}

func handleBroadcasts() {
	for msg := range broadcast {
		socketManager.Mu.RLock()
		for userID, conn := range socketManager.Sockets {
			// Skip if this message is not for this user
			if msg.TargetUsers != nil && !msg.TargetUsers[userID] {
				continue
			}

			if err := conn.WriteJSON(msg.Data); err != nil {
				log.Printf("Error broadcasting to user %d: %v", userID, err)
				conn.Close()
				delete(socketManager.Sockets, userID)
			}
		}
		socketManager.Mu.RUnlock()
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("WebSocket session error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID uint64
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("WebSocket database error: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Configure WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Set connection properties
	conn.SetReadLimit(4096) // 4KB message size limit
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Handle existing connection
	socketManager.Mu.Lock()
	if existingConn, exists := socketManager.Sockets[userID]; exists {
		// Send close message to existing connection
		existingConn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "New connection established"),
		)
		existingConn.Close()
		log.Printf("Closed existing connection for user %s", username)
	}
	socketManager.Sockets[userID] = conn
	socketManager.Mu.Unlock()

	// Start ping ticker
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	// Clean up on disconnect
	defer func() {
		socketManager.Mu.Lock()
		if _, ok := socketManager.Sockets[userID]; ok {
			delete(socketManager.Sockets, userID)
			log.Printf("Removed connection for user %s", username)
		}
		socketManager.Mu.Unlock()
		conn.Close()
	}()

	// Handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch messageType {
		case websocket.TextMessage:

			// getting the request type. from this we can unmarshall into the actual message
			var connectionType m.ConnectionType
			if err := json.Unmarshal(message, &connectionType); err != nil {
				log.Printf("WebSocket json unmarshal error: %v", err)
				break
			}

			// this can be made into a switch to handle future message types. but for now it's chat only
			if connectionType.Type == "chat" {
				var chatMessage m.ChatMessage
				if err := json.Unmarshal(message, &chatMessage); err != nil {
					log.Printf("WebSocket chat json unmarshal error: %v", err)
					break
				}

				SaveMessage(chatMessage)

				// Echo the message back (for real this time)
				if err := conn.WriteMessage(messageType, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					break
				}
			}

			// Add this to your existing socket message types
			type EventRSVPMessage struct {
				Type     string `json:"type"`
				GroupID  int    `json:"groupId"`
				EventID  int    `json:"eventId"`
				Status   string `json:"status"`
				Going    int    `json:"going"`
				NotGoing int    `json:"notGoing"`
			}

			// In your WebSocketHandler function, add handling for event responses
			if connectionType.Type == "eventRSVP" {
				var rsvpMessage EventRSVPMessage
				if err := json.Unmarshal(message, &rsvpMessage); err != nil {
					log.Printf("WebSocket event RSVP unmarshal error: %v", err)
					break
				}

				// Broadcast the RSVP update to all connected clients
				broadcast <- m.BroadcastMessage{
					Data:        rsvpMessage,
					TargetUsers: nil, // Broadcast to all users
				}
			}

		case websocket.BinaryMessage:
			log.Printf("Received Binary Message: %v\n", message)
		case websocket.CloseMessage:
			log.Println("Received Close Message")
			return
		case websocket.PingMessage:
			log.Println("Received Ping Message")
		case websocket.PongMessage:
			log.Println("Received Pong Message")
		}
	}
}

// SendNotification sends a notification to specific users
func SendNotification(userIDs []uint64, notification interface{}) {
	targetUsers := make(map[uint64]bool)
	for _, id := range userIDs {
		targetUsers[id] = true
	}

	broadcast <- m.BroadcastMessage{
		Data:        notification,
		TargetUsers: targetUsers,
	}
}

// Broadcast sends a message to all connected users
func Broadcast(message interface{}) {
	broadcast <- m.BroadcastMessage{
		Data:        message,
		TargetUsers: nil, // nil means broadcast to all
	}
}

// Helper functions for different message types
func handleChatMessage(msg m.WebSocketMessage, senderID uint64) {
	// Handle chat message logic
}

func handleNotification(msg m.WebSocketMessage, senderID uint64) {
	// Handle notification logic
}

func handleTypingIndicator(msg m.WebSocketMessage, senderID uint64) {
	// Handle typing indicator logic
}

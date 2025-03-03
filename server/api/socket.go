package api

import (
	"encoding/json"
	"log"
	"net/http"
	models "social-network/models"
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
	socketManager = &models.SocketManager{
		Sockets: make(map[int]*websocket.Conn),
		Mu:      sync.RWMutex{},
	}

	// Channel for broadcasting messages
	broadcast = make(chan models.BroadcastMessage, 100)
)

func init() {
	// Start the broadcast handler
	go handleBroadcasts()
}

func handleBroadcasts() {
	for msg := range broadcast {
		log.Printf("Starting broadcast: Data=%+v, TargetUsers=%+v", msg.Data, msg.TargetUsers)

		socketManager.Mu.RLock()
		activeConnections := len(socketManager.Sockets)
		log.Printf("Broadcasting to %d active connections", activeConnections)

		for userID, conn := range socketManager.Sockets {
			if msg.TargetUsers != nil && !msg.TargetUsers[userID] {
				continue
			}

			// Create a copy of the connection for this iteration
			currentConn := conn

			// Use a goroutine to handle each send operation
			go func(userID int, conn *websocket.Conn) {
				conn.SetWriteDeadline(time.Now().Add(time.Second * 5))

				// Check connection before sending
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
					log.Printf("Connection check failed for user %d: %v", userID, err)
					socketManager.Mu.Lock()
					conn.Close()
					delete(socketManager.Sockets, userID)
					socketManager.Mu.Unlock()
					return
				}

				if err := conn.WriteJSON(msg.Data); err != nil {
					log.Printf("Failed to send to user %d: %v", userID, err)
					socketManager.Mu.Lock()
					conn.Close()
					delete(socketManager.Sockets, userID)
					socketManager.Mu.Unlock()
				} else {
					log.Printf("Successfully sent notification to user %d: %+v", userID, msg.Data)
				}
			}(userID, currentConn)
		}
		socketManager.Mu.RUnlock()
	}
}

// Add helper function to get connected user IDs
func getConnectedUserIDs() []int {
	userIDs := make([]int, 0, len(socketManager.Sockets))
	for userID := range socketManager.Sockets {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("New WebSocket connection attempt")

	// Use a single mutex for connection management
	var connectionMutex sync.Mutex
	connectionMutex.Lock()
	defer connectionMutex.Unlock()

	// Get user info from session before upgrading
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Error getting username from session: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID from username
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user already has an active connection
	socketManager.Mu.Lock()
	if existingConn, exists := socketManager.Sockets[userID]; exists {
		log.Printf("User %d already has an active connection, closing it", userID)
		existingConn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "New connection"),
			time.Now().Add(time.Second),
		)
		existingConn.Close()
		delete(socketManager.Sockets, userID)
		// Add a small delay to ensure proper cleanup
		time.Sleep(time.Millisecond * 100)
	}
	socketManager.Mu.Unlock()

	// Upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Set up connection cleanup
	defer func() {
		log.Printf("Cleaning up connection for user %d", userID)
		socketManager.Mu.Lock()
		delete(socketManager.Sockets, userID)
		socketManager.Mu.Unlock()
		conn.Close()
	}()

	// Store the new connection
	socketManager.Mu.Lock()
	socketManager.Sockets[userID] = conn
	socketManager.Mu.Unlock()

	log.Printf("WebSocket connection established for user %d", userID)

	// Setup ping/pong handlers
	conn.SetPingHandler(func(data string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second*5))
		if err == nil {
			conn.SetReadDeadline(time.Now().Add(time.Second * 120))
		}
		return err
	})

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(time.Second * 120))
		return nil
	})

	// Main message loop
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
			var msg models.WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("WebSocket json unmarshal error: %v", err)
				break
			}

			switch msg.Type {
			case "chat":
				// Extract the chat message from the data field
				chatData, err := json.Marshal(msg.Data)
				if err != nil {
					log.Printf("Error marshaling chat data: %v", err)
					break
				}
				var chatMessage models.ChatMessage
				if err := json.Unmarshal(chatData, &chatMessage); err != nil {
					log.Printf("WebSocket chat json unmarshal error: %v", err)
					break
				}

				if err := SaveMessage(chatMessage); err != nil {
					log.Printf("Error saving message: %v", err)
					break
				}

				// Echo the message back to the sender
				if err := conn.WriteMessage(messageType, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					break
				}

				// NEW CODE: Send the message to the recipient if they're online
				recipientID := chatMessage.RecipientID
				socketManager.Mu.RLock()
				recipientConn, recipientOnline := socketManager.Sockets[recipientID]
				socketManager.Mu.RUnlock()

				if recipientOnline {
					log.Printf("Recipient %d is online, sending message", recipientID)
					if err := recipientConn.WriteMessage(messageType, message); err != nil {
						log.Printf("Error sending message to recipient %d: %v", recipientID, err)
					}
				} else {
					log.Printf("Recipient %d is offline", recipientID)
				}
			case "eventRSVP":
				var rsvpMessage models.EventRSVPMessage
				if err := json.Unmarshal(message, &rsvpMessage); err != nil {
					log.Printf("WebSocket event RSVP unmarshal error: %v", err)
					break
				}

				// Broadcast the RSVP update to all connected clients
				broadcast <- models.BroadcastMessage{
					Data:        rsvpMessage,
					TargetUsers: nil, // Broadcast to all users
				}
			case "ping":
				pongMessage := models.WebSocketMessage{
					Type: "pong",
					Data: nil,
				}

				// Send the pong response back to the client
				if err := conn.WriteJSON(pongMessage); err != nil {
					log.Printf("Error sending pong: %v", err)
				}
			default:
				log.Printf("Unknown message type: %s", msg.Type)
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

// SaveMessage saves a chat message to the database
func SaveMessage(message models.ChatMessage) error {
	_, err := sqlite.DB.Exec(`
        INSERT INTO chat_messages (
            chat_id,
            sender_id,
            content,
            created_at
        ) VALUES (?, ?, ?, ?)`,
		message.ChatID,
		message.SenderID,
		message.Content,
		message.CreatedAt,
	)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		return err
	}
	return nil
}

// SendNotification sends a notification to specific users
func SendNotification(userIDs []int, notification interface{}) {
	targetUsers := make(map[int]bool)
	for _, id := range userIDs {
		targetUsers[id] = true
	}

	broadcast <- models.BroadcastMessage{
		Data:        notification,
		TargetUsers: targetUsers,
	}
}

// Broadcast sends a message to all connected users
func Broadcast(message interface{}) {
	broadcast <- models.BroadcastMessage{
		Data:        message,
		TargetUsers: nil, // nil means broadcast to all
	}
}

func cleanupConnection(userID int, conn *websocket.Conn) {
	socketManager.Mu.Lock()
	defer socketManager.Mu.Unlock()

	if currentConn, exists := socketManager.Sockets[userID]; exists && currentConn == conn {
		delete(socketManager.Sockets, userID)
		conn.Close()
		log.Printf("Connection cleaned up for user %d", userID)
	}
}

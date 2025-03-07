package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
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
					
					// Send an error response back to the sender
					errorResponse := models.WebSocketMessage{
						Type: "error",
						Data: map[string]interface{}{
							"message": err.Error(),
							"code": "follow_required",
						},
					}
					
					if err := conn.WriteJSON(errorResponse); err != nil {
						log.Printf("Error sending error response to client: %v", err)
					}
					
					break
				}

				// Echo the message back to the sender
				if err := conn.WriteMessage(messageType, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					break
				}

				// Create a notification for the chat message
				if err := createChatNotification(chatMessage); err != nil {
					log.Printf("Error creating chat notification: %v", err)
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

// SaveMessage saves a message to the database and returns its ID
func SaveMessage(message models.ChatMessage) error {
	// Check if the message is for a direct chat
	var chatType string
	err := sqlite.DB.QueryRow("SELECT type FROM chats WHERE id = ?", message.ChatID).Scan(&chatType)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}

	// For direct chats, check if there's at least one follow relationship
	if chatType == "direct" {
		// Get the other participant in the chat
		var otherUserID int
		err := sqlite.DB.QueryRow(`
			SELECT user_id FROM user_chat_status 
			WHERE chat_id = ? AND user_id != ?
		`, message.ChatID, message.SenderID).Scan(&otherUserID)

		if err != nil {
			return fmt.Errorf("failed to find other participant: %w", err)
		}

		// Check if at least one user follows the other
		var followExists bool
		err = sqlite.DB.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM followers 
				WHERE ((follower_id = ? AND followed_id = ?) 
				OR (follower_id = ? AND followed_id = ?))
				AND status = 'accepted'
			)
		`, message.SenderID, otherUserID, otherUserID, message.SenderID).Scan(&followExists)

		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}

		if !followExists {
			return fmt.Errorf("cannot send message: at least one user must follow the other")
		}
	}

	// Insert the message into the database
	statement, err := sqlite.DB.Prepare(`
		INSERT INTO chat_messages (chat_id, sender_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer statement.Close()

	result, err := statement.Exec(message.ChatID, message.SenderID, message.Content)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get message ID: %w", err)
	}

	message.ID = int(messageID)

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

func createChatNotification(message models.ChatMessage) error {
	// Call the notification creation function
	return CreateChatNotification(message.RecipientID, message.SenderID, message.Content)
}

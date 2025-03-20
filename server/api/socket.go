package api

import (
	"database/sql"
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

	// First acquire lock to check for existing connection
	socketManager.Mu.Lock()

	// Check if user already has an active connection
	existingConn, exists := socketManager.Sockets[userID]
	if exists {
		log.Printf("User %d already has an active connection, closing it", userID)
		// First remove from map to prevent race conditions
		delete(socketManager.Sockets, userID)
		socketManager.Mu.Unlock()

		// Then close the connection outside the lock
		go func() {
			// Send close message with a reason
			existingConn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "New connection"),
				time.Now().Add(time.Second),
			)
			// Wait briefly before forcefully closing
			time.Sleep(time.Millisecond * 200)
			existingConn.Close()
		}()
	} else {
		socketManager.Mu.Unlock()
	}

	// Wait a moment to ensure old connection is fully closed
	time.Sleep(time.Millisecond * 100)

	// Upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Store the new connection - acquire lock again
	socketManager.Mu.Lock()
	socketManager.Sockets[userID] = conn
	socketManager.Mu.Unlock()

	log.Printf("WebSocket connection established for user %d", userID)

	// Set up proper close handler
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("Connection closing for user %d: %d %s", userID, code, text)

		// Remove from socket manager
		socketManager.Mu.Lock()
		if currentConn, ok := socketManager.Sockets[userID]; ok && currentConn == conn {
			delete(socketManager.Sockets, userID)
		}
		socketManager.Mu.Unlock()

		// Return nil to use default close behavior
		return nil
	})

	// Set up ping/pong handlers for connection health checks
	conn.SetPingHandler(func(data string) error {
		log.Printf("Ping received from user %d", userID)
		// Update the read deadline when we get a ping
		conn.SetReadDeadline(time.Now().Add(time.Second * 120))
		// Send pong response
		return conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second*5))
	})

	conn.SetPongHandler(func(string) error {
		log.Printf("Pong received from user %d", userID)
		// Update the read deadline when we get a pong
		conn.SetReadDeadline(time.Now().Add(time.Second * 120))
		return nil
	})

	// Set initial read deadline
	conn.SetReadDeadline(time.Now().Add(time.Second * 120))

	// Use a defer to ensure cleanup if the handler exits
	defer func() {
		log.Printf("Cleaning up connection for user %d", userID)
		socketManager.Mu.Lock()
		if currentConn, ok := socketManager.Sockets[userID]; ok && currentConn == conn {
			delete(socketManager.Sockets, userID)
		}
		socketManager.Mu.Unlock()
		conn.Close()
	}()

	// Main message loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %d: %v", userID, err)
			} else {
				log.Printf("Connection closed for user %d: %v", userID, err)
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
				processChatMessage(userID, conn, messageType, message, msg)

			case "groupChat":
				processGroupChatMessage(userID, conn, messageType, message, msg)

			case "ping":
				// Send pong response
				pongMessage := models.WebSocketMessage{
					Type: "pong",
					Data: map[string]interface{}{
						"timestamp": time.Now().UnixMilli(),
					},
				}

				if err := conn.WriteJSON(pongMessage); err != nil {
					log.Printf("Error sending pong to user %d: %v", userID, err)
				}

			default:
				log.Printf("Unknown message type from user %d: %s", userID, msg.Type)
			}

		case websocket.BinaryMessage:
			log.Printf("Received binary message from user %d", userID)

		case websocket.CloseMessage:
			log.Printf("Received close message from user %d", userID)
			return

		case websocket.PingMessage, websocket.PongMessage:
			// These are handled by the SetPingHandler and SetPongHandler
			continue
		}
	}
}

// Separate function to process chat messages
func processChatMessage(userID int, conn *websocket.Conn, messageType int, rawMessage []byte, msg models.WebSocketMessage) {
	// Extract the chat message from the data field
	chatData, err := json.Marshal(msg.Data)
	if err != nil {
		log.Printf("Error marshaling chat data from user %d: %v", userID, err)
		return
	}

	var chatMessage models.ChatMessage
	if err := json.Unmarshal(chatData, &chatMessage); err != nil {
		log.Printf("Error unmarshaling chat message from user %d: %v", userID, err)
		return
	}

	// Validate the message has required fields
	if chatMessage.ChatID == 0 {
		// If missing chat ID but has recipient, try to get or create a chat
		if chatMessage.RecipientID > 0 {
			chatID, err := getOrCreateDirectChat(userID, chatMessage.RecipientID)
			if err != nil {
				log.Printf("Error creating or getting chat for user %d and recipient %d: %v",
					userID, chatMessage.RecipientID, err)

				// Send error back to client
				errorResponse := models.WebSocketMessage{
					Type: "error",
					Data: map[string]interface{}{
						"message":         "Couldn't create chat: " + err.Error(),
						"code":            "chat_creation_failed",
						"originalMessage": msg,
					},
				}

				if err := conn.WriteJSON(errorResponse); err != nil {
					log.Printf("Error sending error response to user %d: %v", userID, err)
				}
				return
			}

			// Update the message with the actual chat ID
			chatMessage.ChatID = chatID

			// Update the original message data with the new chat ID
			updatedMsg := msg
			updatedMsgData := make(map[string]interface{})

			// Extract original data
			originalData, ok := msg.Data.(map[string]interface{})
			if ok {
				for k, v := range originalData {
					updatedMsgData[k] = v
				}
			}

			// Update with new chat ID
			updatedMsgData["chatId"] = chatID
			updatedMsg.Data = updatedMsgData

			// Reserialize the updated message
			updatedRawMessage, err := json.Marshal(updatedMsg)
			if err != nil {
				log.Printf("Error updating message with chat ID: %v", err)
			} else {
				// Replace the raw message with the updated one
				rawMessage = updatedRawMessage
			}
		} else {
			log.Printf("Error: Message from user %d missing both chatId and recipientId", userID)

			// Send error back to client
			errorResponse := models.WebSocketMessage{
				Type: "error",
				Data: map[string]interface{}{
					"message": "Message must include either chatId or recipientId",
					"code":    "invalid_message",
				},
			}

			if err := conn.WriteJSON(errorResponse); err != nil {
				log.Printf("Error sending error response to user %d: %v", userID, err)
			}
			return
		}
	}

	// Ensure the sender ID matches the authenticated user
	chatMessage.SenderID = userID

	// Save the message to the database
	if err := SaveMessage(chatMessage); err != nil {
		log.Printf("Error saving message from user %d: %v", userID, err)

		// Send error response back to the sender
		errorResponse := models.WebSocketMessage{
			Type: "error",
			Data: map[string]interface{}{
				"message": err.Error(),
				"code":    "message_save_failed",
			},
		}

		if err := conn.WriteJSON(errorResponse); err != nil {
			log.Printf("Error sending error response to user %d: %v", userID, err)
		}
		return
	}

	// Echo the message back to the sender with updated fields (like ID)
	if err := conn.WriteMessage(messageType, rawMessage); err != nil {
		log.Printf("Error echoing message to sender (user %d): %v", userID, err)
	}

	// Create notification for the message
	if err := createChatNotification(chatMessage); err != nil {
		log.Printf("Error creating notification for message: %v", err)
	}

	// Send the message to the recipient if they're online
	recipientID := chatMessage.RecipientID
	socketManager.Mu.RLock()
	recipientConn, recipientOnline := socketManager.Sockets[recipientID]
	socketManager.Mu.RUnlock()

	if recipientOnline {
		log.Printf("Recipient %d is online, sending message", recipientID)
		if err := recipientConn.WriteMessage(messageType, rawMessage); err != nil {
			log.Printf("Error sending message to recipient %d: %v", recipientID, err)
		}
	} else {
		log.Printf("Recipient %d is offline, notification will be sent", recipientID)
	}
}

// Function to get or create a direct chat between two users
func getOrCreateDirectChat(userID, recipientID int) (int, error) {
	// Check if a direct chat already exists between these users
	var chatID int
	err := sqlite.DB.QueryRow(`
		SELECT c.id 
		FROM chats c
		JOIN user_chat_status ucs1 ON c.id = ucs1.chat_id AND ucs1.user_id = ?
		JOIN user_chat_status ucs2 ON c.id = ucs2.chat_id AND ucs2.user_id = ?
		WHERE c.type = 'direct'`,
		userID, recipientID,
	).Scan(&chatID)

	if err == nil {
		// Found existing chat
		return chatID, nil
	}

	if err != sql.ErrNoRows {
		// Unexpected error
		return 0, fmt.Errorf("database error: %w", err)
	}

	// No existing chat found, check if there's a follow relationship
	var followExists bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM followers 
			WHERE ((follower_id = ? AND followed_id = ?) 
			OR (follower_id = ? AND followed_id = ?))
			AND status = 'accepted'
		)`,
		userID, recipientID, recipientID, userID,
	).Scan(&followExists)

	if err != nil {
		return 0, fmt.Errorf("database error checking follow status: %w", err)
	}

	if !followExists {
		return 0, fmt.Errorf("cannot create chat: at least one user must follow the other")
	}

	// Create a new chat
	tx, err := sqlite.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert new chat
	result, err := tx.Exec(`
		INSERT INTO chats (type, created_at)
		VALUES ('direct', CURRENT_TIMESTAMP)`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get chat ID: %w", err)
	}
	chatID = int(id)

	// Add both users to the chat
	_, err = tx.Exec(`
		INSERT INTO user_chat_status (user_id, chat_id)
		VALUES (?, ?), (?, ?)`,
		userID, chatID, recipientID, chatID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to add users to chat: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return chatID, nil
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
func processGroupChatMessage(userID int, conn *websocket.Conn, messageType int, message []byte, msg models.WebSocketMessage) {
	// Parse the group chat message
	groupData, err := json.Marshal(msg.Data)
	if err != nil {
		log.Printf("Error marshaling group chat data: %v", err)
		return
	}

	var groupMessage models.GroupMessage
	if err := json.Unmarshal(groupData, &groupMessage); err != nil {
		log.Printf("Error unmarshaling group chat message: %v", err)
		return
	}

	// Ensure the user ID matches the authenticated user
	groupMessage.UserID = userID

	//find the group id from chatid
	var groupID int
	err = sqlite.DB.QueryRow(`
		SELECT id FROM groups WHERE chat_id = ?
	`, groupMessage.ChatId).Scan(&groupID)
	if err != nil {
		log.Printf("Error getting group id from chat id: %v", err)
		return
	}

	// Verify user is a member of the group
	var isMember bool
	err = sqlite.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM group_members
			WHERE group_id = ? AND user_id = ?
		)`, groupID, userID).Scan(&isMember)

	if err != nil {
		log.Printf("Error checking group membership: %v", err)
		return
	}

	if !isMember {
		// Send error back to user
		errorResponse := models.WebSocketMessage{
			Type: "error",
			Data: map[string]interface{}{
				"message": "You are not a member of this group",
				"code":    "not_group_member",
			},
		}

		if err := conn.WriteJSON(errorResponse); err != nil {
			log.Printf("Error sending error response: %v", err)
		}
		return
	}

	// Save the message
	stmt, err := sqlite.DB.Prepare(`
		INSERT INTO chat_messages (chat_id, sender_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		groupMessage.ChatId,
		groupMessage.UserID,
		groupMessage.Content,
	)
	if err != nil {
		log.Printf("Error saving group message: %v", err)
		return
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting message ID: %v", err)
	} else {
		groupMessage.ID = int(messageID)
	}

	// Get user info to include in the message
	var firstName, lastName, avatar string
	err = sqlite.DB.QueryRow(`
		SELECT first_name, last_name, avatar
		FROM users WHERE id = ?
	`, userID).Scan(&firstName, &lastName, &avatar)

	if err != nil {
		log.Printf("Error getting user info: %v", err)
	} else {
		groupMessage.UserName = firstName + " " + lastName
		groupMessage.UserAvatar = avatar
	}

	// Update the message with the user info and send to all group members
	msg.Data = groupMessage

	// Send updated message to all group members
	rows, err := sqlite.DB.Query(`
		SELECT user_id FROM group_members WHERE group_id = ?
	`, groupID)

	if err != nil {
		log.Printf("Error getting group members: %v", err)
		return
	}
	defer rows.Close()

	// Collect all member IDs
	var memberIDs []int
	for rows.Next() {
		var memberID int
		if err := rows.Scan(&memberID); err != nil {
			log.Printf("Error scanning member ID: %v", err)
			continue
		}
		memberIDs = append(memberIDs, memberID)
	}

	// First, return the created message to the sender with server-generated ID
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error returning created message to sender %d: %v", userID, err)
	}

	// Send message to all members EXCEPT the sender (to prevent duplication)
	for _, memberID := range memberIDs {
		// Skip sending to the sender as they already have the message in their local state
		if memberID == userID {
			continue
		}

		socketManager.Mu.RLock()
		memberConn, isOnline := socketManager.Sockets[memberID]
		socketManager.Mu.RUnlock()

		if isOnline {
			if err := memberConn.WriteJSON(msg); err != nil {
				log.Printf("Error sending group message to member %d: %v", memberID, err)
			}
		}
	}
	// Create notifications for offline members
	for _, memberID := range memberIDs {
		if memberID == userID {
			// Skip notification for the sender
			continue
		}

		socketManager.Mu.RLock()
		_, isOnline := socketManager.Sockets[memberID]
		socketManager.Mu.RUnlock()

		if !isOnline {
			// Create notification for offline member
			content := fmt.Sprintf("%s: %s", groupMessage.UserName, truncateMessage(groupMessage.Content))
			_, err := sqlite.DB.Exec(`
				INSERT INTO notifications (
					type, content, user_id, group_id, from_user_id, is_read, created_at
				) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			`, "group_message", content, memberID, groupID, userID, false)

			if err != nil {
				log.Printf("Error creating notification for member %d: %v", memberID, err)
			}
		}
	}
}

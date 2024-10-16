package api

import (
	"encoding/json"
	"log"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Create a socket manager
func makeSocketManager() *m.SocketManager {
	return &m.SocketManager{
		Sockets: make(map[uint64]*websocket.Conn),
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	socketManager := makeSocketManager()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userIDStr := "0" // ! should change to good way 
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	log.Println(userID)

	AddConnection(socketManager, uint64(userID), conn)

	go func() {
		defer RemoveConnection(socketManager, uint64(userID)) // Ensure cleanup
		HandleMessages(conn, uint64(userID))
	}()
}

// Handle messages for example like, notfiction, chat or groupChat and so on.
func HandleMessages(conn *websocket.Conn, userID uint64) {
	defer conn.Close()

	for {
		var connectionType m.ConnectionType

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Handle the message here
		if err := json.Unmarshal(message, &connectionType); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		switch connectionType.Type {
		case "notification":
			// Handle notification message
		case "chat":
			// Handle chat message
			SendMessage(&m.SocketManager{} ,message)
		case "groupChat":
			// Handle group chat message
		case "like":
			// Handle like message
			

		default:
			log.Printf("Unknown message type: %s", connectionType.Type)
		}
	}
}

func SendMessage(sm *m.SocketManager, message []byte) {
	var chatMessage m.Chat_message
	if err := json.Unmarshal(message, &chatMessage); err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}

	chatMessage.SenderID = 1  // ! need to change the way
	chatMessage.UserName = "sss"  // ! need to change the way
	chatMessage.CreatedAt = time.Now()
	chatMessage.RecipientID = 2  // ! need to change the way
    
	// Insert the message into the database
	query := `INSERT INTO chat_messages (sender_id, recipient_id, content, created_at) VALUES (?, ?, ?, ?)`
	_, err := sqlite.DB.Exec(query, chatMessage.SenderID, chatMessage.RecipientID, chatMessage.Content, chatMessage.CreatedAt)
	if err != nil {
		log.Println("Error inserting message:", err)
		return
	}

    //send the message to the client using userid
    responseMessage, err := json.Marshal(chatMessage)
	if err != nil {
		log.Println("Error marshalling chat message for sending:", err)
		return
	}

	// Lock the SocketManager while sending the message
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	// Send the message to the specific recipient
	if conn, exists := sm.Sockets[uint64(chatMessage.RecipientID)]; exists {
		if err := conn.WriteMessage(websocket.TextMessage, responseMessage); err != nil {
			log.Printf("Error sending message to user %d: %v", uint64(chatMessage.RecipientID), err)
			RemoveConnection(sm, uint64(chatMessage.RecipientID)) 
		}
	} else {
		log.Printf("No active connection for recipient ID %d", uint64(chatMessage.RecipientID))
	}
}



func AddConnection(sm *m.SocketManager, userID uint64, conn *websocket.Conn) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	sm.Sockets[userID] = conn  
	log.Printf("Added new connection for user ID %d ", userID)
}

func RemoveConnection(sm *m.SocketManager, userID uint64) { 
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if conn, exists := sm.Sockets[userID]; exists {
		conn.Close()
		delete(sm.Sockets, userID) 
		log.Printf("Removed connection for user ID %d", userID)
	}
}

func Broadcast(sm *m.SocketManager, message []byte) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	for userID, conn := range sm.Sockets {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting to user ID %d: %v", userID, err)
			RemoveConnection(sm, userID) 
		}
	}
}

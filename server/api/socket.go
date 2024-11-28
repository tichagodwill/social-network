package api

import (
	"log"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"sync"

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
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("WebSocket session error: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID uint64
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("WebSocket database error: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Check if connection already exists
	socketManager.Mu.RLock()
	existingConn, exists := socketManager.Sockets[userID]
	socketManager.Mu.RUnlock()

	if exists {
		// Close existing connection
		existingConn.Close()
		socketManager.Mu.Lock()
		delete(socketManager.Sockets, userID)
		socketManager.Mu.Unlock()
		log.Printf("Closed existing WebSocket connection for user %s", username)
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Store connection
	socketManager.Mu.Lock()
	socketManager.Sockets[userID] = conn
	socketManager.Mu.Unlock()

	log.Printf("New WebSocket connection for user %s (ID: %d)", username, userID)

	// Clean up on disconnect
	defer func() {
		socketManager.Mu.Lock()
		if conn, ok := socketManager.Sockets[userID]; ok {
			conn.Close()
			delete(socketManager.Sockets, userID)
		}
		socketManager.Mu.Unlock()
		log.Printf("WebSocket connection closed for user %s", username)
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

		// Echo the message back (for testing)
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// SendNotification sends a notification to a specific user
func SendNotification(userID uint64, notification interface{}) {
	socketManager.Mu.RLock()
	conn, exists := socketManager.Sockets[userID]
	socketManager.Mu.RUnlock()

	if exists {
		if err := conn.WriteJSON(notification); err != nil {
			log.Printf("Error sending notification to user %d: %v", userID, err)
			socketManager.Mu.Lock()
			conn.Close()
			delete(socketManager.Sockets, userID)
			socketManager.Mu.Unlock()
		}
	}
}

package api

import (
	// "database/sql"
	"encoding/json"
	"log"
	"net/http"
	m "social-network/models"
	"strconv"
	"sync/atomic"
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
        SocketCounter: atomic.Uint64{},
        Sockets:       make(map[uint64]*websocket.Conn),
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
        HandelMessages(conn, uint64(userID))
    }()
}

// Handle messages for example like, notfiction, chat or groupChat and so on.
func HandelMessages(conn *websocket.Conn, userID uint64) {
    
    defer conn.Close()

    for{
        var ConnectionType m.ConnectionType
        
        _, Message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message:", err)
            break
        }
        // Handle the message here
        if err := json.Unmarshal(Message, &ConnectionType); err != nil {
            log.Println("Error unmarshalling message:", err)
            continue
        }

        switch ConnectionType.Type {
        case "notification":
            // Handle notification message
        case "chat":
            // Handle chat message
            SendMessage(conn, Message)
        case "groupChat":
            // Handle group chat message
        default:
            // Handle other message types
            log.Printf("Unknown message type: %s", ConnectionType.Type)
        }
    }
}

func SendMessage(conn *websocket.Conn, message []byte) {
    var ChatMessage m.Chat_message
    if err := json.Unmarshal(message, &ChatMessage); err != nil {
        log.Println("Error unmarshalling message:", err)
        return
    }

    ChatMessage.SenderID = 1  // ! need to change the way
    ChatMessage.UserName = "sss"  // ! need to change the way 
    ChatMessage.CreatedAt = time.Now()
    ChatMessage.RecipientID = 2  // ! need to change the way

    // ? query to insert the message to the database
    /*query := `INSERT INTO chat_messages (sender_id, recipient_id, content, created_at) VALUES (?, ?, ?, ?)`

    err := sql.DB.Exec(query, ChatMessage.SenderID, ChatMessage.RecipientID, ChatMessage.Content, ChatMessage.CreatedAt)
    if err != nil {
        log.Println("Error inserting message into database:", err)
        return
    }*/


}


func AddConnection(sm *m.SocketManager, userID uint64, conn *websocket.Conn) {
    sm.Mu.Lock()
    defer sm.Mu.Unlock()

    socketID := sm.SocketCounter.Add(1)
    sm.Sockets[socketID] = conn
    log.Printf("Added new connection for user ID %d with socket ID %d", userID, socketID)
}

func RemoveConnection(sm *m.SocketManager, socketID uint64) {
    sm.Mu.Lock()
    defer sm.Mu.Unlock()

    if conn, exists := sm.Sockets[socketID]; exists {
        conn.Close()
        delete(sm.Sockets, socketID)
        log.Printf("Removed connection with socket ID %d", socketID)
    }
}

func Broadcast(sm *m.SocketManager, message []byte) {
    sm.Mu.Lock()
    defer sm.Mu.Unlock()

    for socketID, conn := range sm.Sockets {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Printf("Error broadcasting to socket ID %d: %v", socketID, err)
            RemoveConnection(sm, socketID)
        }
    }
}


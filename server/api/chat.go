package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
	"time"
)

type DirectChatRequest struct {
	UserId int `json:"userId"`
}

func CreateOrGetDirectChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DirectChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var currentUser struct {
		ID int
	}
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&currentUser.ID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is trying to create a chat with themselves
	if currentUser.ID == req.UserId {
		http.Error(w, "Cannot start a chat with yourself", http.StatusBadRequest)
		return
	}

	// Check if the user exists
	var userExists bool
	err = sqlite.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE id = ?)", req.UserId).Scan(&userExists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !userExists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if there's a follow relationship between users (either user following the other)
	var followExists bool
	err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM followers 
            WHERE ((follower_id = ? AND followed_id = ?) 
            OR (follower_id = ? AND followed_id = ?))
            AND status = 'accepted'
        )`,
		currentUser.ID, req.UserId, req.UserId, currentUser.ID,
	).Scan(&followExists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !followExists {
		http.Error(w, "Cannot start chat: at least one user must follow the other", http.StatusForbidden)
		return
	}

	// Check if a direct chat already exists between these users
	var chatID int
	err = sqlite.DB.QueryRow(`
       SELECT c.id 
       FROM chats c
       JOIN user_chat_status ucs1 ON c.id = ucs1.chat_id AND ucs1.user_id = ?
       JOIN user_chat_status ucs2 ON c.id = ucs2.chat_id AND ucs2.user_id = ?
       WHERE c.type = 'direct'`,
		currentUser.ID, req.UserId,
	).Scan(&chatID)

	if err != nil {
		if err != sql.ErrNoRows {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// No existing chat, create a new one
		result, err := sqlite.DB.Exec(`
          INSERT INTO chats (type, created_at)
          VALUES ('direct', CURRENT_TIMESTAMP)`,
		)
		if err != nil {
			http.Error(w, "Failed to create chat", http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get chat ID", http.StatusInternalServerError)
			return
		}
		chatID = int(id)

		// Add both users to the chat
		_, err = sqlite.DB.Exec(`
          INSERT INTO user_chat_status (user_id, chat_id)
          VALUES (?, ?), (?, ?)`,
			currentUser.ID, chatID, req.UserId, chatID,
		)
		if err != nil {
			http.Error(w, "Failed to add users to chat", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": chatID,
	})
}

// GetUserChats returns all chats for the authenticated user
func GetUserChats(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userId int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get all chats for the user
	rows, err := sqlite.DB.Query(`
        SELECT
            c.id,
            c.type,
            COALESCE(
                (SELECT COUNT(*) FROM chat_messages m
                 JOIN user_chat_status ucs ON ucs.chat_id = m.chat_id AND ucs.user_id = ?
                 WHERE m.chat_id = c.id
                 AND (m.created_at > ucs.last_read_message_id OR ucs.last_read_message_id IS NULL)),
                0
            ) as unread_count,
            (SELECT content FROM chat_messages
             WHERE chat_id = c.id
             ORDER BY created_at DESC LIMIT 1) as last_message,
            (SELECT created_at FROM chat_messages
             WHERE chat_id = c.id
             ORDER BY created_at DESC LIMIT 1) as last_message_time
        FROM chats c
        JOIN user_chat_status ucs ON c.id = ucs.chat_id
        WHERE ucs.user_id = ?
        ORDER BY last_message_time DESC NULLS LAST
    `, userId, userId)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error getting user chats: %v", err)
		return
	}
	defer rows.Close()

	var chats []map[string]interface{}
	for rows.Next() {
		var chat struct {
			ID              int       `json:"id"`
			Type            string    `json:"type"`
			UnreadCount     int       `json:"unread_count"`
			LastMessage     string    `json:"last_message"`
			LastMessageTime time.Time `json:"last_message_time"`
		}

		var lastMessage, lastMessageTime sql.NullString

		if err := rows.Scan(
			&chat.ID,
			&chat.Type,
			&chat.UnreadCount,
			&lastMessage,
			&lastMessageTime,
		); err != nil {
			log.Printf("Error scanning chat: %v", err)
			continue
		}

		if lastMessage.Valid {
			chat.LastMessage = lastMessage.String
		}

		if lastMessageTime.Valid {
			chat.LastMessageTime, _ = time.Parse(time.RFC3339, lastMessageTime.String)
		}

		chatItem := map[string]interface{}{
			"id":           chat.ID,
			"type":         chat.Type,
			"unread_count": chat.UnreadCount,
		}

		if lastMessage.Valid {
			chatItem["last_message"] = lastMessage.String
		}

		if lastMessageTime.Valid {
			chatItem["last_message_time"] = lastMessageTime.String
		}

		// For direct chats, add the other user's info
		if chat.Type == "direct" {
			var otherUser struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Username  string `json:"username"`
				Avatar    string `json:"avatar"`
			}

			err := sqlite.DB.QueryRow(`
                SELECT
                    u.id,
                    u.first_name,
                    u.last_name,
                    u.username,
                    u.avatar
                FROM users u
                JOIN user_chat_status ucs ON u.id = ucs.user_id
                WHERE ucs.chat_id = ? AND u.id != ?
                LIMIT 1
            `, chat.ID, userId).Scan(
				&otherUser.ID,
				&otherUser.FirstName,
				&otherUser.LastName,
				&otherUser.Username,
				&otherUser.Avatar,
			)

			if err == nil {
				chatItem["id"] = chat.ID
				chatItem["participant_id"] = otherUser.ID
				chatItem["first_name"] = otherUser.FirstName
				chatItem["last_name"] = otherUser.LastName
				chatItem["username"] = otherUser.Username
				chatItem["avatar"] = otherUser.Avatar
			}
		} else {
			// For group chats, add the group info
			var groupInfo struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Avatar      string `json:"avatar"`
			}

			err := sqlite.DB.QueryRow(`
                SELECT
                    title as name,
                    description,
                    '' as avatar
                FROM groups g
                WHERE g.id = ?
                LIMIT 1
            `, chat.ID).Scan(
				&groupInfo.Name,
				&groupInfo.Description,
				&groupInfo.Avatar,
			)

			if err == nil {
				chatItem["name"] = groupInfo.Name
				chatItem["description"] = groupInfo.Description
				chatItem["avatar"] = groupInfo.Avatar
			}
		}

		chats = append(chats, chatItem)
	}

	if err := json.NewEncoder(w).Encode(chats); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetChatParticipants returns the participants of a chat
func GetChatParticipants(w http.ResponseWriter, r *http.Request) {
	// Extract chat ID from URL path parameters
	chatIdStr := r.PathValue("chatId")

	chatId, err := strconv.Atoi(chatIdStr)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	// Ensure the authenticated user is a participant of the chat
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var authUserId int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&authUserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is a participant of the chat
	var isParticipant bool
	err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM user_chat_status 
            WHERE chat_id = ? AND user_id = ?
        )`, chatId, authUserId).Scan(&isParticipant)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error checking chat participation: %v", err)
		return
	}

	if !isParticipant {
		http.Error(w, "Unauthorized: not a chat participant", http.StatusUnauthorized)
		return
	}

	// Get all participants
	rows, err := sqlite.DB.Query(`
        SELECT 
            u.id,
            u.username,
            u.first_name,
            u.last_name,
            u.avatar
        FROM users u
        JOIN user_chat_status ucs ON u.id = ucs.user_id
        WHERE ucs.chat_id = ?
        ORDER BY u.username
    `, chatId)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error getting chat participants: %v", err)
		return
	}
	defer rows.Close()

	var participants []map[string]interface{}
	for rows.Next() {
		var participant struct {
			ID        int    `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Avatar    string `json:"avatar"`
		}

		if err := rows.Scan(
			&participant.ID,
			&participant.Username,
			&participant.FirstName,
			&participant.LastName,
			&participant.Avatar,
		); err != nil {
			log.Printf("Error scanning participant: %v", err)
			continue
		}

		participants = append(participants, map[string]interface{}{
			"id":         participant.ID,
			"username":   participant.Username,
			"first_name": participant.FirstName,
			"last_name":  participant.LastName,
			"avatar":     participant.Avatar,
		})
	}

	if err := json.NewEncoder(w).Encode(participants); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

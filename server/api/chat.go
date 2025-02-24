package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
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

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
	// Only POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req DirectChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user details from username
	var currentUser struct {
		ID int
	}
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&currentUser.ID)
	if err != nil {
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

	// Check if a chat already exists between these users
	var chatID int
	err = sqlite.DB.QueryRow(`
		SELECT id FROM chats 
		WHERE ((user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?))
		AND type = 'direct'`,
		currentUser.ID, req.UserId, req.UserId, currentUser.ID,
	).Scan(&chatID)

	if err != nil {
		if err != sql.ErrNoRows {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		result, err := sqlite.DB.Exec(`
			INSERT INTO chats (user1_id, user2_id, type, created_at)
			VALUES (?, ?, 'direct', CURRENT_TIMESTAMP)`,
			currentUser.ID, req.UserId,
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
	}

	// Return the chat ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": chatID,
	})
}

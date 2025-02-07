package api

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

type DirectChatRequest struct {
	UserId int `json:"userId"`
}

func CheckFollowStatus(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "No follow relationship exists", http.StatusForbidden)
		return
	}

	// Return success if a follow relationship exists
	w.WriteHeader(http.StatusOK)
}

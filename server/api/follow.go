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

// FollowUser handles follow requests
func FollowUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		UserToFollowID int `json:"userToFollowId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get follower's ID
	var followerID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&followerID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is trying to follow themselves
	if followerID == req.UserToFollowID {
		http.Error(w, "Cannot follow yourself", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if target user exists and get their privacy setting
	var isPrivate bool
	err = tx.QueryRow(`
		SELECT is_private 
		FROM users  
		WHERE id = ?`, req.UserToFollowID).Scan(&isPrivate)
	if err == sql.ErrNoRows {
		http.Error(w, "User to follow not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if already following
	var existingStatus string
	err = tx.QueryRow(`
		SELECT status 
		FROM followers 
		WHERE follower_id = ? AND followed_id = ?`,
		followerID, req.UserToFollowID).Scan(&existingStatus)

	if err == nil {
		// Already exists
		if existingStatus == "accepted" {
			http.Error(w, "Already following this user", http.StatusBadRequest)
			return
		}
		if existingStatus == "pending" {
			http.Error(w, "Follow request already pending", http.StatusBadRequest)
			return
		}
	}

	// Determine initial status based on privacy setting
	initialStatus := "pending"
	if !isPrivate {
		initialStatus = "accepted"
	}

	// Insert follow relationship
	_, err = tx.Exec(`
		INSERT INTO followers (follower_id, followed_id, status)
		VALUES (?, ?, ?)
		ON CONFLICT(follower_id, followed_id) 
		DO UPDATE SET status = ?`,
		followerID, req.UserToFollowID, initialStatus, initialStatus)
	if err != nil {
		http.Error(w, "Failed to create follow relationship", http.StatusInternalServerError)
		return
	}

	// Create notification for private accounts
	if isPrivate {
		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, type, content, from_user_id)
			VALUES (?, 'follow_request', 'wants to follow you', ?)`,
			req.UserToFollowID, followerID)
		if err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to complete follow request", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  initialStatus,
		"message": "Follow request processed successfully",
	})
}

// UnfollowUser handles unfollow requests
func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		UserToUnfollowID int `json:"userToUnfollowId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var followerID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&followerID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	result, err := sqlite.DB.Exec(`
		DELETE FROM followers 
		WHERE follower_id = ? AND followed_id = ? AND status = 'accepted'`,
		followerID, req.UserToUnfollowID)
	if err != nil {
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Not following this user", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully unfollowed user",
	})
}

func FollowStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the current user's username from the session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the current user's ID from the database
	var currentUserID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&currentUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	//get userid from body
	var req struct {
		FollowedId int `json:"followedId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if the current user is following the specified user
	var isFollowing bool
	var pendeingRequest bool

	err = sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'accepted')",
		currentUserID, req.FollowedId).Scan(&isFollowing)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to check follow status",
		})
		return
	}

	err = sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'pending')",
		currentUserID, req.FollowedId).Scan(&pendeingRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to check follow status",
		})
		return
	}

	//return the status
	json.NewEncoder(w).Encode(map[string]bool{
		"isFollowing":       isFollowing,
		"hasPendingRequest": pendeingRequest,
	})

}

// HandleFollowRequest handles accepting or rejecting follow requests
func HandleFollowRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// We don't need the username for this function since we validate using the request ID
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	var req struct {
		RequestID int  `json:"requestId"`
		Action    bool `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate action value
	var action string
	if req.Action {
		action = "accepted"
	} else {
		// Remove the record if action is rejected
		_, err := sqlite.DB.Exec("DELETE FROM followers WHERE id = ? AND followed_id = ?", req.RequestID, userID)
		if err != nil {
			http.Error(w, "Failed to remove follow request", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Follow request rejected and removed successfully",
		})
		return
	}

	// Check if the user is the owner of the request
	var isOwner bool
	err = sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE id = ? AND followed_id = ?)", req.RequestID, userID).Scan(&isOwner)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !isOwner {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update follow request status
	result, err := tx.Exec(`
		UPDATE followers 
		SET status = ? 
		WHERE id = ? AND status = 'pending' AND followed_id = ?`,
		action, req.RequestID, userID)
	if err != nil {
		http.Error(w, "Failed to process follow request", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Follow request not found or already processed", http.StatusBadRequest)
		return
	}

	// Create notification for the requester
	var followRequest struct {
		FollowerID int
		FollowedID int
	}
	err = tx.QueryRow(`
		SELECT follower_id, followed_id 
		FROM followers 
		WHERE id = ?`, req.RequestID).Scan(&followRequest.FollowerID, &followRequest.FollowedID)
	if err == nil {
		_, err = tx.Exec(`
			INSERT INTO notifications (user_id, type, content, from_user_id)
			VALUES (?, ?, ?, ?)`,
			followRequest.FollowerID,
			"follow_request_"+action,
			"has "+action+" your follow request",
			followRequest.FollowedID)
		if err != nil {
			log.Printf("Failed to create notification: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to complete request", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Follow request " + action + " successfully",
	})
}

// GetFollowers returns the list of followers or following users for a given user
func GetFollowers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get userID from URL path
	userIDStr := r.PathValue("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Query to get followers with their status
	rows, err := sqlite.DB.Query(`
		SELECT 
			f.id,
			f.follower_id,
			f.followed_id,
			f.status,
			f.created_at,
			u.username,
			u.first_name,
			u.last_name,
			u.avatar
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.followed_id = ? AND f.status IN ('accepted')
	`, userID)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Failed to get followers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FollowerInfo struct {
		ID        int       `json:"id"`
		UserID    int       `json:"userId"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		Username  string    `json:"username"`
		FirstName string    `json:"firstName"`
		LastName  string    `json:"lastName"`
		Avatar    string    `json:"avatar"`
	}

	var followers []FollowerInfo
	for rows.Next() {
		var f FollowerInfo
		var avatar sql.NullString
		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&userID, // followed_id
			&f.Status,
			&f.CreatedAt,
			&f.Username,
			&f.FirstName,
			&f.LastName,
			&avatar,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		f.Avatar = avatar.String // Will be empty string if NULL
		followers = append(followers, f)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		http.Error(w, "Error retrieving followers", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(followers)
}

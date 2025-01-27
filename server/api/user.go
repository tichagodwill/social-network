package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
)

type Profile struct {
	Image       string `json:"image"`
	Description string `json:"description"`
	Privacy     bool   `json:"privacy"`
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Parse the JSON data into the Profile struct
	var profile Profile
	err = json.Unmarshal(body, &profile)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Insert the post into the database
	result, err := sqlite.DB.Exec(
		"update users set avatar = ?, about_me = ?, is_private = ? where id = ?",
		profile.Image, profile.Description, profile.Privacy, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create post",
		})
		log.Printf("Error creating post: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Follow request not found or already processed", http.StatusBadRequest)
		return
	}

}

func UserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIdString := r.PathValue("userID")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Get user ID
	var currentUserID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&currentUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	// Convert id to number
	userID, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(w, "Error processing user ID", http.StatusBadRequest)
		return
	}

	var canView bool = true
	if currentUserID != userID {
		// Check if the user is private
		var isPrivate bool
		err = sqlite.DB.QueryRow("SELECT is_private FROM users WHERE id = ?", userID).Scan(&isPrivate)
		if err != nil {
			http.Error(w, "Error getting user privacy settings", http.StatusInternalServerError)
			return
		}
		//check if currentUserID follows the user
		if isPrivate {

			// Check if the user is following the user by the status of the follow request
			var followStatus string
			err = sqlite.DB.QueryRow("SELECT status FROM follows WHERE follower_id = ? AND followee_id = ?", currentUserID, userID).Scan(&followStatus)
			if err != nil {
				canView = false
			}
			if followStatus != "accepted" {
				canView = false
			}
		}
	}

	var userInfo models.User
	var avatar sql.NullString // Handle nullable avatar
	var aboutMe sql.NullString
	if err := sqlite.DB.QueryRow(
		"SELECT id, email, username, first_name, last_name, date_of_birth, avatar, about_me, is_private, created_at FROM users WHERE id = ?",
		userID).Scan(
		&userInfo.ID,
		&userInfo.Email,
		&userInfo.Username,
		&userInfo.FirstName,
		&userInfo.LastName,
		&userInfo.DateOfBirth,
		&avatar,
		&aboutMe,
		&userInfo.IsPrivate,
		&userInfo.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting user info: %v", err)
		return
	}

	// Check if the avatar is valid (i.e., not null)
	if avatar.Valid {
		userInfo.Avatar = avatar.String // Use the value if it's valid
	}

	if aboutMe.Valid {
		userInfo.AboutMe = aboutMe.String
	}

	// If the user cannot view the full profile, restrict the details
	if !canView {
		userInfo.Email = ""
		userInfo.FirstName = ""
		userInfo.LastName = ""
		userInfo.DateOfBirth = nil
		userInfo.AboutMe = ""
		userInfo.CreatedAt = nil
	}

	// Encode the userInfo as JSON and send it as a response
	if err := json.NewEncoder(w).Encode(&userInfo); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

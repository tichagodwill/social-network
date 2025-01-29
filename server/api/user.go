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
			err = sqlite.DB.QueryRow("SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?", currentUserID, userID).Scan(&followStatus)
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

	var following []models.Followers
	var followers []models.Followers
	var requests []models.Followers

	if currentUserID == userID {
		//get the followers that follows the user who are pending requests
		rows, err := sqlite.DB.Query("SELECT f.id, users.id, users.username, users.avatar, users.first_name, users.last_name FROM users INNER JOIN followers f ON users.id = f.follower_id WHERE f.followed_id = ? AND f.status = 'pending'", userID)
		if err != nil {
			http.Error(w, "Error getting followers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var request models.Followers
			var avatar sql.NullString
			err = rows.Scan(&request.ID, &request.UserId, &request.Username, &avatar, &request.FirstName, &request.LastName)
			if err != nil {
				http.Error(w, "Error scanning following", http.StatusInternalServerError)
				return
			}
			if avatar.Valid {
				request.Avatar = avatar.String
			}
			requests = append(requests, request)
		}
	}

	if canView {
		//get the followers that follows the user
		rows, err := sqlite.DB.Query("SELECT users.id, users.username, users.avatar, users.first_name, users.last_name FROM users INNER JOIN followers f ON users.id = f.follower_id WHERE f.followed_id = ? AND f.status = 'accepted'", userID)
		if err != nil {
			http.Error(w, "Error getting followers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var follower models.Followers
			var avatar sql.NullString
			err = rows.Scan(&follower.UserId, &follower.Username, &avatar, &follower.FirstName, &follower.LastName)
			if err != nil {
				http.Error(w, "Error scanning following", http.StatusInternalServerError)
				return
			}
			if avatar.Valid {
				follower.Avatar = avatar.String
			}
			followers = append(followers, follower)
		}

		//get the users that the user follows
		rows, err = sqlite.DB.Query("SELECT users.id, users.username, users.avatar, users.first_name, users.last_name  FROM users INNER JOIN followers f ON users.id = f.followed_id WHERE f.follower_id = ? AND f.status = 'accepted'", userID)
		if err != nil {
			http.Error(w, "Error getting following", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var follow models.Followers
			var avatar sql.NullString
			err = rows.Scan(&follow.UserId, &follow.Username, &avatar, &follow.FirstName, &follow.LastName)
			if err != nil {
				http.Error(w, "Error scanning following", http.StatusInternalServerError)
				return
			}
			if avatar.Valid {
				follow.Avatar = avatar.String
			}
			following = append(following, follow)
		}
	}

	// Encode the userInfo and followers/following JSON and send it as a response
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"user":      userInfo,
		"followers": followers,
		"following": following,
		"requests":  requests,
	}); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}

	//if err := json.NewEncoder(w).Encode(&userInfo); err != nil {
	//	http.Error(w, "Error sending data", http.StatusInternalServerError)
	//}
}

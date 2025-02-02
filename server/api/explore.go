package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

// GetExplore get all the users from the database for explore page
func GetExplore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Get search query from URL parameters
	var requestBody struct {
		Search string `json:"search"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	searchQuery := requestBody.Search

	// Get all users except the current user, optionally filtering by search query
	var rows *sql.Rows
	if searchQuery != "" {
		rows, err = sqlite.DB.Query("SELECT id, username, avatar, is_private FROM users WHERE id != ? AND username LIKE ?", userID, "%"+searchQuery+"%")
	} else {
		rows, err = sqlite.DB.Query("SELECT id, username, avatar, is_private FROM users WHERE id != ?", userID)
	}
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []m.User
	for rows.Next() {
		var user m.User
		var avatar sql.NullString
		err = rows.Scan(&user.ID, &user.Username, &avatar, &user.IsPrivate)
		if err != nil {
			http.Error(w, "Failed to get users", http.StatusInternalServerError)
			return
		}
		if avatar.Valid {
			user.Avatar = avatar.String
		} else {
			user.Avatar = ""
		}
		users = append(users, user)
	}

	// return the users
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"social-network/models"
	"social-network/pkg/db/sqlite"
	"strconv"
)

func GetContact(w http.ResponseWriter, r *http.Request) {
	userIdString := r.PathValue("userID")
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var users []models.User

	// Get users that either follow you or you follow them
	rows, err := sqlite.DB.Query(`
    SELECT DISTINCT
        u.id, u.Email, u.Username, u.first_name, u.last_name, 
        u.date_of_birth, u.Avatar, u.about_me, u.is_private, u.created_at
    FROM users u
    INNER JOIN followers f ON 
        (f.follower_id = u.id AND f.followed_id = ?) OR 
        (f.follower_id = ? AND f.followed_id = u.id)
    WHERE f.status = 'accepted'
    `, userId, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No contacts found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting contacts: %v", err)
		return
	}

	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.DateOfBirth, &u.Avatar, &u.AboutMe, &u.IsPrivate, &u.CreatedAt)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error scanning contact: %v", err)
			return
		}

		users = append(users, u)
	}

	if err := json.NewEncoder(w).Encode(&users); err != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

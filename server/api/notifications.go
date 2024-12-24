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

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Session error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Get user ID from username
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Database error getting user ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	// Get notifications for the user with error handling
	rows, err := sqlite.DB.Query(`
        SELECT 
            n.id,
            n.type,
            n.content,
            n.created_at,
            n.is_read,
            n.from_user_id,
            n.group_id,
            n.user_id
        FROM notifications n
        WHERE n.user_id = ?
        ORDER BY n.created_at DESC
        LIMIT 50`, userID)
	if err != nil {
		log.Printf("Database error querying notifications: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch notifications",
		})
		return
	}
	defer rows.Close()

	notifications := []map[string]interface{}{}
	for rows.Next() {
		var n struct {
			ID         int       `json:"id"`
			Type       string    `json:"type"`
			Content    string    `json:"content"`
			CreatedAt  time.Time `json:"created_at"`
			IsRead     bool      `json:"is_read"`
			FromUserID sql.NullInt64
			GroupID    sql.NullInt64
			UserID     int
		}

		if err := rows.Scan(
			&n.ID,
			&n.Type,
			&n.Content,
			&n.CreatedAt,
			&n.IsRead,
			&n.FromUserID,
			&n.GroupID,
			&n.UserID,
		); err != nil {
			log.Printf("Error scanning notification row: %v", err)
			continue
		}

		notification := map[string]interface{}{
			"id":        n.ID,
			"type":      n.Type,
			"content":   n.Content,
			"createdAt": n.CreatedAt.Format(time.RFC3339),
			"isRead":    n.IsRead,
			"userId":    n.UserID,
		}

		if n.FromUserID.Valid {
			notification["fromUserId"] = n.FromUserID.Int64
		}
		if n.GroupID.Valid {
			notification["groupId"] = n.GroupID.Int64
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating notifications: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error processing notifications",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}

func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	notificationIdString := r.PathValue("id")
	notificationId, err := strconv.Atoi(notificationIdString)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	markQuery := `
UPDATE notifications
SET is_read = 1
WHERE id = ?
`

	_, err = sqlite.DB.Exec(markQuery, notificationId)
	if err != nil {
		log.Fatal("[MarkNotificationAsRead] Error updating notification:", err)
	}
}

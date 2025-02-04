package api

import (
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
)

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	rows, err := sqlite.DB.Query(`
    SELECT 
        n.id,
        n.user_id,
        n.content,
        n.from_user_id,
        n.is_read,
        n.created_at,
        n.group_id,
        g.title AS group_title
    FROM notifications n
    LEFT JOIN groups g ON n.group_id = g.id
    WHERE n.user_id = ? AND n.is_read = false
    ORDER BY n.created_at DESC`, userID)
	if err != nil {
		log.Printf("Error fetching notifications: %v", err)
		sendJSONError(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var notification struct {
			ID         int64   `json:"id"`
			ToUserID   int64   `json:"to_user_id"`
			Content    string  `json:"content"`
			FromUserID *int64  `json:"from_user_id"`
			Read       bool    `json:"read"`
			CreatedAt  string  `json:"created_at"`
			GroupID    *int64  `json:"group_id"`
			GroupTitle *string `json:"group_title"`
		}

		if err := rows.Scan(
			&notification.ID,
			&notification.ToUserID,
			&notification.Content,
			&notification.FromUserID,
			&notification.Read,
			&notification.CreatedAt,
			&notification.GroupID,
			&notification.GroupTitle,
		); err != nil {
			log.Printf("Error scanning notification: %v", err)
			continue
		}

		notifications = append(notifications, map[string]interface{}{
			"id":         notification.ID,
			"toUserId":   notification.ToUserID,
			"content":    notification.Content,
			"fromUserId": notification.FromUserID,
			"read":       notification.Read,
			"createdAt":  notification.CreatedAt,
			"groupId":    notification.GroupID,
			"groupTitle": notification.GroupTitle,
		})
	}

	if notifications == nil {
		notifications = make([]map[string]interface{}, 0)
	}

	sendJSONResponse(w, http.StatusOK, notifications)
}

func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	notificationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		sendJSONError(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	result, err := sqlite.DB.Exec(`
        UPDATE notifications
        SET is_read = true
        WHERE id = ? AND user_id = ?`,
		notificationID, userID)
	if err != nil {
		log.Printf("Error marking notification as read: %v", err)
		sendJSONError(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		sendJSONError(w, "Failed to verify update", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		sendJSONError(w, "Notification not found or not owned by user", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Notification marked as read",
	})
}

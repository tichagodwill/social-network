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
            n.type,
            n.content,
            n.group_id,
            n.is_read,
            n.created_at,
            n.from_user_id,
            g.title as group_name,
            COALESCE(gm.role, '') as user_role
        FROM notifications n
        LEFT JOIN groups g ON n.group_id = g.id
        LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.user_id = n.user_id
        WHERE n.user_id = ?
        ORDER BY n.created_at DESC`,
		userID)
	if err != nil {
		log.Printf("Error fetching notifications: %v", err)
		sendJSONError(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var notification struct {
			ID         int64  `json:"id"`
			UserID     int64  `json:"user_id"`
			Type       string `json:"type"`
			Content    string `json:"content"`
			GroupID    *int64 `json:"group_id"`
			IsRead     bool   `json:"is_read"`
			CreatedAt  string `json:"created_at"`
			FromUserID *int64 `json:"from_user_id"`
			GroupName  string `json:"group_name"`
			UserRole   string `json:"user_role"`
		}
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Content,
			&notification.GroupID,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.FromUserID,
			&notification.GroupName,
			&notification.UserRole,
		); err != nil {
			log.Printf("Error scanning notification: %v", err)
			continue
		}
		notifications = append(notifications, map[string]interface{}{
			"id":         notification.ID,
			"userId":     notification.UserID,
			"type":       notification.Type,
			"content":    notification.Content,
			"groupId":    notification.GroupID,
			"isRead":     notification.IsRead,
			"createdAt":  notification.CreatedAt,
			"fromUserId": notification.FromUserID,
			"groupName":  notification.GroupName,
			"userRole":   notification.UserRole,
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

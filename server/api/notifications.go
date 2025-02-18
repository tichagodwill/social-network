package api

import (
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/util"
	"strconv"
	"time"
)

type Notification struct {
	ID           int       `json:"id"`
	Type         string    `json:"type"`
	Content      string    `json:"content"`
	UserID       int       `json:"user_id"`
	GroupID      *int      `json:"group_id,omitempty"`
	InvitationID *int      `json:"invitation_id,omitempty"`
	FromUserID   *int      `json:"from_user_id,omitempty"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
	UserRole     string    `json:"user_role,omitempty"`
	IsProcessed  bool      `json:"is_processed"`
}

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
            n.type,
            n.content,
            n.user_id,
            n.group_id,
            n.invitation_id,
            n.from_user_id,
            n.is_read,
            n.created_at,
            COALESCE(gm.role, '') as user_role,
            CASE 
                WHEN gi.status != 'pending' OR gi.status IS NULL THEN true
                ELSE false
            END as is_processed
        FROM notifications n
        LEFT JOIN group_members gm ON n.group_id = gm.group_id AND gm.user_id = n.user_id
        LEFT JOIN group_invitations gi ON n.invitation_id = gi.id
        WHERE n.user_id = ?
        ORDER BY n.created_at DESC`, userID)
	if err != nil {
		log.Printf("Error fetching notifications: %v", err)
		sendJSONError(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Content,
			&notification.UserID,
			&notification.GroupID,
			&notification.InvitationID,
			&notification.FromUserID,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.UserRole,
			&notification.IsProcessed,
		); err != nil {
			log.Printf("Error scanning notification: %v", err)
			continue
		}

		log.Printf("Processing notification: %+v", notification)

		notifications = append(notifications, map[string]interface{}{
			"id":           notification.ID,
			"type":         notification.Type,
			"content":      notification.Content,
			"groupId":      notification.GroupID,
			"invitationId": notification.InvitationID,
			"userId":       notification.UserID,
			"fromUserId":   notification.FromUserID,
			"isRead":       notification.IsRead,
			"createdAt":    notification.CreatedAt.Format(time.RFC3339),
			"userRole":     notification.UserRole,
			"isProcessed":  notification.IsProcessed,
		})
	}

	log.Printf("Sending notifications: %+v", notifications)

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

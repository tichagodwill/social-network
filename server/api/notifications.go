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

	// Get current user from session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID
	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Fetch notifications
	rows, err := sqlite.DB.Query(`
        SELECT 
            n.id,
            n.user_id,
            n.type,
            n.content,
            n.group_id,
            n.invitation_id,
            n.read,
            n.created_at,
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
			ID           int64  `json:"id"`
			UserID       int64  `json:"user_id"`
			Type         string `json:"type"`
			Content      string `json:"content"`
			GroupID      int64  `json:"group_id"`
			InvitationID int64  `json:"invitation_id"`
			Read         bool   `json:"read"`
			CreatedAt    string `json:"created_at"`
			GroupName    string `json:"group_name"`
			UserRole     string `json:"user_role"`
		}
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Type, &notification.Content, &notification.GroupID, &notification.InvitationID, &notification.Read, &notification.CreatedAt, &notification.GroupName, &notification.UserRole); err != nil {
			continue
		}
		notifications = append(notifications, map[string]interface{}{
			"id":           notification.ID,
			"userId":       notification.UserID,
			"type":         notification.Type,
			"content":      notification.Content,
			"groupId":      notification.GroupID,
			"invitationId": notification.InvitationID,
			"read":         notification.Read,
			"createdAt":    notification.CreatedAt,
			"groupName":    notification.GroupName,
			"userRole":     notification.UserRole,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating notifications: %v", err)
		sendJSONError(w, "Error processing notifications", http.StatusInternalServerError)
		return
	}

	// If no notifications found, return empty array instead of null
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

	// Get current user from session to verify ownership
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
        SET read = true
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

package api

import (
	"encoding/json"
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
		log.Printf("Session error: %v", err)
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

	log.Printf("Fetching notifications for user ID: %d", userID)

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
                WHEN n.type IN ('group_invitation', 'join_request') THEN
                    CASE
                        WHEN gi.status IS NOT NULL AND gi.status != 'pending' THEN true
                        ELSE false
                    END
                ELSE false
            END as is_processed
        FROM notifications n
        LEFT JOIN group_members gm ON n.group_id = gm.group_id AND gm.user_id = n.user_id
        LEFT JOIN group_invitations gi ON n.invitation_id = gi.id
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

	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		log.Printf("Session error: %v", err)
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationID := r.URL.Query().Get("id")
	if notificationID == "" {
		log.Printf("No notification ID provided")
		sendJSONError(w, "No notification ID provided", http.StatusBadRequest)
		return
	}

	nID, err := strconv.Atoi(notificationID)
	if err != nil {
		log.Printf("Invalid notification ID: %v", err)
		sendJSONError(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		sendJSONError(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	var exists bool
	err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM notifications 
            WHERE id = ? AND user_id = ?
        )`, nID, userID).Scan(&exists)

	if err != nil {
		log.Printf("Error checking notification existence: %v", err)
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Printf("Notification %d not found for user %d", nID, userID)
		sendJSONError(w, "Notification not found", http.StatusNotFound)
		return
	}

	tx, err := sqlite.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
        UPDATE notifications 
        SET is_read = true
        WHERE id = ? AND user_id = ?`,
		nID, userID)

	if err != nil {
		log.Printf("Error updating notification: %v", err)
		sendJSONError(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Printf("Error updating notification or no rows affected: %v", err)
		sendJSONError(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	var unreadCount int
	err = tx.QueryRow(`
        SELECT COUNT(*) 
        FROM notifications 
        WHERE user_id = ? AND is_read = false`,
		userID).Scan(&unreadCount)

	if err != nil {
		log.Printf("Error getting unread count: %v", err)
		sendJSONError(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		sendJSONError(w, "Failed to save changes", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully marked notification %d as read", nID)
	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Notification marked as read",
		"unreadCount":    unreadCount,
		"notificationId": nID,
		"isRead":         true,
	})
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

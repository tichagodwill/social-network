package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social-network/pkg/db/sqlite"
	"social-network/server/models"
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

	// Get notification ID from URL
	notificationID := r.PathValue("id")
	if notificationID == "" {
		sendJSONError(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	nID, err := strconv.Atoi(notificationID)
	if err != nil {
		sendJSONError(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	// Get current user
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

	// Start transaction
	tx, err := sqlite.DB.Begin()
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify notification belongs to user
	var notificationExists bool
	err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM notifications 
            WHERE id = ? AND user_id = ?
        )`, nID, userID).Scan(&notificationExists)

	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !notificationExists {
		sendJSONError(w, "Notification not found", http.StatusNotFound)
		return
	}

	// Update notification
	_, err = tx.Exec(`
        UPDATE notifications 
        SET is_read = true 
        WHERE id = ? AND user_id = ?`,
		nID, userID)

	if err != nil {
		sendJSONError(w, "Failed to update notification", http.StatusInternalServerError)
		return
	}

	// Get updated unread count
	var unreadCount int
	err = tx.QueryRow(`
        SELECT COUNT(*) 
        FROM notifications 
        WHERE user_id = ? AND is_read = false`,
		userID).Scan(&unreadCount)

	if err != nil {
		sendJSONError(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		sendJSONError(w, "Failed to save changes", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "Notification marked as read",
		"unreadCount":    unreadCount,
		"notificationId": nID,
		"isRead":         true,
	})
}

func CreateChatNotification(recipientID, senderID int, content string) error {
	// Get sender info
	var senderName, senderAvatar string
	err := sqlite.DB.QueryRow(
		"SELECT first_name || ' ' || last_name, avatar FROM users WHERE id = ?", 
		senderID).Scan(&senderName, &senderAvatar)
	if err != nil {
		return fmt.Errorf("error getting sender info: %w", err)
	}

	// Create notification
	result, err := sqlite.DB.Exec(
		`INSERT INTO notifications (type, content, user_id, from_user_id, is_read, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"message",
		fmt.Sprintf("%s sent you a message: %s", senderName, truncateMessage(content)),
		recipientID,
		senderID,
		false,
		time.Now())
	if err != nil {
		return fmt.Errorf("error inserting notification: %w", err)
	}

	// Get the notification ID
	notificationID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting notification ID: %w", err)
	}

	// Send WebSocket notification
	notifyUsers := []int{recipientID}
	notification := map[string]interface{}{
		"id":           notificationID,
		"type":         "message",
		"content":      fmt.Sprintf("%s sent you a message: %s", senderName, truncateMessage(content)),
		"userId":       recipientID,
		"fromUserId":   senderID,
		"fromUserName": senderName,
		"fromUserAvatar": senderAvatar,
		"isRead":       false,
		"createdAt":    time.Now().Format(time.RFC3339),
	}

	// Broadcast the notification
	broadcast <- models.BroadcastMessage{
		Data:        models.WebSocketMessage{Type: "notification", Data: notification},
		TargetUsers: mapIntSliceToMap(notifyUsers),
	}

	return nil
}

func truncateMessage(message string) string {
	maxLen := 30
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen] + "..."
}

func mapIntSliceToMap(slice []int) map[int]bool {
	result := make(map[int]bool, len(slice))
	for _, v := range slice {
		result[v] = true
	}
	return result
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

package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

func GetMessages(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("userId")
	contactIdStr := r.URL.Query().Get("contactId")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	contactId, err := strconv.Atoi(contactIdStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	// Ensure the authenticated user is the one requesting the messages
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var authUserId int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&authUserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if authUserId != userId {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := sqlite.DB.Query(`
        SELECT 
            m.id,
            m.sender_id,
            m.recipient_id,
            m.content,
            m.created_at,
            u.username as sender_name,
            u.avatar as sender_avatar
        FROM chat_messages m
        JOIN users u ON m.sender_id = u.id 
        WHERE (m.sender_id = ? AND m.recipient_id = ?)
            OR (m.recipient_id = ? AND m.sender_id = ?)
        ORDER BY m.created_at ASC`,
		userId, contactId, userId, contactId)

	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.SenderName,
			&msg.SenderAvatar,
		); err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func handleDBError(w http.ResponseWriter, err error) {
	if err == sql.ErrNoRows {
		http.Error(w, "No messages found", http.StatusNotFound)
		return
	}
	http.Error(w, "Database error", http.StatusInternalServerError)
	log.Printf("Database error: %v", err)
}

func GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	groupIdStr := r.URL.Query().Get("groupId")

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Ensure the authenticated user is part of the group
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var authUserId int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&authUserId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var isMember bool
	err = sqlite.DB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM user_chat_status 
            WHERE chat_id = ? AND user_id = ?
        )`, groupId, authUserId).Scan(&isMember)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := sqlite.DB.Query(`
        SELECT 
            m.id,
            m.sender_id,
            m.content,
            m.created_at,
            u.username as sender_name,
            u.avatar as sender_avatar
        FROM chat_messages m
        JOIN users u ON m.sender_id = u.id
        WHERE m.chat_id = ?
        ORDER BY m.created_at ASC`,
		groupId)

	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.SenderName,
			&msg.SenderAvatar,
		); err != nil {
			log.Printf("Error scanning group message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

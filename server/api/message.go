package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"strconv"
)

func GetMessages(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	contactIdStr := r.PathValue("contactId")

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

	rows, err := sqlite.DB.Query(`
        SELECT 
            m.id,
            m.sender_id,
            m.recipient_id,
            m.content,
            m.status,
            m.message_type,
            m.file_data,
            m.file_name,
            m.file_type,
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

	var messages []m.ChatMessage
	for rows.Next() {
		var msg m.ChatMessage
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.Status,
			&msg.MessageType,
			&msg.FileData,
			&msg.FileName,
			&msg.FileType,
			&msg.CreatedAt,
			&msg.SenderName,
			&msg.SenderAvatar,
		)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func SaveMessage(message m.ChatMessage) error {
	_, err := sqlite.DB.Exec(`
        INSERT INTO chat_messages (
            sender_id,
            recipient_id, 
            content,
            status,
            message_type,
            file_data,
            file_name,
            file_type,
            created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		message.SenderID,
		message.RecipientID,
		message.Content,
		message.Status,
		message.MessageType,
		message.FileData,
		message.FileName,
		message.FileType,
		message.CreatedAt,
	)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		return err
	}
	return nil
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
	groupIdStr := r.PathValue("groupId")

	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	rows, err := sqlite.DB.Query(`
       SELECT 
           m.id,
           m.group_id,
           m.user_id,
           m.content,
           m.media,
           m.created_at,
           u.username,
           u.avatar
       FROM group_messages m
       JOIN users u ON m.user_id = u.id
       WHERE group_id = ?
       ORDER BY created_at ASC
   `, groupId)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No messages found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying group messages: %v", err)
		return
	}
	defer rows.Close()

	var messages []m.GroupMessage
	for rows.Next() {
		var msg m.GroupMessage
		err := rows.Scan(
			&msg.ID,
			&msg.GroupID,
			&msg.UserID,
			&msg.Content,
			&msg.Media,
			&msg.CreatedAt,
			&msg.UserName,
			&msg.UserAvatar,
		)
		if err != nil {
			log.Printf("Error scanning group message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	if err = json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

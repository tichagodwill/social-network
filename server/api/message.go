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
	userIdString := r.PathValue("userId")
	contactIdString := r.PathValue("contactId")

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	contactId, err := strconv.Atoi(contactIdString)
	if err != nil {
		http.Error(w, "Invalid contact id", http.StatusBadRequest)
		return
	}

	var messages []m.Chat_message

	rows, err := sqlite.DB.Query(`
SELECT sender_id, recipient_id, content, created_at
FROM chat_messages
where sender_id = ? and recipient_id = ?
   OR recipient_id = ? and sender_id = ?
`, userId, contactId, userId, contactId)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No messages found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error getting messages: %v", err)
		return
	}

	for rows.Next() {
		var m m.Chat_message
		err := rows.Scan(&m.SenderID, &m.RecipientID, &m.Content, &m.CreatedAt)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("Error scanning messages: %v", err)
			return
		}

		messages = append(messages, m)
	}

	encodeErr := json.NewEncoder(w).Encode(&messages)
	if encodeErr != nil {
		http.Error(w, "Error sending data", http.StatusInternalServerError)
	}
}

func SaveMessage(message m.Chat_message) {

	insertQuery := `
INSERT INTO chat_messages (sender_id, recipient_id, content, created_at)
VALUES (?, ?, ?, ?)
	`

	_, err := sqlite.DB.Exec(insertQuery, message.SenderID, message.RecipientID, message.Content, message.CreatedAt)
	if err != nil {
		log.Fatal("[SaveMessage] Error inserting message:", err)
	}
}

package models

import "time"

// models/messages.go
type ChatMessage struct {
	ID          int       `json:"id"`
	SenderID    int       `json:"senderId"`
	RecipientID int       `json:"recipientId"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	MessageType string    `json:"messageType"`
	FileData    []byte    `json:"fileData,omitempty"`
	FileName    string    `json:"fileName,omitempty"`
	FileType    string    `json:"fileType,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	// Frontend fields
	SenderName   string `json:"senderName,omitempty"`
	SenderAvatar string `json:"senderAvatar,omitempty"`
}

type GroupMessage struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"groupId"`
	UserID    int       `json:"userId"`
	Content   string    `json:"content"`
	Media     string    `json:"media,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	// Additional fields for frontend
	UserName   string `json:"userName,omitempty"`
	UserAvatar string `json:"userAvatar,omitempty"`
}

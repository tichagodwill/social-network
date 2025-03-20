package models

import (
	"time"
)

// ChatMessage represents a message in a chat
type ChatMessage struct {
	ID          int       `json:"id"`
	ChatID      int       `json:"chatId"` // Added ChatID to match the schema
	SenderID    int       `json:"senderId"`
	RecipientID int       `json:"recipientId"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`      // Added Status to match the schema
	MessageType string    `json:"messageType"` // Added MessageType to match the schema
	CreatedAt   time.Time `json:"createdAt"`
	// Additional fields for frontend
	SenderName   string `json:"senderName,omitempty"`
	SenderAvatar string `json:"senderAvatar,omitempty"`
}

// GroupMessage represents a message in a group chat
type GroupMessage struct {
	ID        int       `json:"id"`
	ChatId    int       `json:"chatId"`
	UserID    int       `json:"userId"`
	Content   string    `json:"content"`
	Media     string    `json:"media,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	// Additional fields for frontend
	UserName   string `json:"userName,omitempty"`
	UserAvatar string `json:"userAvatar,omitempty"`
}

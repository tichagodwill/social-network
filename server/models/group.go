package models

import "time"

type Group struct {
	ID          uint      `json:"id,omitempty"`
	CreatorID   uint      `json:"creator_id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type GroupMemebers struct {
	ID        uint      `json:"id,omitempty"`
	GroupID   uint      `json:"group_id,omitempty"`
	UserID    uint      `json:"user_id,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

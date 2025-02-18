package models

import "time"

type Group struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CreatorID       int       `json:"creator_id"`
	CreatorUsername string    `json:"creator_username"`
	CreatedAt       time.Time `json:"created_at"`
}

type GroupMemebers struct {
	ID        uint      `json:"id,omitempty"`
	GroupID   uint      `json:"group_id,omitempty"`
	UserID    uint      `json:"user_id,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type GroupEvent struct {
	ID          int       `json:"id"`
	GroupID     int       `json:"group_id"`
	CreatorID   int       `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	CreatedAt   time.Time `json:"created_at"`
}

type GroupEventRSVP struct {
	ID         int       `json:"id"`
	EventID    int       `json:"event_id"`
	UserID     int       `json:"user_id"`
	RSVPStatus string    `json:"rsvp_status"`
	CreatedAt  time.Time `json:"created_at"`
}

type GroupInvaitation struct {
	GroupID   uint `json:"group_id"`
	InviterID uint `json:"inviter_id"`
	ReciverID uint `json:"reciver_id"`
}

type GroupLeave struct {
	GroupID uint `json:"group_id"`
	UserID  uint `json:"user_id"`
}

type GroupPost struct {
	ID        int                `json:"id"`
	GroupID   int                `json:"group_id"`
	AuthorID  int                `json:"author_id"`
	Author    string             `json:"author"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Media     string             `json:"media,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Comments  []GroupPostComment `json:"comments,omitempty"`
}

type GroupPostComment struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	AuthorID  int    `json:"author_id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

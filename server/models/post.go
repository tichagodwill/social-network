package models

import "time"

type Post struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Media         string    `json:"media"`
	Privacy       int       `json:"privacy"`                 // 1: public, 2: almost-private (followers), 3: private
	SelectedUsers []int     `json:"selectedUsers,omitempty"` // Only used when Privacy = 3
	Author        int       `json:"author"`
	AuthorName    string    `json:"authorName"`
	AuthorAvatar  string    `json:"authorAvatar"`
	CreatedAt     time.Time `json:"created_at"`
	GroupID       int       `json:"group_id,omitempty"`
}

type PostPrivateView struct {
	ID     int `json:"id"`
	PostID int `json:"post_id"`
	UserID int `json:"user_id"`
}

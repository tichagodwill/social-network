package models

import "time"

type Comment struct {
	ID           uint      `json:"id,omitempty"`
	Content      string    `json:"content,omitempty"`
	Media        string    `json:"media,omitempty"`
	PostID       uint      `json:"post_id,omitempty"`
	Author       uint      `json:"author,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	AuthorName   string    `json:"author_name,omitempty"`
	AuthorAvatar string    `json:"avatar,omitempty"`
}

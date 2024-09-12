package models

import "time"

type Post struct {
	ID         uint      `json:"id,omitempty"`
	Title      string    `json:"title,omitempty"`
	Content    string    `json:"content,omitempty"`
	Media      string    `json:"media,omitempty"`
	Privay     uint      `json:"privay,omitempty"`
	Created_at time.Time `json:"created_at,omitempty"`
	Author     uint      `json:"author,omitempty"`
}

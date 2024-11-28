package models

import "time"

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Media     string    `json:"media"`
	Privacy   int       `json:"privacy"`
	Author    int       `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	GroupID   int       `json:"group_id,omitempty"`
}

type PostPrivateView struct {
    ID     int `json:"id"`      
    PostID int `json:"post_id"`  
    UserID int `json:"user_id"`  
}

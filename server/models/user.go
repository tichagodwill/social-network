package models

import "time"

type User struct {
	ID          uint      `json:"id,omitempty"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"password,omitempty"`
	FirstName   string    `json:"firstname,omitempty"`
	LastName    string    `json:"lastname,omitempty"`
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
	Username    string    `json:"username,omitempty"`
	AboutMe     string    `json:"about_me,omitempty"`
	IsPrivate   bool      `json:"is_private,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type UserResponse struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

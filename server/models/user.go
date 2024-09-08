package models

import "time"

type User struct {
	ID          uint      `json:"id,omitempty"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"password,omitempty"`
	FirstName   string    `json:"first_name,omitempty"`
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
	Username    string    `json:"username,omitempty"`
	AboutMe     string    `json:"about_me,omitempty"`
}

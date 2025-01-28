package models

type Followers struct {
	ID       uint   `json:"id,omitempty"`
	UserId   uint   `json:"followed_id,omitempty"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

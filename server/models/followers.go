package models

type Followers struct {
	ID        uint   `json:"id,omitempty"`
	UserId    uint   `json:"followed_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

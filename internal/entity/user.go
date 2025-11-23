package entity

import "time"

type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	TeamName  string    `json:"team_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) Activate() {
	u.IsActive = true
}

func (u *User) Deactivate() {
	u.IsActive = false
}

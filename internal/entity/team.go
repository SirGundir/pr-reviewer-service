package entity

import "time"

type Team struct {
	TeamName  string    `json:"team_name"`
	Members   []User    `json:"members"`
	CreatedAt time.Time `json:"created_at"`
}

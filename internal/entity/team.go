package entity

import "time"

type Team struct {
	TeamName  string
	Members   []User
	CreatedAt time.Time
}

package model

import "time"

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	LemonBalance   int       `json:"lemonBalance"`
	TotalHarvested int       `json:"totalHarvested"`
	LastHarvest    time.Time `json:"lastHarvest"`
	JoinedAt       time.Time `json:"joinedAt"`
	Instances      []string  `json:"instances"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

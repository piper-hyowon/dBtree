package user

import "time"

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	LemonBalance int        `json:"lemonBalance"`
	LastHarvest  *time.Time `json:"lastHarvest"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

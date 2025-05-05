package user

import "time"

type ProfileResponse struct {
	Email          string     `json:"email"`
	LemonBalance   int        `json:"lemonBalance"`
	TotalHarvested int        `json:"totalHarvested"`
	LastHarvest    *time.Time `json:"lastHarvest"`
	JoinedAt       time.Time  `json:"joinedAt"`
}
